package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Build-time variables overridden via ldflags, e.g.:
// go build -ldflags "-X main.appVersion=v1.2.3 -X main.telegramAppIDBaked=123456 -X main.telegramAppHashBaked=abcdef"
var (
	appVersion           = "dev"
	telegramAppIDBaked   = ""
	telegramAppHashBaked = ""
)

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
	_ = selectedMode // Always "user" after ResolveMode succeeds

	cli := NewTelegramCLI(cfg.TelegramAppID, cfg.TelegramAppHash, cfg.SessionPath)
	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printConfigErrorAndExit(err error) {
	fmt.Println("Error:", err)
	fmt.Println("\nCredential lookup order: TELEGRAM_APP_ID/TELEGRAM_APP_HASH env vars → config.json → baked-in build credentials")
	fmt.Println("\nIf this build does not include baked-in credentials, set up user mode with one of these options:")
	fmt.Printf("1. Copy %s.example next to the termigram executable as %s and add your Telegram app_id/app_hash\n", localConfigFile, localConfigFile)
	fmt.Println("   cp /path/to/termigram/config.json.example /path/to/termigram/config.json")
	fmt.Println("2. Or export environment variables:")
	fmt.Println("   export TELEGRAM_APP_ID=your_app_id")
	fmt.Println("   export TELEGRAM_APP_HASH=your_app_hash")
	fmt.Println("3. Or build termigram with baked-in credentials (see CREDENTIALS.md)")
	fmt.Println("\nNeed your own credentials? Create them at https://my.telegram.org")
	os.Exit(1)
}

func parseRootFlags(args []string) (remaining []string, mode string, handled bool, err error) {
	fs := flag.NewFlagSet("termigram", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var help bool
	var version bool
	var modeFlag string
	fs.BoolVar(&help, "help", false, "Show help")
	fs.BoolVar(&help, "h", false, "Show help")
	fs.BoolVar(&version, "version", false, "Show version")
	fs.BoolVar(&version, "v", false, "Show version")
	fs.StringVar(&modeFlag, "mode", "", "Auth mode: user (default: user)")

	if err := fs.Parse(args); err != nil {
		return nil, "", false, err
	}

	if help {
		printRootHelp()
		return nil, "", true, nil
	}
	if version {
		fmt.Printf("termigram %s\n", appVersion)
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
	limitFlag := fs.Int("limit", 10, "Limit for get/contacts commands")
	offsetFlag := fs.Int("offset", 0, "Offset for contacts command")
	timeoutFlag := fs.Duration("timeout", 30*time.Second, "Request timeout")
	modeFlag := fs.String("mode", "", "Auth mode: user (default: user)")

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
	_ = selectedMode // Only user mode is supported

	cmd := CLICommand{
		Name:    command,
		Args:    positionalArgs,
		JSON:    *jsonFlag,
		Limit:   limit,
		Offset:  *offsetFlag,
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
	fmt.Println(`termigram - Telegram MTProto CLI

Usage:
  ./termigram [--help|-h] [--version|-v]
  ./termigram                                          # interactive mode
  ./termigram <command> [options]                      # one-shot mode

One-shot commands:
  send [--json] [--timeout 30s] <user_id|@username> <message>
  get [--json] [--timeout 30s] [--limit N] <user_id|@username>
  contacts [--json] [--timeout 30s] [--limit N] [--offset N]
  me [--json] [--timeout 30s]
  find [--json] [--timeout 30s] <prefix>

Examples:
  ./termigram
  ./termigram send @ken "Hello"
  ./termigram get --json --limit 20 @ken
  ./termigram contacts --json
  ./termigram send --help

Global flags:
  -h, --help       Show this help
  -v, --version    Show app version

Command help:
  ./termigram <command> --help`)
}

func printCommandHelp(command string) {
	switch command {
	case "send":
		fmt.Println(`send - send a message to a user

Usage:
  ./termigram send [--json] [--timeout 30s] <user_id|@username> <message>

Examples:
  ./termigram send @ken "Hello from script"
  ./termigram send --json 123456789 "Hello"`)
	case "get":
		fmt.Println(`get - fetch recent messages from a user

Usage:
  ./termigram get [--json] [--timeout 30s] [--limit N] <user_id|@username>

Flags:
  --limit N   Number of messages to fetch (default 10)

Examples:
  ./termigram get @ken
  ./termigram get --json --limit 20 @ken`)
	case "contacts":
		fmt.Println(`contacts - list contacts with pagination

Usage:
  ./termigram contacts [--json] [--timeout 30s] [--limit N] [--offset N]

Flags:
  --limit N    Number of contacts per page (default 10)
  --offset N   Starting offset for pagination (default 0)

Examples:
  ./termigram contacts
  ./termigram contacts --limit 20 --offset 40
  ./termigram contacts --json`)
	case "me":
		fmt.Println(`me - show current account info

Usage:
  ./termigram me [--json] [--timeout 30s]

Examples:
  ./termigram me
  ./termigram me --json`)
	case "find":
		fmt.Println(`find - find cached usernames by prefix

Usage:
  ./termigram find [--json] [--timeout 30s] <prefix>

Examples:
  ./termigram find ken
  ./termigram find --json ken`)
	default:
		printRootHelp()
	}
}
