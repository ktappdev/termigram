package ui

import "strings"

func (m *Model) syncHeaderChat() {
	m.Header.Title = "Telegram CLI"

	chat, ok := m.selectedChat()
	if !ok {
		return
	}

	label := chat.Title
	if label == "" {
		label = strings.TrimPrefix(chat.Target, "@")
	}
	if label == "" {
		return
	}

	if chat.Target != "" {
		m.Header.Title = "Telegram CLI · " + label + " (" + chat.Target + ")"
		return
	}

	m.Header.Title = "Telegram CLI · " + label
}
