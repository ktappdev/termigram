package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type modalTestBackend struct {
	resolved *BackendDialog
}

func (b modalTestBackend) IsAuthorized(ctx context.Context) (bool, error) { return true, nil }
func (b modalTestBackend) GetSelf(ctx context.Context) (*BackendUser, error) {
	return &BackendUser{ID: 1, Username: "tester"}, nil
}
func (b modalTestBackend) SendMessage(ctx context.Context, target string, text string) error {
	return nil
}
func (b modalTestBackend) GetMessages(ctx context.Context, target string, limit int) ([]BackendMessage, error) {
	return nil, nil
}
func (b modalTestBackend) GetDialogs(ctx context.Context, limit int) ([]BackendDialog, error) {
	return nil, nil
}
func (b modalTestBackend) ResolveChat(ctx context.Context, target string) (*BackendDialog, error) {
	if b.resolved != nil {
		return b.resolved, nil
	}
	return &BackendDialog{Title: target, Target: target}, nil
}
func (b modalTestBackend) SetActiveChat(target string, title string) {}

func TestModelViewFitsWindowAcrossResizes(t *testing.T) {
	m := NewModel(context.Background(), nil)
	m.ChatList.Chats = []ChatItem{
		{Title: "Ken's Butler", Target: "@Ken592Bot", LastMessage: "I don't understand", LastTime: "07:29", UnreadCount: 1},
		{Title: "Alice", Target: "@alice", LastMessage: "See you soon", LastTime: "07:18"},
	}
	m.ChatList.SelectedIdx = 0
	m.Messages.SetMessages([]Message{
		{ID: 1, Sender: "Ken's Butler", Chat: "Ken's Butler (@Ken592Bot)", Text: "This is a long incoming message that should wrap cleanly as the full UI resizes.", Time: "07:28", Outgoing: false},
		{ID: 2, Sender: "You", Chat: "Ken's Butler (@Ken592Bot)", Text: "This is a test message that should stay inside the rendered window width.", Time: "07:29", Outgoing: true},
	})

	for _, size := range []struct {
		width  int
		height int
	}{
		{120, 32},
		{90, 30},
		{60, 28},
		{40, 26},
		{24, 24},
		{60, 28},
		{100, 30},
	} {
		m.applyWindowSize(tea.WindowSizeMsg{Width: size.width, Height: size.height})
		assertNoRenderedLineExceedsWidth(t, m.View(), size.width)
	}
}

func TestCtrlNOpensNewChatModalAndOpensSelectedChat(t *testing.T) {
	m := NewModel(context.Background(), modalTestBackend{})
	m.applyWindowSize(tea.WindowSizeMsg{Width: 90, Height: 30})
	m.ChatList.Chats = []ChatItem{{Title: "Alice", Target: "@alice"}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	model := updated.(Model)
	if model.ActiveModal != modalNewChat {
		t.Fatalf("expected new chat modal to open, got %v", model.ActiveModal)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.ActiveModal != modalNone {
		t.Fatalf("expected modal to close after opening chat, got %v", model.ActiveModal)
	}
	if got := model.selectedChatTarget(); got != "@alice" {
		t.Fatalf("expected @alice to be selected, got %q", got)
	}
	if model.Focus != focusInput {
		t.Fatalf("expected focus to move to input, got %v", model.Focus)
	}
	if cmd == nil {
		t.Fatalf("expected open chat to trigger a load command")
	}
}

func TestCtrlNResolveOpensTypedChat(t *testing.T) {
	m := NewModel(context.Background(), modalTestBackend{resolved: &BackendDialog{Title: "Bob", Target: "@bob"}})
	m.applyWindowSize(tea.WindowSizeMsg{Width: 90, Height: 30})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	model := updated.(Model)
	model.NewChatModal.Search.SetValue("@bob")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)
	if got := model.selectedChatTarget(); got != "@bob" {
		t.Fatalf("expected resolved chat to be selected, got %q", got)
	}
}

func TestOpenSettingsModal(t *testing.T) {
	m := NewModel(context.Background(), nil)
	m.applyWindowSize(tea.WindowSizeMsg{Width: 90, Height: 30})

	updated, _ := m.openSettingsModal()
	model := updated.(Model)
	if model.ActiveModal != modalSettings {
		t.Fatalf("expected settings modal to open, got %v", model.ActiveModal)
	}
}
