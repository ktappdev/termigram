package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// CLI commands
func printHelp() {
	fmt.Println()
	fmt.Println(bold(cyan("Commands")))
	fmt.Println("  \\me                 Show current user info")
	fmt.Println("  \\contacts           Browse contacts by page and switch on selection")
	fmt.Println("  \\find <query>       Find cached chats/usernames and switch via selector")
	fmt.Println("  \\msg <id|@user> <text>  Send message and enter chat mode")
	fmt.Println("  \\image <source> [caption]  Send an image into the active chat")
	fmt.Println("  \\openimage [id|last]  Download/open an image from the active chat")
	fmt.Println("  \\to <id|@user>      Switch active chat")
	fmt.Println("  \\here               Show active chat")
	fmt.Println("  \\chats              Interactive recent chats picker (↑/↓, Enter, Esc, filter)")
	fmt.Println("  \\unread             Pick from chats with unread messages and enter the selection")
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
		unread := ""
		if c.UnreadCount > 0 {
			unread = " " + blue(fmt.Sprintf("[%d unread]", c.UnreadCount))
		}
		fmt.Printf(" %s %s %s %s%s%s\n", marker, bold(c.Label), dim("("+c.Target+")"), timeLabel, preview, unread)
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
	chat, err := cli.resolveChatTarget(ctx, target)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error:"), err)
		return
	}
	cli.activateCachedChat(chat, silent)
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

	result := cli.pickCachedChat("Recent chats", "", chats)
	if !result.Interactive {
		cli.showCachedChats()
		return
	}
	if result.Cancelled {
		fmt.Println(dim("Chat switch cancelled."))
		return
	}
	if result.Chosen != nil {
		cli.activateCachedChat(*result.Chosen, false)
	}
}

func (cli *TelegramCLI) runUnreadPicker(ctx context.Context) {
	chats, err := cli.fetchDialogs(ctx, 50, true)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error loading unread chats:"), err)
		return
	}
	if len(chats) == 0 {
		fmt.Println(dim("No unread chats."))
		return
	}

	chosen, ok := cli.selectCachedChat(
		"Unread chats",
		"",
		chats,
		"No unread chats.",
		"Unread chat selection cancelled.",
		"Unread chats",
	)
	if !ok {
		return
	}
	if loadErr := cli.ensureLegacyTranscriptContext(ctx, chosen.Target, chosen.Label, unreadTranscriptMinContextEntries); loadErr != nil {
		fmt.Printf("%s %v\n", yellow("Warning:"), loadErr)
	}
	cli.activateCachedChat(*chosen, false)
}

func (cli *TelegramCLI) commandLoop(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	for {
		if ctx.Err() != nil {
			fmt.Println("\nExiting...")
			return
		}

		input, err := cli.readInteractiveLine(ctx, sigChan)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				fmt.Println("\nExiting...")
				return
			case errors.Is(err, errLegacyPromptInterrupted):
				fmt.Println("\nInterrupt received, exiting...")
				if cli.cancel != nil {
					cli.cancel()
				}
				return
			default:
				if ctx.Err() != nil {
					fmt.Println("\nExiting...")
					return
				}
				return
			}
		}

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

		parts, splitErr := splitCommandTokens(input)
		if splitErr != nil {
			fmt.Printf("%s %v\n", yellow("Error parsing command:"), splitErr)
			continue
		}
		if len(parts) == 0 {
			continue
		}
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "\\me":
			cli.showSelf(ctx)
		case "\\contacts":
			cli.showContacts(ctx)
		case "\\find":
			if len(parts) < 2 {
				fmt.Println(yellow("Usage:"), "\\find <query>")
				continue
			}

			query := strings.Join(parts[1:], " ")
			chosen, ok := cli.selectCachedChat(
				"Find chat",
				query,
				cli.cachedChatsForPartial(query, 10),
				fmt.Sprintf("No cached chats match %q.", query),
				"Find cancelled.",
				"Cached chat matches",
			)
			if !ok {
				continue
			}
			cli.activateCachedChat(*chosen, false)
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
		case "\\image":
			if len(parts) < 2 {
				fmt.Println(yellow("Usage:"), "\\image <source> [caption]")
				continue
			}
			target, label := cli.currentChat()
			if target == "" {
				fmt.Println(yellow("No active chat."), "Use \\to or \\msg first, then run \\image <source> [caption].")
				continue
			}
			caption := ""
			if len(parts) > 2 {
				caption = strings.Join(parts[2:], " ")
			}
			if err := cli.sendImage(ctx, target, label, parts[1], caption); err != nil {
				fmt.Printf("%s %v\n", red("Error sending image:"), err)
			}
		case "\\openimage":
			selector := "last"
			if len(parts) > 2 {
				fmt.Println(yellow("Usage:"), "\\openimage [message-id|last]")
				continue
			}
			if len(parts) == 2 {
				selector = parts[1]
			}
			path, err := cli.openImageFromCurrentChat(ctx, selector)
			if err != nil {
				fmt.Printf("%s %v\n", red("Error opening image:"), err)
				continue
			}
			fmt.Printf("%s %s\n", dim("Opened image:"), path)
		case "\\to":
			if len(parts) < 2 {
				fmt.Println(yellow("Usage:"), "\\to <id|@user>")
				continue
			}

			target := parts[1]
			if chat, err := cli.resolveChatTarget(ctx, target); err == nil {
				cli.activateCachedChat(chat, false)
				continue
			}

			matches := cli.cachedChatsForPartial(target, 10)
			if len(matches) == 1 {
				cli.activateCachedChat(matches[0], false)
				continue
			}

			chosen, ok := cli.selectCachedChat(
				"Switch chat",
				target,
				matches,
				fmt.Sprintf("No cached chats match %q. Try \\contacts, \\find <query>, or an exact @username/user id.", target),
				"Chat switch cancelled.",
				"Cached chat matches",
			)
			if !ok {
				continue
			}
			cli.activateCachedChat(*chosen, false)
		case "\\here":
			cli.showActiveChat()
		case "\\chats":
			cli.runChatsPicker()
		case "\\unread":
			cli.runUnreadPicker(ctx)
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
