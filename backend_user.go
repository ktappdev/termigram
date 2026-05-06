package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

// UserBackend implements TelegramBackend and all optional capabilities using
// MTProto via gotd/td.
type UserBackend struct {
	cli *TelegramCLI
}

func NewUserBackend(cfg Config) *UserBackend {
	return &UserBackend{cli: NewTelegramCLIWithConfig(cfg)}
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
		return false, fmt.Errorf("failed to get auth status: %w", err)
	}
	return status.Authorized, nil
}

func (b *UserBackend) GetSelf(ctx context.Context) (*UserOutput, error) {
	self, err := b.cli.client.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get self: %w", err)
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

func (b *UserBackend) SendMessage(ctx context.Context, target string, text string, opts SendOptions) error {
	_, err := b.sendText(ctx, target, text, opts)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (b *UserBackend) sendText(ctx context.Context, target string, text string, opts SendOptions) (int64, error) {
	if b.cli.api == nil || b.cli.sender == nil {
		return 0, fmt.Errorf("user backend is not initialized")
	}
	request := b.cli.sender.Resolve(target)
	if opts.ReplyToMessageID > 0 {
		request.Reply(int(opts.ReplyToMessageID))
	}
	updates, err := request.Text(ctx, text)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}
	return sentMessageIDFromUpdates(updates), nil
}

func (b *UserBackend) SendImage(ctx context.Context, target string, source string, caption string, opts SendOptions) error {
	if b.cli.api == nil || b.cli.sender == nil {
		return fmt.Errorf("user backend is not initialized")
	}

	prepared, err := prepareImageSourceWithLimit(ctx, source, b.cli.httpClient, b.cli.maxRemoteImageBytesLimit())
	if err != nil {
		return fmt.Errorf("failed to prepare image source: %w", err)
	}
	defer prepared.Cleanup()

	_, err = b.sendPreparedImage(ctx, target, prepared, caption, opts)
	if err != nil {
		return fmt.Errorf("failed to send image: %w", err)
	}
	return nil
}

func (b *UserBackend) sendPreparedImage(ctx context.Context, target string, prepared preparedImageSource, caption string, opts SendOptions) (int64, error) {
	request := b.cli.sender.Resolve(target)
	if opts.ReplyToMessageID > 0 {
		request.Reply(int(opts.ReplyToMessageID))
	}
	captionOptions := imageCaptionOptions(caption)

	var (
		updates tg.UpdatesClass
		err     error
	)
	if prepared.SendAsFile {
		updates, err = request.Upload(message.FromPath(prepared.Path)).File(ctx, captionOptions...)
		if err != nil {
			return 0, fmt.Errorf("failed to send image file: %w", err)
		}
	} else {
		updates, err = request.Upload(message.FromPath(prepared.Path)).Photo(ctx, captionOptions...)
		if err != nil {
			return 0, fmt.Errorf("failed to send photo: %w", err)
		}
	}
	return sentMessageIDFromUpdates(updates), nil
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
		return nil, fmt.Errorf("failed to resolve target: %w", err)
	}

	inputPeer := &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
	response, err := b.cli.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: inputPeer, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	users, messageClasses, err := unpackMessagesResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack messages: %w", err)
	}
	b.cli.cacheUsersFromClasses(users)
	return b.messageOutputsFromClasses(ctx, messageClasses, limit), nil
}

func (b *UserBackend) messageOutputsFromClasses(ctx context.Context, classes []tg.MessageClass, limit int) []MessageOutput {
	messages := messageClassesToMessages(classes)
	replyRefs := resolveReplyReferencesForMessages(ctx, b.cli, messages)
	out := make([]MessageOutput, 0, limit)
	for _, message := range messages {
		out = append(out, messageOutputFromTGMessageWithReply(b.cli, message, replyRefs[int64(message.ID)]))
	}
	return out
}

func imageCaptionOptions(caption string) []message.StyledTextOption {
	caption = strings.TrimSpace(caption)
	if caption == "" {
		return nil
	}
	return []message.StyledTextOption{styling.Plain(caption)}
}

func (b *UserBackend) GetContacts(ctx context.Context) ([]ContactOutput, error) {
	if b.cli.api == nil {
		return nil, fmt.Errorf("user backend is not initialized")
	}

	contacts, err := b.cli.fetchContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	return contacts, nil
}

func (b *UserBackend) GetDialogs(ctx context.Context, limit int) ([]DialogOutput, error) {
	chats, err := b.cli.fetchDialogs(ctx, limit, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
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
		return nil, fmt.Errorf("failed to resolve target: %w", err)
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
