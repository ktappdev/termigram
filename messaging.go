package main

import (
	"context"
	"fmt"

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

	req := &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  text,
		RandomID: int64(uint32(user.ID))<<32 | int64(uint32(len(text))),
	}

	_, err = cli.api.MessagesSendMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		return
	}

	displayName := user.Username
	if displayName == "" {
		displayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	fmt.Printf("Message sent to @%s!\n", displayName)
}

func (cli *TelegramCLI) printMessage(msg *tg.Message) {
	fromID, _ := msg.GetFromID()
	var fromName string

	if fromID == nil {
		fromName = "Unknown"
	} else if peerUser, ok := fromID.(*tg.PeerUser); ok {
		if user, found := cli.getUserByID(peerUser.UserID); found {
			fromName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		} else {
			fromName = fmt.Sprintf("User %d", peerUser.UserID)
		}
	} else {
		fromName = "Unknown"
	}

	fmt.Printf("\n[%v] %s: %s\n> ", msg.Date, fromName, msg.Message)
}
