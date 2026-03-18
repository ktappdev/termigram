package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"
)

type dialogsEnvelope interface {
	GetDialogs() []tg.DialogClass
	GetMessages() []tg.MessageClass
	GetUsers() []tg.UserClass
}

func (cli *TelegramCLI) fetchDialogs(ctx context.Context, limit int, unreadOnly bool) ([]CachedChat, error) {
	if cli.api == nil {
		return nil, fmt.Errorf("telegram client is not initialized")
	}

	fetchLimit := limit
	if fetchLimit <= 0 {
		fetchLimit = 100
	}
	if fetchLimit < 50 {
		fetchLimit = 50
	}

	resp, err := cli.api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      fetchLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	payload, ok := resp.(dialogsEnvelope)
	if !ok {
		return nil, fmt.Errorf("unsupported dialogs response type: %T", resp)
	}

	cli.cacheUsersFromClasses(payload.GetUsers())

	messagesByID := make(map[int]*tg.Message, len(payload.GetMessages()))
	for _, msg := range payload.GetMessages() {
		if m, ok := msg.(*tg.Message); ok {
			messagesByID[m.ID] = m
		}
	}

	out := make([]CachedChat, 0, len(payload.GetDialogs()))
	for _, dialogClass := range payload.GetDialogs() {
		dialog, ok := dialogClass.(*tg.Dialog)
		if !ok {
			continue
		}
		if unreadOnly && dialog.UnreadCount == 0 {
			continue
		}

		peerUser, ok := dialog.Peer.(*tg.PeerUser)
		if !ok {
			continue
		}

		user, found := cli.getUserByID(peerUser.UserID)
		chat := CachedChat{UnreadCount: dialog.UnreadCount}
		if found {
			chat = cachedChatFromUser(user)
			chat.UnreadCount = dialog.UnreadCount
		} else {
			chat.Label = fmt.Sprintf("User %d", peerUser.UserID)
			chat.Target = fmt.Sprintf("%d", peerUser.UserID)
		}

		if top, ok := messagesByID[dialog.TopMessage]; ok {
			chat.LastMessage = strings.TrimSpace(top.Message)
			chat.LastActivity = time.Unix(int64(top.Date), 0)
			cli.markChatActivity(chat.Target, top.Message, chat.LastActivity)
		}
		cli.setChatUnreadCount(chat.Target, dialog.UnreadCount)

		out = append(out, chat)
		if limit > 0 && len(out) >= limit {
			break
		}
	}

	return out, nil
}
