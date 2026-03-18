package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ktappdev/termigram/ui"
)

// uiBackendAdapter maps UserBackend data to ui.Backend types.
type uiBackendAdapter struct {
	backend *UserBackend
	selfID  int64
}

func newUIBackendAdapter(backend *UserBackend) ui.Backend {
	return &uiBackendAdapter{backend: backend}
}

func (a *uiBackendAdapter) IsAuthorized(ctx context.Context) (bool, error) {
	return a.backend.IsAuthorized(ctx)
}

func (a *uiBackendAdapter) GetSelf(ctx context.Context) (*ui.BackendUser, error) {
	self, err := a.backend.GetSelf(ctx)
	if err != nil {
		return nil, err
	}
	a.selfID = self.ID
	return &ui.BackendUser{
		ID:       self.ID,
		Username: self.Username,
	}, nil
}

func (a *uiBackendAdapter) SendMessage(ctx context.Context, target string, text string) error {
	return a.backend.SendMessage(ctx, target, text)
}

func (a *uiBackendAdapter) GetMessages(ctx context.Context, target string, limit int) ([]ui.BackendMessage, error) {
	msgs, err := a.backend.GetMessages(ctx, target, limit)
	if err != nil {
		return nil, err
	}

	out := make([]ui.BackendMessage, 0, len(msgs))
	for _, msg := range msgs {
		outgoing := msg.FromID == a.selfID
		sender := msg.FromName
		if outgoing {
			sender = "You"
		}
		out = append(out, ui.BackendMessage{
			Text:     msg.Message,
			Time:     time.Unix(msg.Date, 0).Format("15:04"),
			Sender:   sender,
			Chat:     target,
			Outgoing: outgoing,
			Read:     outgoing,
		})
	}

	// Keep chronological feel in UI for this stub.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (a *uiBackendAdapter) GetDialogs(ctx context.Context, limit int) ([]ui.BackendDialog, error) {
	contacts, err := a.backend.GetContacts(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > len(contacts) {
		limit = len(contacts)
	}

	dialogs := make([]ui.BackendDialog, 0, limit)
	for i := 0; i < limit; i++ {
		c := contacts[i]
		title := strings.TrimSpace(c.FirstName + " " + c.LastName)
		if title == "" {
			title = fmt.Sprintf("User %d", c.UserID)
		}
		target := fmt.Sprintf("%d", c.UserID)
		if c.Username != "" {
			target = "@" + c.Username
		}
		dialogs = append(dialogs, ui.BackendDialog{
			Title:       title,
			Target:      target,
			LastMessage: "",
			LastTime:    "",
			Online:      false,
			UnreadCount: 0,
		})
	}
	return dialogs, nil
}
