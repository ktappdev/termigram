package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

func formatTranscriptBlock(header string, body string, accent func(string) string) string {
	lines := strings.Split(body, "\n")
	rendered := make([]string, 0, len(lines)+1)
	rendered = append(rendered, header)
	for _, line := range lines {
		content := line
		if strings.TrimSpace(content) == "" {
			content = dim("(blank line)")
		}
		rendered = append(rendered, fmt.Sprintf("  %s %s", accent("│"), content))
	}
	return strings.Join(rendered, "\n")
}

func outgoingTranscriptHeader(ts string, displayName string, targetLabel string) string {
	return fmt.Sprintf("%s %s %s %s", dim("["+ts+"]"), colorize(ansiBold+ansiGreen, "YOU →"), bold(displayName), dim("("+targetLabel+")"))
}

func incomingTranscriptHeader(ts string, fromName string, fromTarget string) string {
	return fmt.Sprintf("%s %s %s %s", dim("["+ts+"]"), colorize(ansiBold+ansiBlue, "FROM ←"), bold(cyan(fromName)), dim("("+fromTarget+")"))
}

func printTranscriptMessage(header string, body string, accent func(string) string) {
	divider := dim(strings.Repeat("─", 52))
	fmt.Printf("%s\n%s\n", divider, formatTranscriptBlock(header, body, accent))
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

	printTranscriptMessage(outgoingTranscriptHeader(now.Format("15:04:05"), displayName, targetLabel), text, green)
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

	printTranscriptMessage(incomingTranscriptHeader(msgTime.Format("15:04:05"), fromName, fromTarget), msg.Message, cyan)
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
