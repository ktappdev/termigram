package ui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestMessageViewReflowsWithinViewportAcrossResizes(t *testing.T) {
	view := NewMessageView()
	messages := []Message{
		{ID: 1, Sender: "Ken's Butler", Chat: "Ken's Butler (@Ken592Bot)", Text: "This is a long incoming message that should wrap cleanly when the terminal gets narrower and then expand again without overflowing the viewport.", Time: "07:28", Outgoing: false},
		{ID: 2, Sender: "You", Chat: "Ken's Butler (@Ken592Bot)", Text: "This is a test message that should only appear once and should keep its bubble width inside the current viewport.", Time: "07:29", Outgoing: true},
	}

	for _, width := range []int{120, 90, 60, 40, 24, 60, 100} {
		view.Resize(width, 18)
		view.SetMessages(messages)
		assertNoRenderedLineExceedsWidth(t, view.View(), width)
	}
}

func TestAppendOutgoingMessageRemovesTranscriptEchoButAllowsRepeats(t *testing.T) {
	m := NewModel(context.Background(), nil)
	m.Width = 80
	m.Height = 30
	m.ChatList.Chats = []ChatItem{{Title: "Ken's Butler", Target: "@Ken592Bot"}}
	m.Messages.Resize(60, 18)
	m.Messages.SetMessages([]Message{{ID: 99, Text: "[Ken's Butler] > hello", Sender: "Ken's Butler", Chat: "Ken's Butler (@Ken592Bot)", Time: "07:28", Outgoing: false}})

	m.appendOutgoingMessage("hello")
	if got := len(m.Messages.Messages); got != 1 {
		t.Fatalf("expected transcript echo to be replaced by one outgoing message, got %d messages", got)
	}
	if !m.Messages.Messages[0].Outgoing || m.Messages.Messages[0].Text != "hello" {
		t.Fatalf("expected a single outgoing hello message, got %#v", m.Messages.Messages[0])
	}

	m.appendOutgoingMessage("hello")
	if got := len(m.Messages.Messages); got != 2 {
		t.Fatalf("expected repeated identical outgoing messages to be allowed, got %d messages", got)
	}
}

func TestIncomingMessageUpdatesCurrentChatAndUnreadState(t *testing.T) {
	m := NewModel(context.Background(), nil)
	m.Width = 90
	m.Height = 32
	m.ChatList.Chats = []ChatItem{
		{Title: "Alice", Target: "@alice"},
		{Title: "Bob", Target: "@bob"},
	}
	m.ChatList.SelectedIdx = 0
	m.Messages.Resize(60, 18)
	m.Messages.SetMessages([]Message{{ID: 1, Text: "Existing", Sender: "Alice", Chat: "Alice (@alice)", Time: "09:00"}})

	m.handleIncomingMessage(IncomingMessageMsg{
		Target:    "@alice",
		ChatTitle: "Alice",
		Message:   BackendMessage{ID: 2, Text: "Current chat update", Sender: "Alice", Time: "09:01"},
	})
	if got := len(m.Messages.Messages); got != 2 {
		t.Fatalf("expected current chat message to append to transcript, got %d messages", got)
	}
	if m.ChatList.Chats[m.findChatIndexByTarget("@alice")].UnreadCount != 0 {
		t.Fatalf("expected current chat unread count to stay cleared")
	}

	m.handleIncomingMessage(IncomingMessageMsg{
		Target:    "@bob",
		ChatTitle: "Bob",
		Message:   BackendMessage{ID: 3, Text: "Other chat update", Sender: "Bob", Time: "09:02"},
	})
	if got := m.ChatList.Chats[m.findChatIndexByTarget("@bob")].UnreadCount; got != 1 {
		t.Fatalf("expected other chat unread count to increment to 1, got %d", got)
	}
	if selected := m.selectedChatTarget(); selected != "@alice" {
		t.Fatalf("expected selected chat to remain @alice, got %q", selected)
	}
}

func TestModelUpdateProcessesSendResultOnce(t *testing.T) {
	m := NewModel(context.Background(), nil)
	m.Width = 80
	m.Height = 30
	m.ChatList.Chats = []ChatItem{{Title: "Ken", Target: "@ken"}}
	m.ChatList.SelectedIdx = 0
	m.Messages.Resize(60, 18)
	m.applyWindowSize(tea.WindowSizeMsg{Width: 80, Height: 30})

	updated, _ := m.Update(backendSendMsg{Text: "hello"})
	model := updated.(Model)
	if got := len(model.Messages.Messages); got != 1 {
		t.Fatalf("expected one outgoing transcript entry after send result, got %d", got)
	}
}

func assertNoRenderedLineExceedsWidth(t *testing.T, rendered string, width int) {
	t.Helper()
	for _, line := range strings.Split(rendered, "\n") {
		if lipgloss.Width(line) > width {
			t.Fatalf("rendered line width %d exceeded viewport width %d: %q", lipgloss.Width(line), width, line)
		}
	}
}
