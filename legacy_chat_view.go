package main

import (
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

func (cli *TelegramCLI) redrawLegacyChatView() {
	console := cli.currentLegacyConsole()
	if console == nil {
		return
	}

	target, label := cli.currentChat()
	if target == "" {
		return
	}

	_ = console.Resize()

	width, height := currentLegacyTerminalSize()
	entries, _ := cli.legacyTranscriptSnapshot(target)
	view := cli.renderActiveLegacyChatView(label, target, entries, width, height)
	_ = console.WriteBlock("\033[2J\033[H" + view + "\n")
}

func (cli *TelegramCLI) handleLegacyResize() {
	console := cli.currentLegacyConsole()
	if console == nil {
		return
	}

	_ = console.Resize()
	target, _ := cli.currentChat()
	if target == "" {
		return
	}
	cli.redrawLegacyChatView()
}

func currentLegacyTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 || height <= 0 {
		return 80, 24
	}
	return width, height
}

func renderLegacyChatView(label string, target string, entries []legacyTranscriptEntry, width int, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 1 {
		height = 2
	}

	headerRows := legacyChatHeaderRows(label, target, width, "")

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

func (cli *TelegramCLI) renderActiveLegacyChatView(label string, target string, entries []legacyTranscriptEntry, width int, height int) string {
	pendingReply := cli.pendingReplyBanner(target)
	cfg := currentInlineImageConfig()
	if !cfg.enabled() {
		return renderLegacyChatViewWithPendingReply(label, target, entries, width, height, pendingReply)
	}

	view, ok := cli.renderLegacyChatViewWithInlineImages(label, target, entries, width, height, cfg, pendingReply)
	if !ok {
		return renderLegacyChatViewWithPendingReply(label, target, entries, width, height, pendingReply)
	}
	return view
}

func truncateVisibleWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}

	chunk, _ := splitVisibleWidth(text, width-1)
	return chunk + "…"
}

func renderLegacyChatViewWithPendingReply(label string, target string, entries []legacyTranscriptEntry, width int, height int, pendingReply string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 1 {
		height = 2
	}

	headerRows := legacyChatHeaderRows(label, target, width, pendingReply)

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
