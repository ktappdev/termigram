package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const appVersion = "v0.1.0"

func main() {
	remainingArgs, rootMode, handled, err := parseRootFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printRootHelp()
		os.Exit(2)
	}
	if handled {
		return
	}

	if len(remainingArgs) > 0 && isCLICommand(remainingArgs[0]) && hasHelpFlag(remainingArgs[1:]) {
		printCommandHelp(remainingArgs[0])
		return
	}

	isCLIMode := len(remainingArgs) > 0 && !strings.HasPrefix(remainingArgs[0], "-")

	cfg, err := loadConfig()
	if err != nil {
		printConfigErrorAndExit(err)
	}

	if isCLIMode {
		runCLIMode(cfg, remainingArgs, rootMode)
		return
	}

	selectedMode, err := cfg.ResolveMode(rootMode, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if selectedMode != "user" {
		fmt.Fprintln(os.Stderr, "Error: interactive mode only supports --mode user")
		os.Exit(1)
	}

	cli := NewTelegramCLI(cfg.TelegramAppID, cfg.TelegramAppHash, cfg.SessionPath)
	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printConfigErrorAndExit(err error) {
	fmt.Println("Error:", err)
	fmt.Println("\nUser mode setup steps:")
	fmt.Println("1. Go to https://my.telegram.org")
	fmt.Println("2. Log in with your phone number")
	fmt.Println("3. Create a new application")
	fmt.Println("4. Copy the app_id and app_hash")
	fmt.Printf("\nThen place %s next to the modern-telegram-cli executable (for local builds, that's typically this directory) by copying %s.example, or set environment variables:\n", localConfigFile, localConfigFile)
	fmt.Println("  cp /path/to/modern-telegram-cli/config.json.example /path/to/modern-telegram-cli/config.json")
	fmt.Println("  export TELEGRAM_APP_ID=your_app_id")
	fmt.Println("  export TELEGRAM_APP_HASH=your_app_hash")
	fmt.Println("  export TELEGRAM_BOT_TOKEN=your_bot_token")
	os.Exit(1)
}

func parseRootFlags(args []string) (remaining []string, mode string, handled bool, err error) {
	fs := flag.NewFlagSet("modern-telegram-cli", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var help bool
	var version bool
	var modeFlag string
	fs.BoolVar(&help, "help", false, "Show help")
	fs.BoolVar(&help, "h", false, "Show help")
	fs.BoolVar(&version, "version", false, "Show version")
	fs.BoolVar(&version, "v", false, "Show version")
	fs.StringVar(&modeFlag, "mode", "", "Auth mode: user|bot")

	if err := fs.Parse(args); err != nil {
		return nil, "", false, err
	}

	if help {
		printRootHelp()
		return nil, "", true, nil
	}
	if version {
		fmt.Printf("modern-telegram-cli %s\n", appVersion)
		return nil, "", true, nil
	}

	return fs.Args(), modeFlag, false, nil
}

func createBackend(cfg Config) *UserBackend {
	return NewUserBackend(cfg)
}

func runCLIMode(cfg Config, argv []string, rootMode string) {
	if len(argv) < 1 {
		printRootHelp()
		os.Exit(1)
	}

	command := argv[0]
	if !isCLICommand(command) {
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n\n", command)
		printRootHelp()
		os.Exit(1)
	}

	args := argv[1:]
	if hasHelpFlag(args) {
		printCommandHelp(command)
		return
	}

	fs := flag.NewFlagSet("cli", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output as JSON")
	limitFlag := fs.Int("limit", 10, "Limit for get command")
	timeoutFlag := fs.Duration("timeout", 30*time.Second, "Request timeout")
	modeFlag := fs.String("mode", "", "Auth mode: user|bot")

	fs.Parse(args)
	positionalArgs := fs.Args()

	effectiveMode := strings.TrimSpace(rootMode)
	if strings.TrimSpace(*modeFlag) != "" {
		if effectiveMode != "" && effectiveMode != strings.TrimSpace(*modeFlag) {
			fmt.Fprintln(os.Stderr, "Error: conflicting --mode values between global and command flags")
			os.Exit(1)
		}
		effectiveMode = strings.TrimSpace(*modeFlag)
	}

	positionalArgs, limit := parseLimitArg(positionalArgs)
	if *limitFlag != 10 {
		limit = *limitFlag
	}

	selectedMode, err := cfg.ResolveMode(effectiveMode, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if selectedMode == "bot" {
		fmt.Fprintln(os.Stderr, "Error: --mode bot is configured but bot backend is not available in this build")
		os.Exit(1)
	}

	cmd := CLICommand{
		Name:    command,
		Args:    positionalArgs,
		JSON:    *jsonFlag,
		Limit:   limit,
		Timeout: *timeoutFlag,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	defer cancel()

	backend := createBackend(cfg)
	err = backend.Run(ctx, func(runCtx context.Context) error {
		return RunCLICommand(runCtx, backend, cmd)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isCLICommand(name string) bool {
	switch name {
	case "send", "get", "contacts", "me", "find":
		return true
	default:
		return false
	}
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

func printRootHelp() {
	fmt.Println(`modern-telegram-cli - Telegram MTProto CLI

Usage:
  ./modern-telegram-cli [--help|-h] [--version|-v] [--mode user|bot]
  ./modern-telegram-cli [--mode user]                # interactive mode
  ./modern-telegram-cli <command> [options]          # one-shot mode

One-shot commands:
  send [--mode user|bot] [--json] [--timeout 30s] <user_id|@username> <message>
  get [--mode user|bot] [--json] [--timeout 30s] [--limit N] <user_id|@username>
  contacts [--mode user|bot] [--json] [--timeout 30s]
  me [--mode user|bot] [--json] [--timeout 30s]
  find [--mode user|bot] [--json] [--timeout 30s] <prefix>

Examples:
  ./modern-telegram-cli
  ./modern-telegram-cli send @ken "Hello"
  ./modern-telegram-cli get --json --limit 20 @ken
  ./modern-telegram-cli contacts --json
  ./modern-telegram-cli send --help

Global flags:
  -h, --help       Show this help
  -v, --version    Show app version
  --mode           Auth mode: user|bot

Command help:
  ./modern-telegram-cli <command> --help`)
}

func printCommandHelp(command string) {
	switch command {
	case "send":
		fmt.Println(`send - send a message to a user

Usage:
  ./modern-telegram-cli send [--mode user|bot] [--json] [--timeout 30s] <user_id|@username> <message>

Examples:
  ./modern-telegram-cli send @ken "Hello from script"
  ./modern-telegram-cli send --json 123456789 "Hello"`)
	case "get":
		fmt.Println(`get - fetch recent messages from a user

Usage:
  ./modern-telegram-cli get [--mode user|bot] [--json] [--timeout 30s] [--limit N] <user_id|@username>

Flags:
  --limit N   Number of messages to fetch (default 10)

Examples:
  ./modern-telegram-cli get @ken
  ./modern-telegram-cli get --json --limit 20 @ken`)
	case "contacts":
		fmt.Println(`contacts - list contacts

Usage:
  ./modern-telegram-cli contacts [--mode user|bot] [--json] [--timeout 30s]

Examples:
  ./modern-telegram-cli contacts
  ./modern-telegram-cli contacts --json`)
	case "me":
		fmt.Println(`me - show current account info

Usage:
  ./modern-telegram-cli me [--mode user|bot] [--json] [--timeout 30s]

Examples:
  ./modern-telegram-cli me
  ./modern-telegram-cli me --json`)
	case "find":
		fmt.Println(`find - find cached usernames by prefix

Usage:
  ./modern-telegram-cli find [--mode user|bot] [--json] [--timeout 30s] <prefix>

Examples:
  ./modern-telegram-cli find ken
  ./modern-telegram-cli find --json ken`)
	default:
		printRootHelp()
	}
}
