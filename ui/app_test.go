package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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
