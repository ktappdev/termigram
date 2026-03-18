package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// CLI commands
func printHelp() {
	fmt.Println()
	fmt.Println(bold(cyan("Commands")))
	fmt.Println("  \\me                 Show current user info")
	fmt.Println("  \\contacts           List contacts")
	fmt.Println("  \\find <prefix>      Find usernames by prefix (cache)")
	fmt.Println("  \\msg <id|@user> <text>  Send message and enter chat mode")
	fmt.Println("  \\to <id|@user>      Switch active chat")
	fmt.Println("  \\here               Show active chat")
	fmt.Println("  \\chats              Interactive recent chats picker (↑/↓, Enter, Esc, filter)")
	fmt.Println("  \\close              Exit chat mode")
	fmt.Println("  \\chat / \\back      Deprecated aliases for \\here/\\to and \\close")
	fmt.Println("  \\help               Show this help")
	fmt.Println("  \\quit               Exit")
	fmt.Println()
}

type inputReadResult struct {
	line string
	err  error
}

func (cli *TelegramCLI) showActiveChat() {
	target, label := cli.currentChat()
	if target == "" {
		fmt.Println(dim("No active chat."))
		return
	}
	fmt.Printf("%s %s %s\n", dim("Active chat:"), bold(label), dim("("+target+")"))
}

func (cli *TelegramCLI) showCachedChats() {
	chats := cli.listCachedChats(0)
	if len(chats) == 0 {
		fmt.Println(dim("No cached conversations yet."))
		return
	}

	activeTarget, _ := cli.currentChat()
	fmt.Println(bold(cyan("Cached conversations")))
	for _, c := range chats {
		marker := " "
		if normalizeUsername(activeTarget) == normalizeUsername(c.Target) {
			marker = yellow("●")
		}
		timeLabel := ""
		if !c.LastActivity.IsZero() {
			timeLabel = dim(c.LastActivity.Format("15:04"))
		}
		preview := ""
		if c.LastMessage != "" {
			preview = dim(" — " + truncateInline(c.LastMessage, 32))
		}
		fmt.Printf(" %s %s %s %s%s\n", marker, bold(c.Label), dim("("+c.Target+")"), timeLabel, preview)
	}
}

func (cli *TelegramCLI) clearActiveChat(silent bool) {
	target, label := cli.currentChat()
	if target == "" {
		if !silent {
			fmt.Println(dim("No active chat to close."))
		}
		return
	}
	cli.setCurrentChat("", "")
	fmt.Printf("%s %s\n", dim("Closed chat mode:"), bold(label))
}

func (cli *TelegramCLI) switchActiveChat(ctx context.Context, target string, silent bool) {
	user, err := cli.findUserByIDOrUsername(ctx, target)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error:"), err)
		return
	}
	label := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if label == "" {
		label = user.Username
	}
	if label == "" {
		label = fmt.Sprintf("User %d", user.ID)
	}
	chatTarget := fmt.Sprintf("%d", user.ID)
	if user.Username != "" {
		chatTarget = "@" + user.Username
	}
	cli.setCurrentChat(chatTarget, label)
	if !silent {
		fmt.Printf("%s %s %s\n", green("✓ Active chat:"), bold(label), dim("("+chatTarget+")"))
	}
}

func truncateInline(s string, max int) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ")
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return s[:max-1] + "…"
}

func fuzzyMatch(query string, text string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	t := strings.ToLower(text)
	qi := 0
	for i := 0; i < len(t) && qi < len(q); i++ {
		if t[i] == q[qi] {
			qi++
		}
	}
	return qi == len(q)
}

func filterChats(chats []CachedChat, query string) []CachedChat {
	if strings.TrimSpace(query) == "" {
		return chats
	}
	filtered := make([]CachedChat, 0, len(chats))
	for _, c := range chats {
		haystack := c.Label + " " + c.Target + " " + c.LastMessage
		if fuzzyMatch(query, haystack) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func readByteWithTimeout(fd int, timeoutMs int) (byte, bool) {
	pollFds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	n, err := unix.Poll(pollFds, timeoutMs)
	if err != nil || n == 0 || pollFds[0].Revents&unix.POLLIN == 0 {
		return 0, false
	}
	buf := make([]byte, 1)
	if _, err := os.Stdin.Read(buf); err != nil {
		return 0, false
	}
	return buf[0], true
}

func (cli *TelegramCLI) runChatsPicker() {
	chats := cli.listCachedChats(0)
	if len(chats) == 0 {
		fmt.Println(dim("No cached conversations yet."))
		return
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		cli.showCachedChats()
		return
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("%s %v\n", red("Could not start interactive chats mode:"), err)
		cli.showCachedChats()
		return
	}
	defer term.Restore(fd, oldState)

	out := func(format string, args ...any) {
		fmt.Printf(strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}

	query := ""
	selected := 0

	render := func() []CachedChat {
		filtered := filterChats(chats, query)
		if selected >= len(filtered) {
			selected = len(filtered) - 1
		}
		if selected < 0 {
			selected = 0
		}

		fmt.Print("\033[2J\033[H")
		out("%s %s\n", bold(cyan("Recent chats")), dim("(Esc to cancel, Enter to select)"))
		out("%s %s\n\n", dim("Filter:"), query)
		if len(filtered) == 0 {
			out("%s\n", dim("  No matches"))
			return filtered
		}

		for i, c := range filtered {
			cursor := "  "
			lineStyle := func(s string) string { return s }
			if i == selected {
				cursor = yellow("➤ ")
				lineStyle = bold
			}
			timeLabel := ""
			if !c.LastActivity.IsZero() {
				timeLabel = " " + dim(c.LastActivity.Format("15:04"))
			}
			out("%s%s %s%s\n", cursor, lineStyle(c.Label), dim("("+c.Target+")"), timeLabel)
		}
		return filtered
	}

	filtered := render()
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}

		b := buf[0]
		switch b {
		case 3: // Ctrl+C
			fmt.Print("\033[2J\033[H")
			out("\n")
			return
		case 13: // Enter
			fmt.Print("\033[2J\033[H")
			if len(filtered) > 0 {
				chosen := filtered[selected]
				cli.setCurrentChat(chosen.Target, chosen.Label)
				out("\n%s %s %s\n", green("✓ Active chat:"), bold(chosen.Label), dim("("+chosen.Target+")"))
			}
			out("\n")
			return
		case 27: // ESC or arrows
			next, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return
			}
			if next != '[' {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return
			}
			arrow, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return
			}
			switch arrow {
			case 'A':
				if selected > 0 {
					selected--
				}
			case 'B':
				if selected < len(filtered)-1 {
					selected++
				}
			default:
				fmt.Print("\033[2J\033[H")
				out("\n")
				return
			}
		case 127, 8: // backspace
			if len(query) > 0 {
				query = query[:len(query)-1]
			}
		default:
			if b >= 32 && b <= 126 {
				query += string(b)
			}
		}
		filtered = render()
	}
	fmt.Print("\033[2J\033[H")
	out("\n")
}

func (cli *TelegramCLI) commandLoop(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	for {
		fmt.Print(cli.promptLabel())

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
				target, _ := cli.currentChat()
				if target != "" {
					cli.sendMessage(ctx, target, input)
					continue
				}
				fmt.Println(yellow("No active chat."), "Use \\msg <id|@user> <text> to start, or \\help.")
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
					fmt.Println(yellow("Usage:"), "\\find <prefix>")
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
					fmt.Println(yellow("Usage:"), "\\msg <user_id|@username> <message>")
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
			case "\\to":
				if len(parts) < 2 {
					fmt.Println(yellow("Usage:"), "\\to <id|@user>")
					continue
				}
				cli.switchActiveChat(ctx, parts[1], false)
			case "\\here":
				cli.showActiveChat()
			case "\\chats":
				cli.runChatsPicker()
			case "\\close":
				cli.clearActiveChat(false)
			case "\\chat":
				fmt.Println(dim("\\chat is deprecated. Use \\here or \\to <id|@user>."))
				if len(parts) == 1 {
					cli.showActiveChat()
				} else {
					cli.switchActiveChat(ctx, parts[1], true)
				}
			case "\\back":
				fmt.Println(dim("\\back is deprecated. Use \\close."))
				cli.clearActiveChat(true)
			case "\\help":
				printHelp()
			case "\\quit", "\\exit":
				fmt.Println("Goodbye!")
				if cli.cancel != nil {
					cli.cancel()
				}
				return
			default:
				fmt.Printf("%s %s. Type \\help for available commands.\n", yellow("Unknown command:"), cmd)
			}
		}
	}
}
