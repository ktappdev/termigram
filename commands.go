package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// CLI commands
func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  \\me        - Show current user info")
	fmt.Println("  \\contacts  - List contacts")
	fmt.Println("  \\find <prefix> - Find usernames starting with prefix (from cache)")
	fmt.Println("  \\msg <id|@username> <text> - Send message to user")
	fmt.Println("  \\help      - Show this help")
	fmt.Println("  \\quit      - Exit")
	fmt.Println()
}

type inputReadResult struct {
	line string
	err  error
}

func (cli *TelegramCLI) commandLoop(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	for {
		fmt.Print("> ")

		inputChan := make(chan inputReadResult, 1)
		go func() {
			line, err := cli.reader.ReadString('\n')
			inputChan <- inputReadResult{line: line, err: err}
		}()

		select {
		case <-ctx.Done():
			fmt.Println("\nExiting...")
			return
		case <-sigChan:
			fmt.Println("\nInterrupt received, exiting...")
			if cli.cancel != nil {
				cli.cancel()
			}
			return
		case result := <-inputChan:
			if result.err != nil {
				if ctx.Err() != nil {
					fmt.Println("\nExiting...")
					return
				}
				return
			}

			input := strings.TrimSpace(result.line)
			if input == "" {
				continue
			}

			if !strings.HasPrefix(input, "\\") {
				fmt.Println("Unknown command. Type \\help for available commands.")
				continue
			}

			parts := strings.Fields(input)
			cmd := strings.ToLower(parts[0])

			switch cmd {
			case "\\me":
				cli.showSelf(ctx)
			case "\\contacts":
				cli.showContacts(ctx)
			case "\\find":
				if len(parts) < 2 {
					fmt.Println("Usage: \\find <prefix>")
					continue
				}
				matches := cli.findMatchingUsernames(parts[1], 10)
				if len(matches) == 0 {
					fmt.Printf("No cached usernames found starting with %q.\n", parts[1])
				} else {
					fmt.Println("Cached matches:")
					for _, m := range matches {
						fmt.Println(" ", m)
					}
				}
			case "\\msg":
				if len(parts) < 3 {
					fmt.Println("Usage: \\msg <user_id|@username> <message>")
					continue
				}
				target := parts[1]
				msgText := strings.Join(parts[2:], " ")

				if strings.HasPrefix(target, "@") {
					username := normalizeUsername(target)
					if _, found := cli.getUserByUsername(username); found {
						cli.sendMessage(ctx, target, msgText)
					} else {
						suggestions := cli.findMatchingUsernames(username, 5)
						if len(suggestions) > 0 {
							fmt.Printf("Message not sent. Did you mean: %s?\n", strings.Join(suggestions, ", "))
						} else {
							cli.sendMessage(ctx, target, msgText)
						}
					}
				} else {
					cli.sendMessage(ctx, target, msgText)
				}
			case "\\help":
				printHelp()
			case "\\quit", "\\exit":
				fmt.Println("Goodbye!")
				if cli.cancel != nil {
					cli.cancel()
				}
				return
			default:
				fmt.Printf("Unknown command: %s. Type \\help for available commands.\n", cmd)
			}
		}
	}
}
