package ui

import "strings"

func (m *Model) syncHeaderChat() {
	m.Header.Title = "termigram"

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
		m.Header.Title = "termigram · " + label + " (" + chat.Target + ")"
		return
	}

	m.Header.Title = "termigram · " + label
}
