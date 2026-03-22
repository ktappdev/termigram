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

var sendPreparedImageWithBackend = func(ctx context.Context, backend *UserBackend, target string, prepared preparedImageSource, caption string) error {
	return backend.sendPreparedImage(ctx, target, prepared, caption)
}

type transcriptTheme struct {
	border func(string) string
	fill   func(string) string
	title  func(string) string
	meta   func(string) string
}

func outgoingTranscriptHeader(displayName string, targetLabel string, includeContext bool) string {
	if includeContext && strings.TrimSpace(displayName) != "" && strings.TrimSpace(targetLabel) != "" {
		return fmt.Sprintf("You • %s", targetLabel)
	}
	return "You"
}

func incomingTranscriptHeader(fromName string, fromTarget string, includeContext bool) string {
	name := strings.TrimSpace(fromName)
	if name == "" || name == "Unknown sender" {
		name = strings.TrimSpace(fromTarget)
	}
	if name == "" {
		name = "Unknown"
	}
	if includeContext && strings.TrimSpace(fromTarget) != "" && normalizeLabelValue(name) != normalizeLabelValue(fromTarget) {
		return fmt.Sprintf("%s • %s", name, fromTarget)
	}
	return name
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
	case totalWidth < 32:
		return 0
	default:
		return 1
	}
}

func transcriptThemeFor(outgoing bool) transcriptTheme {
	if outgoing {
		fillStyle := ansiBgSoftBlue + ansiWhite
		return transcriptTheme{
			border: func(text string) string { return colorize(ansiDim+ansiBlue, text) },
			fill:   func(text string) string { return colorize(fillStyle, text) },
			title:  func(text string) string { return colorize(fillStyle+ansiBold, text) },
			meta:   func(text string) string { return colorize(fillStyle+ansiDim, text) },
		}
	}

	fillStyle := ansiBgSoftGreen + ansiWhite
	return transcriptTheme{
		border: func(text string) string { return colorize(ansiDim+ansiGreen, text) },
		fill:   func(text string) string { return colorize(fillStyle, text) },
		title:  func(text string) string { return colorize(fillStyle+ansiBold, text) },
		meta:   func(text string) string { return colorize(fillStyle+ansiDim, text) },
	}
}

func renderTranscriptBubble(outgoing bool, header string, body string, meta string) string {
	return renderTranscriptBubbleForWidth(outgoing, header, body, meta, transcriptWidth())
}

func renderTranscriptBubbleForWidth(outgoing bool, header string, body string, meta string, totalWidth int) string {
	bubbleWidth := transcriptBubbleWidth(totalWidth)
	padding := transcriptPadding(totalWidth)
	innerWidth := bubbleWidth - 2 - (padding * 2)
	if innerWidth < 1 {
		innerWidth = 1
	}

	theme := transcriptThemeFor(outgoing)
	headerLines := wrapTranscriptText(header, innerWidth)
	bodyLines := wrapTranscriptText(body, innerWidth)
	metaLines := wrapTranscriptText(meta, innerWidth)

	contentLines := make([]string, 0, len(headerLines)+len(bodyLines)+len(metaLines))
	contentLines = append(contentLines, headerLines...)
	contentLines = append(contentLines, bodyLines...)
	contentLines = append(contentLines, metaLines...)
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}

	metaStart := len(headerLines) + len(bodyLines)
	blockWidth := innerWidth + (padding * 2) + 2
	indent := 0
	if outgoing && totalWidth > blockWidth {
		indent = totalWidth - blockWidth
	}
	prefix := strings.Repeat(" ", maxInt(indent, 0))

	rendered := make([]string, 0, len(contentLines))
	for i, line := range contentLines {
		styledLine := theme.fill(transcriptInnerLine(line, innerWidth, padding, false))
		if i >= metaStart {
			styledLine = theme.meta(transcriptInnerLine(line, innerWidth, padding, true))
		}
		switch {
		case i == 0:
			styledLine = theme.title(transcriptInnerLine(line, innerWidth, padding, false))
			rendered = append(rendered, prefix+theme.border("╭")+styledLine+theme.border("╮"))
		case i == len(contentLines)-1:
			rendered = append(rendered, prefix+theme.border("╰")+styledLine+theme.border("╯"))
		default:
			rendered = append(rendered, prefix+theme.border("│")+styledLine+theme.border("│"))
		}
	}

	return strings.Join(rendered, "\n")
}

func transcriptInnerLine(line string, width int, padding int, alignRight bool) string {
	inner := visiblePad(line, width, alignRight)
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

func visiblePad(text string, width int, alignRight bool) string {
	pad := width - runewidth.StringWidth(text)
	if pad <= 0 {
		return text
	}
	if alignRight {
		return strings.Repeat(" ", pad) + text
	}
	return text + strings.Repeat(" ", pad)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func normalizeLabelValue(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "@")
	return value
}

func printTranscriptMessage(outgoing bool, header string, body string, meta string) {
	fmt.Printf("%s\n", renderTranscriptBubble(outgoing, header, body, meta))
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
	_ = cli.ensureLegacyTranscript(ctx, targetLabel, displayName)

	entry := legacyTranscriptEntry{
		Outgoing: true,
		Header:   outgoingTranscriptHeader(displayName, targetLabel, false),
		Body:     text,
		Meta:     outgoingTranscriptMeta(now.Format("15:04:05")),
	}
	cli.appendLegacyTranscriptEntry(targetLabel, entry)

	if interactiveTTYAvailable() {
		return
	}

	printTranscriptMessage(true, entry.Header, entry.Body, entry.Meta)
}

func (cli *TelegramCLI) sendImage(ctx context.Context, target string, label string, source string, caption string) error {
	backend := &UserBackend{cli: cli}
	prepared, err := prepareImageSource(ctx, source)
	if err != nil {
		return err
	}
	defer prepared.Cleanup()

	if err := sendPreparedImageWithBackend(ctx, backend, target, prepared, caption); err != nil {
		return err
	}

	attachment := &ImageAttachment{
		Kind:     imageKind,
		Name:     prepared.Name,
		MIMEType: prepared.MIMEType,
	}
	cachedPath := prepared.Path
	if !prepared.Persistent {
		cachedPath, err = cacheOutboundImageFunc(target, attachment, prepared.Path)
		if err != nil {
			return err
		}
	}
	if cachedPath != "" {
		attachment.CachedPath = cachedPath
	}
	cli.recordOutgoingImage(target, label, attachment, caption)
	if interactiveTTYAvailable() {
		cli.redrawLegacyChatView()
		return nil
	}

	entry := legacyTranscriptEntry{
		Outgoing: true,
		Header:   outgoingTranscriptHeader(label, target, false),
		Body:     imagePlaceholderBody(0, prepared.Name, caption),
		Meta:     outgoingTranscriptMeta(time.Now().Format("15:04:05")),
		Image:    attachment,
	}
	printTranscriptMessage(true, entry.Header, entry.Body, entry.Meta)
	return nil
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

	attachment, _ := imageAttachmentFromMessage(msg)
	body := messageBodyText(int64(msg.ID), msg.Message, attachment)
	if strings.TrimSpace(body) == "" {
		return
	}

	msgTime := time.Unix(int64(msg.Date), 0)
	cli.markChatActivity(fromTarget, messagePreviewText(msg.Message, attachment), msgTime)
	activeTarget, activeLabel := cli.currentChat()
	incomingTarget := normalizeUsername(fromTarget)
	mismatch := activeTarget == "" || normalizeUsername(activeTarget) != incomingTarget
	if mismatch {
		cli.incrementChatUnreadCount(fromTarget)
	} else {
		cli.clearChatUnreadCount(fromTarget)
	}

	entry := legacyTranscriptEntry{
		MessageID: int64(msg.ID),
		Outgoing:  false,
		Header:    incomingTranscriptHeader(fromName, fromTarget, mismatch),
		Body:      body,
		Meta:      incomingTranscriptMeta(msgTime.Format("15:04:05")),
		Image:     attachment,
	}
	cli.appendLegacyTranscriptEntry(fromTarget, entry)

	if !mismatch && cli.currentLegacyConsole() != nil {
		cli.redrawLegacyChatView()
		return
	}
	if !mismatch && interactiveTTYAvailable() {
		return
	}

	if console := cli.currentLegacyConsole(); console != nil {
		text := renderTranscriptBubbleForWidth(false, entry.Header, entry.Body, entry.Meta, transcriptWidth())
		if mismatch {
			focusLabel := activeLabel
			if strings.TrimSpace(focusLabel) == "" {
				focusLabel = activeTarget
			}
			if strings.TrimSpace(focusLabel) == "" {
				focusLabel = "another chat"
			}
			text += "\n" + fmt.Sprintf("  %s %s %s", yellow("↪"), dim("Current focus:"), bold(focusLabel))
		}
		_ = console.WriteString(text)
		return
	}

	printTranscriptMessage(false, entry.Header, entry.Body, entry.Meta)
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
