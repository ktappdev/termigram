package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// UserBackend implements TelegramBackend and all optional capabilities using
// MTProto via gotd/td.
type UserBackend struct {
	cli *TelegramCLI
}

func NewUserBackend(cfg Config) *UserBackend {
	return &UserBackend{cli: NewTelegramCLI(cfg.TelegramAppID, cfg.TelegramAppHash, cfg.SessionPath)}
}

// Run initializes MTProto state for CLI commands and executes fn.
func (b *UserBackend) Run(ctx context.Context, fn func(context.Context) error) error {
	fmt.Println("Connecting to Telegram...")
	return b.cli.client.Run(ctx, func(runCtx context.Context) error {
		b.cli.api = b.cli.client.API()
		b.cli.sender = message.NewSender(b.cli.client.API())

		authorized, err := b.IsAuthorized(runCtx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}
		if !authorized {
			return fmt.Errorf("not authenticated - run interactive mode first: ./termigram")
		}

		if _, err := b.GetSelf(runCtx); err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}

		if contactsBackend, ok := interface{}(b).(ContactsBackend); ok {
			if _, err := contactsBackend.GetContacts(runCtx); err != nil {
				fmt.Printf("Warning: could not load contacts: %v\n", err)
			}
		}

		return fn(runCtx)
	})
}

func (b *UserBackend) IsAuthorized(ctx context.Context) (bool, error) {
	status, err := b.cli.client.Auth().Status(ctx)
	if err != nil {
		return false, err
	}
	return status.Authorized, nil
}

func (b *UserBackend) GetSelf(ctx context.Context) (*UserOutput, error) {
	self, err := b.cli.client.Self(ctx)
	if err != nil {
		return nil, err
	}
	b.cli.cacheUser(self)
	return &UserOutput{
		ID:        self.ID,
		FirstName: self.FirstName,
		LastName:  self.LastName,
		Username:  self.Username,
		Phone:     self.Phone,
	}, nil
}

func (b *UserBackend) SendMessage(ctx context.Context, target string, text string) error {
	if b.cli.api == nil {
		return fmt.Errorf("user backend is not initialized")
	}

	user, err := b.cli.findUserByIDOrUsername(ctx, target)
	if err != nil {
		return err
	}

	resolver := peer.DefaultResolver(b.cli.api)
	inputPeer, err := resolver.ResolveDomain(ctx, fmt.Sprintf("user#%d", user.ID))
	if err != nil {
		inputPeer = &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
	}

	req := &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  text,
		RandomID: int64(uint32(user.ID))<<32 | int64(uint32(len(text))),
	}

	if _, err := b.cli.api.MessagesSendMessage(ctx, req); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (b *UserBackend) GetMessages(ctx context.Context, target string, limit int) ([]MessageOutput, error) {
	if b.cli.api == nil {
		return nil, fmt.Errorf("user backend is not initialized")
	}
	if limit <= 0 {
		limit = 10
	}

	user, err := b.cli.findUserByIDOrUsername(ctx, target)
	if err != nil {
		return nil, err
	}

	inputPeer := &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
	messages, err := b.cli.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: inputPeer, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	switch m := messages.(type) {
	case *tg.MessagesMessages:
		b.cli.cacheUsersFromClasses(m.GetUsers())
		return b.messageOutputsFromClasses(m.GetMessages(), limit), nil
	case *tg.MessagesMessagesSlice:
		b.cli.cacheUsersFromClasses(m.GetUsers())
		return b.messageOutputsFromClasses(m.GetMessages(), limit), nil
	case *tg.MessagesChannelMessages:
		b.cli.cacheUsersFromClasses(m.GetUsers())
		return b.messageOutputsFromClasses(m.GetMessages(), limit), nil
	default:
		return nil, fmt.Errorf("unsupported messages history response type: %T", messages)
	}
}

func (b *UserBackend) messageOutputsFromClasses(messages []tg.MessageClass, limit int) []MessageOutput {
	out := make([]MessageOutput, 0, limit)
	for _, msg := range messages {
		message, ok := msg.(*tg.Message)
		if !ok {
			continue
		}

		fromID := int64(0)
		fromName := "Unknown"
		if from, _ := message.GetFromID(); from != nil {
			if peerUser, ok := from.(*tg.PeerUser); ok {
				fromID = peerUser.UserID
				if u, found := b.cli.getUserByID(peerUser.UserID); found {
					fromName = strings.TrimSpace(u.FirstName + " " + u.LastName)
					if fromName == "" {
						fromName = fmt.Sprintf("User %d", peerUser.UserID)
					}
				} else {
					fromName = fmt.Sprintf("User %d", peerUser.UserID)
				}
			}
		}

		out = append(out, MessageOutput{
			ID:       int64(message.ID),
			FromID:   fromID,
			FromName: fromName,
			Message:  message.Message,
			Date:     int64(message.Date),
		})
	}
	return out
}

func (b *UserBackend) GetContacts(ctx context.Context) ([]ContactOutput, error) {
	if b.cli.api == nil {
		return nil, fmt.Errorf("user backend is not initialized")
	}

	contacts, err := b.cli.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	out := make([]ContactOutput, 0)
	switch c := contacts.(type) {
	case *tg.ContactsContacts:
		out = make([]ContactOutput, 0, len(c.Users))
		for _, user := range c.Users {
			u, ok := user.(*tg.User)
			if !ok {
				continue
			}
			b.cli.cacheUser(u)
			out = append(out, ContactOutput{
				UserID:    u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Username:  u.Username,
				Phone:     u.Phone,
			})
		}
	}

	return out, nil
}

func (b *UserBackend) GetDialogs(ctx context.Context, limit int) ([]DialogOutput, error) {
	chats, err := b.cli.fetchDialogs(ctx, limit, false)
	if err != nil {
		return nil, err
	}

	out := make([]DialogOutput, 0, len(chats))
	for _, chat := range chats {
		lastTime := int64(0)
		if !chat.LastActivity.IsZero() {
			lastTime = chat.LastActivity.Unix()
		}
		out = append(out, DialogOutput{
			Title:       chat.Label,
			Target:      chat.Target,
			LastMessage: chat.LastMessage,
			LastTime:    lastTime,
			UnreadCount: chat.UnreadCount,
		})
	}
	return out, nil
}

func (b *UserBackend) ResolveTarget(ctx context.Context, target string) (*UserOutput, error) {
	user, err := b.cli.findUserByIDOrUsername(ctx, target)
	if err != nil {
		return nil, err
	}

	return &UserOutput{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Phone:     user.Phone,
	}, nil
}

func (b *UserBackend) FindUserByUsername(ctx context.Context, username string) (*UserOutput, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if !strings.HasPrefix(username, "@") {
		username = "@" + username
	}

	return b.ResolveTarget(ctx, username)
}

func (b *UserBackend) CacheUser(user *UserOutput) {
	if user == nil {
		return
	}
	b.cli.cacheUser(&tg.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Phone:     user.Phone,
	})
}

func (b *UserBackend) FindMatchingUsernames(prefix string, limit int) []string {
	return b.cli.findMatchingUsernames(prefix, limit)
}
