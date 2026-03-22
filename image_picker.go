package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const imagePickerLimit = 25

var errImagePickerCancelled = errors.New("image open cancelled")

type imagePickerItem struct {
	Entry     legacyTranscriptEntry
	Title     string
	Subtitle  string
	QueryText string
}

type imagePickerResult struct {
	Chosen      *legacyTranscriptEntry
	Cancelled   bool
	Interactive bool
}

func imagePickerItems(entries []legacyTranscriptEntry, limit int) []imagePickerItem {
	if limit <= 0 {
		limit = imagePickerLimit
	}

	items := make([]imagePickerItem, 0, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Image == nil {
			continue
		}

		title := "[image]"
		if name := strings.TrimSpace(entry.Image.Name); name != "" {
			title = name
		}

		caption := imageCaptionText(entry.Body)
		subtitleParts := []string{
			strings.TrimSpace(entry.Header),
			strings.TrimSpace(entry.Meta),
		}
		if caption != "" {
			subtitleParts = append(subtitleParts, truncateInline(caption, 48))
		}

		queryText := strings.TrimSpace(strings.Join([]string{
			title,
			caption,
			entry.Header,
			entry.Meta,
		}, " "))

		items = append(items, imagePickerItem{
			Entry:     entry,
			Title:     title,
			Subtitle:  joinNonEmpty(subtitleParts, " • "),
			QueryText: queryText,
		})

		if len(items) >= limit {
			break
		}
	}

	return items
}

func imageCaptionText(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	lines := strings.SplitN(body, "\n", 2)
	if len(lines) < 2 {
		return ""
	}
	return strings.TrimSpace(lines[1])
}

func joinNonEmpty(values []string, sep string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		filtered = append(filtered, value)
	}
	return strings.Join(filtered, sep)
}

func filterImagePickerItems(items []imagePickerItem, query string) []imagePickerItem {
	if strings.TrimSpace(query) == "" {
		return items
	}

	filtered := make([]imagePickerItem, 0, len(items))
	for _, item := range items {
		if fuzzyMatch(query, item.QueryText) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (cli *TelegramCLI) pickImageEntry(title string, initialQuery string, entries []legacyTranscriptEntry) imagePickerResult {
	items := imagePickerItems(entries, imagePickerLimit)
	if len(items) == 0 {
		return imagePickerResult{Interactive: true}
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return imagePickerResult{Interactive: false}
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("%s %v\n", red("Could not start image picker:"), err)
		return imagePickerResult{Interactive: false}
	}
	defer term.Restore(fd, oldState)

	out := func(format string, args ...any) {
		fmt.Printf(strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}

	query := initialQuery
	selected := 0

	render := func() []imagePickerItem {
		filtered := filterImagePickerItems(items, query)
		if selected >= len(filtered) {
			selected = len(filtered) - 1
		}
		if selected < 0 {
			selected = 0
		}

		fmt.Print("\033[2J\033[H")
		out("%s %s\n", bold(cyan(title)), dim("(Esc to cancel, Enter to open)"))
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

		b := buf[0]
		switch b {
		case 3:
			fmt.Print("\033[2J\033[H")
			out("\n")
			return imagePickerResult{Cancelled: true, Interactive: true}
		case 13:
			if len(filtered) == 0 {
				continue
			}
			fmt.Print("\033[2J\033[H")
			out("\n")
			chosen := filtered[selected].Entry
			return imagePickerResult{Chosen: &chosen, Interactive: true}
		case 27:
			next, ok := readByteWithTimeout(fd, 15)
			if !ok || next != '[' {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return imagePickerResult{Cancelled: true, Interactive: true}
			}
			arrow, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				return imagePickerResult{Cancelled: true, Interactive: true}
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
				return imagePickerResult{Cancelled: true, Interactive: true}
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
	return imagePickerResult{Cancelled: true, Interactive: true}
}
