package main

import (
	"context"
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
		outgoing := msg.Outgoing || (a.selfID != 0 && msg.FromID == a.selfID)
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
	dialogs, err := a.backend.GetDialogs(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]ui.BackendDialog, 0, len(dialogs))
	for _, dialog := range dialogs {
		lastTime := ""
		if dialog.LastTime > 0 {
			lastTime = time.Unix(dialog.LastTime, 0).Format("15:04")
		}
		out = append(out, ui.BackendDialog{
			Title:       dialog.Title,
			Target:      dialog.Target,
			LastMessage: dialog.LastMessage,
			LastTime:    lastTime,
			Online:      false,
			UnreadCount: dialog.UnreadCount,
		})
	}
	return out, nil
}
