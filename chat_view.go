package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

func (cli *TelegramCLI) redrawChatView() {
	console := cli.transcriptStore.currentConsole()
	if console == nil {
		return
	}

	target, label := cli.currentChat()
	if target == "" {
		return
	}

	if err := console.Resize(); err != nil {
		fmt.Fprintf(os.Stderr, "termigram: resize chat console: %v\n", err)
	}

	width, height := currentTerminalSize()
	entries, _ := cli.transcriptStore.transcriptSnapshot(target)
	view := cli.renderActiveChatView(label, target, entries, width, height)
	if err := console.WriteBlock("\033[2J\033[H" + view + "\n"); err != nil {
		fmt.Fprintf(os.Stderr, "termigram: write chat view: %v\n", err)
	}
}

func (cli *TelegramCLI) handleChatResize() {
	console := cli.transcriptStore.currentConsole()
	if console == nil {
		return
	}

	if err := console.Resize(); err != nil {
		fmt.Fprintf(os.Stderr, "termigram: resize chat console: %v\n", err)
	}
	target, _ := cli.currentChat()
	if target == "" {
		return
	}
	cli.redrawChatView()
}

func currentTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 || height <= 0 {
		return 80, 24
	}
	return width, height
}

func renderChatView(label string, target string, entries []transcriptEntry, width int, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 1 {
		height = 2
	}

	headerRows := chatHeaderRows(label, target, width, "")

	bodyRows := make([]string, 0, len(entries)*4+1)
	if len(entries) == 0 {
		bodyRows = append(bodyRows, dim("No messages yet."))
	} else {
		for idx, entry := range entries {
			bubble := renderTranscriptBubbleForWidth(entry.Outgoing, entry.Header, entry.Body, entry.Meta, width)
			bodyRows = append(bodyRows, strings.Split(bubble, "\n")...)
			if idx < len(entries)-1 {
				bodyRows = append(bodyRows, "")
			}
		}
	}

	maxRows := height - 1
	if maxRows < len(headerRows) {
		maxRows = len(headerRows)
	}

	availableBodyRows := maxRows - len(headerRows)
	if availableBodyRows < 0 {
		availableBodyRows = 0
	}
	if len(bodyRows) > availableBodyRows {
		bodyRows = bodyRows[len(bodyRows)-availableBodyRows:]
	}

	rows := append(append([]string{}, headerRows...), bodyRows...)
	for len(rows) < maxRows {
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
}

func (cli *TelegramCLI) renderActiveChatView(label string, target string, entries []transcriptEntry, width int, height int) string {
	pendingReply := cli.pendingReplyBanner(target)
	cfg := cli.currentInlineImageConfig()
	if !cfg.enabled() {
		return renderChatViewWithPendingReply(label, target, entries, width, height, pendingReply)
	}

	view, ok := cli.renderChatViewWithInlineImages(label, target, entries, width, height, cfg, pendingReply)
	if !ok {
		return renderChatViewWithPendingReply(label, target, entries, width, height, pendingReply)
	}
	return view
}

var ansiEscapeSequence = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(text string) string {
	return ansiEscapeSequence.ReplaceAllString(text, "")
}

func truncateVisibleWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(stripANSI(text)) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}

	chunk, _ := splitVisibleWidth(text, width-1)
	return chunk + "…"
}

func renderChatViewWithPendingReply(label string, target string, entries []transcriptEntry, width int, height int, pendingReply string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 1 {
		height = 2
	}

	headerRows := chatHeaderRows(label, target, width, pendingReply)

	bodyRows := make([]string, 0, len(entries)*4+1)
	if len(entries) == 0 {
		bodyRows = append(bodyRows, dim("No messages yet."))
	} else {
		for idx, entry := range entries {
			bubble := renderTranscriptBubbleForWidth(entry.Outgoing, entry.Header, entry.Body, entry.Meta, width)
			bodyRows = append(bodyRows, strings.Split(bubble, "\n")...)
			if idx < len(entries)-1 {
				bodyRows = append(bodyRows, "")
			}
		}
	}

	maxRows := height - 1
	if maxRows < len(headerRows) {
		maxRows = len(headerRows)
	}

	availableBodyRows := maxRows - len(headerRows)
	if availableBodyRows < 0 {
		availableBodyRows = 0
	}
	if len(bodyRows) > availableBodyRows {
		bodyRows = bodyRows[len(bodyRows)-availableBodyRows:]
	}

	rows := append(append([]string{}, headerRows...), bodyRows...)
	for len(rows) < maxRows {
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
}
