package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gotd/td/tg"
	"golang.org/x/term"
)

type chatPickerResult struct {
	Chosen      *CachedChat
	Cancelled   bool
	Interactive bool
}

func cachedChatFromUser(user *tg.User) CachedChat {
	label := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if label == "" {
		label = user.Username
	}
	if label == "" {
		label = fmt.Sprintf("User %d", user.ID)
	}

	target := fmt.Sprintf("%d", user.ID)
	if user.Username != "" {
		target = "@" + user.Username
	}

	return CachedChat{Label: label, Target: target}
}

func (cli *TelegramCLI) resolveChatTarget(ctx context.Context, target string) (CachedChat, error) {
	user, err := cli.findUserByIDOrUsername(ctx, target)
	if err != nil {
		return CachedChat{}, err
	}
	return cachedChatFromUser(user), nil
}

func (cli *TelegramCLI) activateCachedChat(chat CachedChat, silent bool) {
	cli.setCurrentChat(chat.Target, chat.Label)
	if !silent {
		fmt.Printf("%s %s %s\n", green("✓ Active chat:"), bold(chat.Label), dim("("+chat.Target+")"))
	}
}

func (cli *TelegramCLI) cachedChatsForUsernamePrefix(prefix string, limit int) []CachedChat {
	prefix = normalizeUsername(prefix)
	if prefix == "" {
		return nil
	}

	matches := make([]CachedChat, 0)
	for _, chat := range cli.listCachedChats(0) {
		if strings.HasPrefix(chat.Target, "@") && strings.HasPrefix(normalizeUsername(chat.Target), prefix) {
			matches = append(matches, chat)
		}
	}
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func (cli *TelegramCLI) cachedChatsForPartial(query string, limit int) []CachedChat {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	rawQuery := strings.ToLower(query)
	normalizedQuery := normalizeUsername(query)
	type scoredChat struct {
		chat  CachedChat
		score int
	}

	matches := make([]scoredChat, 0)
	for _, chat := range cli.listCachedChats(0) {
		username := normalizeUsername(chat.Target)
		label := strings.ToLower(chat.Label)
		score := -1

		switch {
		case normalizedQuery != "" && username != "" && strings.HasPrefix(username, normalizedQuery):
			score = 0
		case strings.HasPrefix(label, rawQuery):
			score = 1
		case normalizedQuery != "" && username != "" && strings.Contains(username, normalizedQuery):
			score = 2
		case strings.Contains(label, rawQuery):
			score = 3
		case fuzzyMatch(query, chat.Label+" "+chat.Target):
			score = 4
		}

		if score >= 0 {
			matches = append(matches, scoredChat{chat: chat, score: score})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].score < matches[j].score
	})

	results := make([]CachedChat, 0, len(matches))
	for _, match := range matches {
		results = append(results, match.chat)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func (cli *TelegramCLI) printChatCandidates(title string, chats []CachedChat) {
	fmt.Println(bold(cyan(title)))
	for _, c := range chats {
		fmt.Printf("  %s %s\n", bold(c.Label), dim("("+c.Target+")"))
	}
}

func (cli *TelegramCLI) pickCachedChat(title string, initialQuery string, chats []CachedChat) chatPickerResult {
	if len(chats) == 0 {
		return chatPickerResult{Interactive: true}
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return chatPickerResult{Interactive: false}
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("%s %v\n", red("Could not start interactive selector:"), err)
		return chatPickerResult{Interactive: false}
	}
	defer term.Restore(fd, oldState)

	out := func(format string, args ...any) {
		fmt.Printf(strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}

	query := initialQuery
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
		out("%s %s\n", bold(cyan(title)), dim("(Esc to cancel, Enter to select)"))
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
		case 3:
			fmt.Print("\033[2J\033[H")
			out("\n")
			return chatPickerResult{Cancelled: true, Interactive: true}
		case 13:
			if len(filtered) == 0 {
				continue
			}
			fmt.Print("\033[2J\033[H")
			out("\n")
			chosen := filtered[selected]
			return chatPickerResult{Chosen: &chosen, Interactive: true}
		case 27:
			next, ok := readByteWithTimeout(fd, 15)
			if !ok || next != '[' {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return chatPickerResult{Cancelled: true, Interactive: true}
			}
			arrow, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return chatPickerResult{Cancelled: true, Interactive: true}
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
				return chatPickerResult{Cancelled: true, Interactive: true}
			}
		case 127, 8:
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
	return chatPickerResult{Cancelled: true, Interactive: true}
}
