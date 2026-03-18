package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

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
	now := time.Now()
	cli.setCurrentChat(targetLabel, displayName)
	cli.markChatActivity(targetLabel, text, now)

	divider := dim(strings.Repeat("─", 46))
	ts := dim("[" + now.Format("15:04:05") + "]")
	fmt.Printf("%s\n", divider)
	fmt.Printf("%s %s %s %s\n", ts, colorize(ansiBold+ansiGreen, "→ YOU"), bold(displayName), dim("("+targetLabel+")"))
	for _, line := range strings.Split(text, "\n") {
		fmt.Printf("  %s %s\n", green("›"), line)
	}
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
	fromName := "Unknown"
	fromTarget := "unknown"

	if peerUser, ok := fromID.(*tg.PeerUser); ok {
		if user, found := cli.getUserByID(peerUser.UserID); found {
			fromName = strings.TrimSpace(user.FirstName + " " + user.LastName)
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

	msgTime := time.Unix(int64(msg.Date), 0)
	ts := msgTime.Format("15:04:05")
	cli.markChatActivity(fromTarget, msg.Message, msgTime)
	activeTarget, activeLabel := cli.currentChat()
	incomingTarget := normalizeUsername(fromTarget)
	mismatch := activeTarget != "" && normalizeUsername(activeTarget) != incomingTarget

	divider := dim(strings.Repeat("─", 46))
	fmt.Printf("%s\n", divider)
	fmt.Printf("%s %s %s %s\n", dim("["+ts+"]"), colorize(ansiBold+ansiBlue, "← FROM"), bold(cyan(fromName)), dim("("+fromTarget+")"))
	if mismatch {
		fmt.Printf("%s %s %s\n", yellow("↪ Chat context:"), dim("currently focused on"), bold(activeLabel))
	}
	for _, line := range strings.Split(msg.Message, "\n") {
		fmt.Printf("  %s %s\n", cyan("‹"), line)
	}
}
