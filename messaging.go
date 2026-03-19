package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

type transcriptTheme struct {
	border func(string) string
	fill   func(string) string
	title  func(string) string
	meta   func(string) string
}

func outgoingTranscriptHeader(displayName string, targetLabel string) string {
	return fmt.Sprintf("You → %s (%s)", displayName, targetLabel)
}

func incomingTranscriptHeader(fromName string, fromTarget string) string {
	return fmt.Sprintf("%s (%s)", fromName, fromTarget)
}

func outgoingTranscriptMeta(ts string) string {
	return ts + "  ✓"
}

func incomingTranscriptMeta(ts string) string {
	return ts
}

func transcriptWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func transcriptBubbleWidth(totalWidth int) int {
	if totalWidth <= 0 {
		return 60
	}
	available := totalWidth - 2
	if available < 16 {
		return maxInt(totalWidth, 1)
	}
	bubbleWidth := (available * 3) / 4
	if bubbleWidth < 18 {
		bubbleWidth = available
	}
	if bubbleWidth > available {
		bubbleWidth = available
	}
	return bubbleWidth
}

func transcriptPadding(totalWidth int) int {
	switch {
	case totalWidth < 28:
		return 0
	case totalWidth < 40:
		return 1
	default:
		return 2
	}
}

func transcriptThemeFor(outgoing bool) transcriptTheme {
	if outgoing {
		fillStyle := ansiBgSoftBlue + ansiWhite
		return transcriptTheme{
			border: func(text string) string { return colorize(ansiBold+ansiBlue, text) },
			fill:   func(text string) string { return colorize(fillStyle, text) },
			title:  func(text string) string { return colorize(fillStyle+ansiBold, text) },
			meta:   func(text string) string { return colorize(fillStyle+ansiDim, text) },
		}
	}

	fillStyle := ansiBgSoftGreen + ansiWhite
	return transcriptTheme{
		border: func(text string) string { return colorize(ansiBold+ansiGreen, text) },
		fill:   func(text string) string { return colorize(fillStyle, text) },
		title:  func(text string) string { return colorize(fillStyle+ansiBold, text) },
		meta:   func(text string) string { return colorize(fillStyle+ansiDim, text) },
	}
}

func renderTranscriptBubble(outgoing bool, header string, body string, meta string) string {
	totalWidth := transcriptWidth()
	bubbleWidth := transcriptBubbleWidth(totalWidth)
	padding := transcriptPadding(totalWidth)
	innerWidth := bubbleWidth - 2 - (padding * 2)
	if innerWidth < 1 {
		innerWidth = 1
	}

	theme := transcriptThemeFor(outgoing)
	contentLines := make([]string, 0)
	contentLines = append(contentLines, wrapTranscriptText(header, innerWidth)...)
	contentLines = append(contentLines, wrapTranscriptText(body, innerWidth)...)
	contentLines = append(contentLines, wrapTranscriptText(meta, innerWidth)...)
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}

	blockWidth := innerWidth + (padding * 2) + 2
	indent := 0
	if outgoing && totalWidth > blockWidth {
		indent = totalWidth - blockWidth
	}
	prefix := strings.Repeat(" ", maxInt(indent, 0))

	rendered := make([]string, 0, len(contentLines))
	for i, line := range contentLines {
		styledLine := theme.fill(transcriptInnerLine(line, innerWidth, padding))
		switch {
		case i == 0:
			styledLine = theme.title(transcriptInnerLine(line, innerWidth, padding))
			rendered = append(rendered, prefix+theme.border("╭")+styledLine+theme.border("╮"))
		case i == len(contentLines)-1:
			styledLine = theme.meta(transcriptInnerLine(line, innerWidth, padding))
			rendered = append(rendered, prefix+theme.border("╰")+styledLine+theme.border("╯"))
		default:
			rendered = append(rendered, prefix+theme.border("│")+styledLine+theme.border("│"))
		}
	}

	return strings.Join(rendered, "\n")
}

func transcriptInnerLine(line string, width int, padding int) string {
	inner := visiblePadRight(line, width)
	return strings.Repeat(" ", padding) + inner + strings.Repeat(" ", padding)
}

func wrapTranscriptText(text string, width int) []string {
	if width < 1 {
		width = 1
	}

	lines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		remaining := line
		for runewidth.StringWidth(remaining) > width {
			chunk, rest := splitVisibleWidth(remaining, width)
			wrapped = append(wrapped, chunk)
			remaining = rest
		}
		wrapped = append(wrapped, remaining)
	}
	return wrapped
}

func splitVisibleWidth(text string, width int) (string, string) {
	if width < 1 {
		return "", text
	}

	visible := 0
	lastByte := 0
	for idx, r := range text {
		runeWidth := runewidth.RuneWidth(r)
		if runeWidth == 0 {
			runeWidth = 1
		}
		if visible+runeWidth > width {
			if lastByte == 0 {
				next := idx + len(string(r))
				return text[:next], text[next:]
			}
			return text[:lastByte], text[lastByte:]
		}
		visible += runeWidth
		lastByte = idx + len(string(r))
	}
	return text, ""
}

func visiblePadRight(text string, width int) string {
	pad := width - runewidth.StringWidth(text)
	if pad <= 0 {
		return text
	}
	return text + strings.Repeat(" ", pad)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func printTranscriptMessage(outgoing bool, header string, body string, meta string) {
	fmt.Printf("\n%s\n", renderTranscriptBubble(outgoing, header, body, meta))
}

func (cli *TelegramCLI) sendMessage(ctx context.Context, target string, text string) {
	user, err := cli.findUserByIDOrUsername(ctx, target)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	resolver := peer.DefaultResolver(cli.api)
	inputPeer, err := resolver.ResolveDomain(ctx, fmt.Sprintf("user#%d", user.ID))
	if err != nil {
		inputPeer = &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		}
	}

	randomID := time.Now().UnixNano() ^ (user.ID << 16)
	req := &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  text,
		RandomID: randomID,
	}

	_, err = cli.api.MessagesSendMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		return
	}

	displayName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if displayName == "" {
		displayName = user.Username
	}
	if displayName == "" {
		displayName = fmt.Sprintf("User %d", user.ID)
	}

	targetLabel := fmt.Sprintf("%d", user.ID)
	if user.Username != "" {
		targetLabel = "@" + user.Username
	}

	if text == "" {
		return
	}

	now := time.Now()
	cli.setCurrentChat(targetLabel, displayName)
	cli.markChatActivity(targetLabel, text, now)

	printTranscriptMessage(true, outgoingTranscriptHeader(displayName, targetLabel), text, outgoingTranscriptMeta(now.Format("15:04:05")))
}

func (cli *TelegramCLI) shouldPrintIncoming(msg *tg.Message, fromTarget string) bool {
	key := fmt.Sprintf("%d:%d:%s", msg.ID, msg.Date, normalizeUsername(fromTarget))
	now := time.Now()

	cli.mu.Lock()
	defer cli.mu.Unlock()

	for k, seenAt := range cli.seenIncoming {
		if now.Sub(seenAt) > 2*time.Minute {
			delete(cli.seenIncoming, k)
		}
	}
	if _, exists := cli.seenIncoming[key]; exists {
		return false
	}
	cli.seenIncoming[key] = now
	return true
}

func (cli *TelegramCLI) printMessage(msg *tg.Message) {
	if msg.GetOut() {
		return
	}

	fromID, _ := msg.GetFromID()
	fromName := "Unknown sender"
	fromTarget := "unknown"

	if peerUser, ok := fromID.(*tg.PeerUser); ok {
		if user, found := cli.getUserByID(peerUser.UserID); found {
			fromName = strings.TrimSpace(user.FirstName + " " + user.LastName)
			if fromName == "" && user.Username != "" {
				fromName = "@" + user.Username
			}
			if fromName == "" {
				fromName = fmt.Sprintf("User %d", peerUser.UserID)
			}
			if user.Username != "" {
				fromTarget = "@" + user.Username
			} else {
				fromTarget = fmt.Sprintf("%d", peerUser.UserID)
			}
		} else {
			fromName = fmt.Sprintf("User %d", peerUser.UserID)
			fromTarget = fmt.Sprintf("%d", peerUser.UserID)
		}
	}

	if !cli.shouldPrintIncoming(msg, fromTarget) {
		return
	}

	if msg.Message == "" {
		return
	}

	msgTime := time.Unix(int64(msg.Date), 0)
	cli.markChatActivity(fromTarget, msg.Message, msgTime)
	activeTarget, activeLabel := cli.currentChat()
	incomingTarget := normalizeUsername(fromTarget)
	mismatch := activeTarget != "" && normalizeUsername(activeTarget) != incomingTarget

	printTranscriptMessage(false, incomingTranscriptHeader(fromName, fromTarget), msg.Message, incomingTranscriptMeta(msgTime.Format("15:04:05")))
	if mismatch {
		focusLabel := activeLabel
		if strings.TrimSpace(focusLabel) == "" {
			focusLabel = activeTarget
		}
		if strings.TrimSpace(focusLabel) == "" {
			focusLabel = "another chat"
		}
		fmt.Printf("  %s %s %s\n", yellow("↪"), dim("Current focus:"), bold(focusLabel))
	}
}
