package ui

import "context"

// BackendUser is minimal account data needed by the UI.
type BackendUser struct {
	ID       int64
	Username string
}

// BackendMessage is a backend message mapped into UI-friendly fields.
type BackendMessage struct {
	ID       int64
	Text     string
	Time     string
	Sender   string
	Chat     string
	Outgoing bool
	Read     bool
}

// BackendDialog represents a chat entry for sidebar display.
type BackendDialog struct {
	Title       string
	Target      string
	LastMessage string
	LastTime    string
	Online      bool
	UnreadCount int
}

// Backend defines Telegram operations the UI needs.
type Backend interface {
	IsAuthorized(ctx context.Context) (bool, error)
	GetSelf(ctx context.Context) (*BackendUser, error)
	SendMessage(ctx context.Context, target string, text string) error
	GetMessages(ctx context.Context, target string, limit int) ([]BackendMessage, error)
	GetDialogs(ctx context.Context, limit int) ([]BackendDialog, error)
	SetActiveChat(target string, title string)
}
