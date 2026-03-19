package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) applyDialogs(dialogs []BackendDialog) tea.Cmd {
	selectedTarget := m.selectedChatTarget()
	m.ChatList.Chats = make([]ChatItem, 0, len(dialogs))
	for _, dialog := range dialogs {
		m.ChatList.Chats = append(m.ChatList.Chats, ChatItem{
			Title:       dialog.Title,
			Target:      dialog.Target,
			LastMessage: dialog.LastMessage,
			LastTime:    dialog.LastTime,
			Online:      dialog.Online,
			UnreadCount: dialog.UnreadCount,
		})
	}
	if len(m.ChatList.Chats) == 0 {
		m.ChatList.SelectedIdx = 0
		m.Messages.SetMessages(nil)
		m.syncHeaderChat()
		return nil
	}
	filtered, _ := m.ChatList.filteredChats()
	if !m.selectChatByTarget(selectedTarget) && m.ChatList.SelectedIdx >= len(filtered) {
		m.ChatList.SelectedIdx = 0
	}
	m.markSelectedChatRead()
	m.syncHeaderChat()
	return m.loadMessagesCmd()
}

func (m *Model) handleIncomingMessage(msg IncomingMessageMsg) {
	target := strings.TrimSpace(msg.Target)
	if target == "" {
		return
	}
	chatTitle := strings.TrimSpace(msg.ChatTitle)
	if chatTitle == "" {
		chatTitle = strings.TrimSpace(msg.Message.Sender)
	}
	if chatTitle == "" {
		chatTitle = target
	}

	isCurrentChat := normalizeChatTarget(target) == normalizeChatTarget(m.currentTarget())
	m.upsertChatActivity(target, chatTitle, msg.Message.Text, msg.Message.Time, !isCurrentChat)
	if isCurrentChat {
		m.Messages.AppendMessage(Message{
			ID:       msg.Message.ID,
			Text:     msg.Message.Text,
			Time:     msg.Message.Time,
			Sender:   msg.Message.Sender,
			Chat:     chatLabel(chatTitle, target, msg.Message.Chat),
			Outgoing: msg.Message.Outgoing,
			Read:     msg.Message.Read,
		}, true)
		m.markSelectedChatRead()
		return
	}
	m.lastNotice = "New message from " + chatTitle
}

func (m *Model) appendOutgoingMessage(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	m.removeTrailingTranscriptEcho(text)
	messageTime := time.Now().Format("15:04")
	m.Messages.AppendMessage(Message{
		Text:     text,
		Time:     messageTime,
		Sender:   "You",
		Chat:     chatLabel(m.selectedChatTitle(), m.currentTarget(), ""),
		Outgoing: true,
		Read:     false,
	}, true)
	m.upsertChatActivity(m.currentTarget(), m.selectedChatTitle(), text, messageTime, false)
	m.markSelectedChatRead()
}

func (m *Model) removeTrailingTranscriptEcho(sentText string) {
	if len(m.Messages.Messages) == 0 {
		return
	}

	trimmedSent := strings.TrimSpace(sentText)
	filtered := m.Messages.Messages[:0]
	for i, msg := range m.Messages.Messages {
		if i >= len(m.Messages.Messages)-2 && isTranscriptEcho(msg.Text, trimmedSent) && !msg.Outgoing {
			continue
		}
		filtered = append(filtered, msg)
	}
	m.Messages.Messages = filtered
}

func isTranscriptEcho(text string, sentText string) bool {
	normalized, ok := transcriptEchoPayload(text)
	if !ok || sentText == "" {
		return false
	}
	return normalized == sentText
}

func transcriptEchoPayload(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}

	original := trimmed
	if strings.HasPrefix(trimmed, "[") {
		if close := strings.Index(trimmed, "]"); close != -1 {
			trimmed = strings.TrimSpace(trimmed[close+1:])
		}
	}
	if !strings.HasPrefix(trimmed, ">") {
		return "", false
	}

	normalized := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
	if normalized == "" || original == normalized {
		return "", false
	}

	return normalized, true
}

func (m *Model) handleDeleteShortcut() (tea.Model, tea.Cmd) {
	switch m.Focus {
	case focusChatList:
		filtered, indices := m.ChatList.filteredChats()
		if len(filtered) == 0 {
			m.lastNotice = "No chat to delete"
			return m, nil
		}
		sel := m.ChatList.SelectedIdx
		if sel < 0 || sel >= len(filtered) {
			sel = 0
		}
		actualIdx := indices[sel]
		deleted := m.ChatList.Chats[actualIdx].Title
		m.ChatList.Chats = append(m.ChatList.Chats[:actualIdx], m.ChatList.Chats[actualIdx+1:]...)
		if m.ChatList.SelectedIdx >= len(filtered)-1 && m.ChatList.SelectedIdx > 0 {
			m.ChatList.SelectedIdx--
		}
		m.lastNotice = "Deleted chat: " + deleted
		return m, m.loadMessagesCmd()
	case focusMessages:
		if !m.Messages.RemoveLastMessage() {
			m.lastNotice = "No message to delete"
			return m, nil
		}
		m.lastNotice = "Deleted latest message"
		return m, nil
	default:
		m.lastNotice = "Delete works in chats/messages panels"
		return m, nil
	}
}

func (m *Model) markSelectedChatRead() {
	chat, ok := m.selectedChat()
	if !ok {
		return
	}
	idx := m.findChatIndexByTarget(chat.Target)
	if idx >= 0 {
		m.ChatList.Chats[idx].UnreadCount = 0
	}
	if m.backend != nil {
		m.backend.SetActiveChat(chat.Target, chat.Title)
	}
}

func (m *Model) upsertChatActivity(target string, title string, lastMessage string, lastTime string, incrementUnread bool) {
	if strings.TrimSpace(target) == "" {
		return
	}
	selectedTarget := m.selectedChatTarget()
	idx := m.findChatIndexByTarget(target)
	item := ChatItem{Title: title, Target: target, LastMessage: lastMessage, LastTime: lastTime}
	if idx >= 0 {
		item = m.ChatList.Chats[idx]
		if strings.TrimSpace(title) != "" {
			item.Title = title
		}
		if strings.TrimSpace(lastMessage) != "" {
			item.LastMessage = strings.TrimSpace(lastMessage)
		}
		if strings.TrimSpace(lastTime) != "" {
			item.LastTime = lastTime
		}
		if incrementUnread {
			item.UnreadCount++
		} else {
			item.UnreadCount = 0
		}
		m.ChatList.Chats = append(m.ChatList.Chats[:idx], m.ChatList.Chats[idx+1:]...)
	} else {
		item.Title = fallbackChatTitle(title, target)
		item.LastMessage = strings.TrimSpace(lastMessage)
		item.LastTime = strings.TrimSpace(lastTime)
		if incrementUnread {
			item.UnreadCount = 1
		}
	}
	m.ChatList.Chats = append([]ChatItem{item}, m.ChatList.Chats...)
	m.selectChatByTarget(selectedTarget)
}

func fallbackChatTitle(title string, target string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	target = strings.TrimSpace(strings.TrimPrefix(target, "@"))
	if target == "" {
		return "chat"
	}
	return target
}

func normalizeChatTarget(target string) string {
	target = strings.TrimSpace(strings.ToLower(target))
	return strings.TrimPrefix(target, "@")
}

func (m *Model) findChatIndexByTarget(target string) int {
	needle := normalizeChatTarget(target)
	for i, chat := range m.ChatList.Chats {
		if normalizeChatTarget(chat.Target) == needle {
			return i
		}
	}
	return -1
}

func (m *Model) selectChatByTarget(target string) bool {
	needle := normalizeChatTarget(target)
	if needle == "" {
		return false
	}
	filtered, _ := m.ChatList.filteredChats()
	for i, chat := range filtered {
		if normalizeChatTarget(chat.Target) == needle {
			m.ChatList.SelectedIdx = i
			return true
		}
	}
	return false
}

func (m Model) selectedChatTarget() string {
	chat, ok := m.selectedChat()
	if !ok {
		return ""
	}
	return chat.Target
}

func (m Model) currentTarget() string {
	chat, ok := m.selectedChat()
	if !ok {
		return ""
	}
	if chat.Target != "" {
		return chat.Target
	}
	fallback := strings.ToLower(strings.ReplaceAll(chat.Title, " ", ""))
	if fallback == "" {
		return ""
	}
	return "@" + fallback
}

func (m Model) selectedChatTitle() string {
	chat, ok := m.selectedChat()
	if !ok {
		return "chat"
	}
	return chat.Title
}

func (m Model) selectedChat() (ChatItem, bool) {
	filtered, _ := m.ChatList.filteredChats()
	if len(filtered) == 0 {
		return ChatItem{}, false
	}
	idx := m.ChatList.SelectedIdx
	if idx < 0 || idx >= len(filtered) {
		idx = 0
	}
	return filtered[idx], true
}

func (m Model) focusLabel() string {
	switch m.Focus {
	case focusChatList:
		return "chats"
	case focusMessages:
		return "messages"
	default:
		return "input"
	}
}
