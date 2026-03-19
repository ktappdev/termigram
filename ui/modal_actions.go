package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type backendResolveChatMsg struct {
	Chat *BackendDialog
	Err  error
}

func (m Model) openNewChatModal() (tea.Model, tea.Cmd) {
	modal := NewNewChatModal().SetChats(m.ChatList.Chats)
	modal.Width = m.Width
	modal.Height = m.Height
	m.NewChatModal = modal
	m.ActiveModal = modalNewChat
	m.lastNotice = "New chat"
	return m, nil
}

func (m Model) openSettingsModal() (tea.Model, tea.Cmd) {
	m.SettingsModal.Width = m.Width
	m.SettingsModal.Height = m.Height
	m.ActiveModal = modalSettings
	m.lastNotice = "Settings"
	return m, nil
}

func (m Model) resolveNewChatCmd(target string) tea.Cmd {
	if m.backend == nil {
		return nil
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return nil
	}
	return func() tea.Msg {
		chat, err := m.backend.ResolveChat(m.ctx, target)
		return backendResolveChatMsg{Chat: chat, Err: err}
	}
}

func (m *Model) openChat(chat ChatItem) tea.Cmd {
	if strings.TrimSpace(chat.Target) == "" {
		return nil
	}
	if strings.TrimSpace(chat.Title) == "" {
		chat.Title = chat.Target
	}
	m.upsertChatActivity(chat.Target, chat.Title, chat.LastMessage, chat.LastTime, false)
	m.selectChatByTarget(chat.Target)
	m.Focus = focusInput
	m.ActiveModal = modalNone
	m.markSelectedChatRead()
	m.syncHeaderChat()
	m.lastNotice = "Opened " + m.selectedChatTitle()
	return m.loadMessagesCmd()
}
