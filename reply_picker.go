package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	replyPickerLimit          = 25
	replyPickerContextEntries = 20
)

var errReplyPickerCancelled = errors.New("reply selection cancelled")

type replyPickerItem struct {
	Entry     transcriptEntry
	Title     string
	Subtitle  string
	QueryText string
}

type replyPickerResult struct {
	Chosen      *transcriptEntry
	Cancelled   bool
	Interactive bool
}

func replyPickerItems(entries []transcriptEntry, limit int) []replyPickerItem {
	if limit <= 0 {
		limit = replyPickerLimit
	}

	items := make([]replyPickerItem, 0, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.MessageID <= 0 {
			continue
		}
		preview := entryPreviewText(entry)
		if preview == "" {
			preview = "[message]"
		}
		sender := strings.TrimSpace(entry.Sender)
		if sender == "" && entry.Outgoing {
			sender = "You"
		}
		subtitle := joinNonEmpty([]string{sender, strings.TrimSpace(entry.Meta), fmt.Sprintf("#%d", entry.MessageID)}, " • ")
		queryText := strings.TrimSpace(strings.Join([]string{preview, sender, entry.Text, entry.Body, entry.Meta}, " "))
		items = append(items, replyPickerItem{
			Entry:     entry,
			Title:     truncateInline(preview, 72),
			Subtitle:  subtitle,
			QueryText: queryText,
		})
		if len(items) >= limit {
			break
		}
	}
	return items
}

func filterReplyPickerItems(items []replyPickerItem, query string) []replyPickerItem {
	if strings.TrimSpace(query) == "" {
		return items
	}
	filtered := make([]replyPickerItem, 0, len(items))
	for _, item := range items {
		if fuzzyMatch(query, item.QueryText) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (cli *TelegramCLI) pickReplyEntry(title string, initialQuery string, entries []transcriptEntry) replyPickerResult {
	items := replyPickerItems(entries, replyPickerLimit)
	if len(items) == 0 {
		return replyPickerResult{Interactive: true}
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return replyPickerResult{Interactive: false}
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("%s %v\n", red("Could not start reply picker:"), err)
		return replyPickerResult{Interactive: false}
	}
	defer term.Restore(fd, oldState)

	out := func(format string, args ...any) {
		fmt.Printf(strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}

	query := initialQuery
	selected := 0

	render := func() []replyPickerItem {
		filtered := filterReplyPickerItems(items, query)
		if selected >= len(filtered) {
			selected = len(filtered) - 1
		}
		if selected < 0 {
			selected = 0
		}

		fmt.Print("\033[2J\033[H")
		out("%s %s\n", bold(cyan(title)), dim("(Esc to cancel, Enter to reply)"))
		out("%s %s\n\n", dim("Filter:"), query)
		if len(filtered) == 0 {
			out("%s\n", dim("  No matches"))
			return filtered
		}

		for i, item := range filtered {
			cursor := "  "
			lineStyle := func(s string) string { return s }
			if i == selected {
				cursor = yellow("➤ ")
				lineStyle = bold
			}
			out("%s%s\n", cursor, lineStyle(item.Title))
			if item.Subtitle != "" {
				out("   %s\n", dim(item.Subtitle))
			}
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

		switch b := buf[0]; b {
		case 3:
			fmt.Print("\033[2J\033[H")
			out("\n")
			return replyPickerResult{Cancelled: true, Interactive: true}
		case 13:
			if len(filtered) == 0 {
				continue
			}
			fmt.Print("\033[2J\033[H")
			out("\n")
			chosen := filtered[selected].Entry
			return replyPickerResult{Chosen: &chosen, Interactive: true}
		case 27:
			next, ok := readByteWithTimeout(fd, 15)
			if !ok || next != '[' {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return replyPickerResult{Cancelled: true, Interactive: true}
			}
			arrow, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return replyPickerResult{Cancelled: true, Interactive: true}
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
				return replyPickerResult{Cancelled: true, Interactive: true}
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
	return replyPickerResult{Cancelled: true, Interactive: true}
}

func (cli *TelegramCLI) selectReplyTarget(ctx context.Context, selector string) (*ReplyReference, error) {
	target, label := cli.currentChat()
	if target == "" {
		return nil, fmt.Errorf("no active chat; switch chats with \\to or \\msg first")
	}
	if err := cli.transcriptStore.ensureTranscriptContext(ctx, cli, target, label, replyPickerContextEntries); err != nil {
		return nil, err
	}

	entries, _ := cli.transcriptStore.transcriptSnapshot(target)
	if len(entries) == 0 {
		return nil, fmt.Errorf("no messages available for the active chat")
	}

	selector = strings.TrimSpace(selector)
	switch {
	case selector == "":
		result := cli.pickReplyEntry("Reply to message", "", entries)
		switch {
		case !result.Interactive:
			return nil, fmt.Errorf("reply picker requires an interactive terminal")
		case result.Cancelled:
			return nil, errReplyPickerCancelled
		case result.Chosen == nil:
			return nil, fmt.Errorf("no replyable messages found in the active chat")
		default:
			return replyReferenceFromEntry(*result.Chosen), nil
		}
	case strings.EqualFold(selector, "last"):
		items := replyPickerItems(entries, 1)
		if len(items) == 0 {
			return nil, fmt.Errorf("no replyable messages found in the active chat")
		}
		return replyReferenceFromEntry(items[0].Entry), nil
	default:
		messageID, err := parseMessageID(selector)
		if err == nil {
			entry, ok := findTranscriptEntryByID(entries, messageID)
			if !ok || entry.MessageID <= 0 {
				return nil, fmt.Errorf("message %d not found in the active chat transcript", messageID)
			}
			return replyReferenceFromEntry(*entry), nil
		}

		result := cli.pickReplyEntry("Reply to message", selector, entries)
		switch {
		case !result.Interactive:
			return nil, fmt.Errorf("reply picker requires an interactive terminal")
		case result.Cancelled:
			return nil, errReplyPickerCancelled
		case result.Chosen == nil:
			return nil, fmt.Errorf("no messages match %q in the active chat", selector)
		default:
			return replyReferenceFromEntry(*result.Chosen), nil
		}
	}
}
