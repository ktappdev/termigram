package ui

func (m *Model) syncHeaderChat() {
	m.Header.Title = homeTitle
}

func (m Model) headerCurrentChat() string {
	chat, ok := m.selectedChat()
	if !ok {
		return ""
	}

	if m.isHomeHeaderMode() {
		return ""
	}

	return chatLabel(chat.Title, chat.Target, "")
}

func (m Model) isHomeHeaderMode() bool {
	return m.isCompactLayout() && m.Focus == focusChatList
}
