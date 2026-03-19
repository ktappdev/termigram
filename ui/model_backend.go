package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) loadAuthCmd() tea.Cmd {
	if m.backend == nil {
		return nil
	}
	return func() tea.Msg {
		authorized, err := m.backend.IsAuthorized(m.ctx)
		if err != nil {
			return backendAuthMsg{Err: err}
		}
		if !authorized {
			return backendAuthMsg{Err: fmt.Errorf("not authenticated")}
		}
		self, err := m.backend.GetSelf(m.ctx)
		if err != nil {
			return backendAuthMsg{Err: err}
		}
		return backendAuthMsg{Username: self.Username}
	}
}

func (m Model) loadDialogsCmd() tea.Cmd {
	if m.backend == nil {
		return nil
	}
	return func() tea.Msg {
		dialogs, err := m.backend.GetDialogs(m.ctx, 50)
		return backendDialogsMsg{Dialogs: dialogs, Err: err}
	}
}

func (m Model) loadMessagesCmd() tea.Cmd {
	if m.backend == nil {
		return nil
	}
	chat, ok := m.selectedChat()
	if !ok {
		return nil
	}
	target := m.currentTarget()
	if target == "" {
		return nil
	}
	if m.backend != nil {
		m.backend.SetActiveChat(target, chat.Title)
	}
	return func() tea.Msg {
		messages, err := m.backend.GetMessages(m.ctx, target, 20)
		return backendMessagesMsg{Target: target, ChatTitle: chat.Title, Messages: messages, Err: err}
	}
}

func (m *Model) consumeSentMessagesCmd() tea.Cmd {
	if m.lastSentSeen >= len(m.Input.Sent) {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(m.Input.Sent)-m.lastSentSeen)
	target := m.currentTarget()
	for m.lastSentSeen < len(m.Input.Sent) {
		text := strings.TrimSpace(m.Input.Sent[m.lastSentSeen])
		m.lastSentSeen++
		if text == "" {
			continue
		}
		if m.backend == nil || target == "" {
			msgText := text
			cmds = append(cmds, func() tea.Msg {
				return backendSendMsg{Text: msgText}
			})
			continue
		}
		msgText := text
		cmds = append(cmds, func() tea.Msg {
			err := m.backend.SendMessage(m.ctx, target, msgText)
			return backendSendMsg{Text: msgText, Err: err}
		})
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}
