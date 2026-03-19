package ui

func (m *Model) syncHeaderChat() {
	m.Header.Title = homeTitle
}

func (m Model) headerCurrentChat() string {
	if m.Focus == focusChatList {
		return ""
	}

	chat, ok := m.selectedChat()
	if !ok {
		return ""
	}

	return chatLabel(chat.Title, chat.Target, "")
}
