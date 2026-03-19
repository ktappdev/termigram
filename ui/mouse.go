package ui

import tea "github.com/charmbracelet/bubbletea"

type layoutMetrics struct {
	compact      bool
	sidebarWidth int
	messageWidth int
	bodyY        int
	bodyHeight   int
	inputY       int
	inputHeight  int
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.ActiveModal != modalNone {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress && !m.pointInActiveModal(msg.X, msg.Y) {
			m.ActiveModal = modalNone
		}
		return m, nil
	}

	if m.ShowHelp {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			m.ShowHelp = false
		}
		return m, nil
	}

	layout := m.layoutMetrics()

	if msg.Action == tea.MouseActionMotion {
		if m.pointInChatPanel(msg.X, msg.Y, layout) {
			m.ChatList.HoveredIdx = m.chatIndexAt(msg.Y, layout)
		} else {
			m.ChatList.HoveredIdx = -1
		}
		return m, nil
	}

	if m.pointInMessagePanel(msg.X, msg.Y, layout) {
		switch msg.Button {
		case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
			m.Focus = focusMessages
			m.Messages = m.Messages.Update(msg)
			return m, nil
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionPress {
				m.Focus = focusMessages
				m.lastNotice = "Focus: messages"
				return m, nil
			}
		}
	}

	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	if m.pointInChatPanel(msg.X, msg.Y, layout) {
		m.Focus = focusChatList
		m.lastNotice = "Focus: chats"
		idx := m.chatIndexAt(msg.Y, layout)
		filtered, _ := m.ChatList.filteredChats()
		if idx >= 0 && idx < len(filtered) {
			if m.ChatList.SelectedIdx != idx {
				m.ChatList.SelectedIdx = idx
				m.markSelectedChatRead()
				return m, m.loadMessagesCmd()
			}
		}
		return m, nil
	}

	if m.pointInInputPanel(msg.Y, layout) {
		m.Focus = focusInput
		m.lastNotice = "Focus: input"
		return m, nil
	}

	return m, nil
}

func (m Model) layoutMetrics() layoutMetrics {
	compact := m.isCompactLayout()
	sidebarWidth, messageWidth := m.computePaneWidths()
	bodyHeight := m.bodyHeight()
	inputHeight := m.inputBlockHeight()
	inputY := m.headerBlockHeight() + bodyHeight + 1 + m.statusLineHeight()
	if compact {
		inputY++
	}

	return layoutMetrics{
		compact:      compact,
		sidebarWidth: sidebarWidth,
		messageWidth: messageWidth,
		bodyY:        m.headerBlockHeight(),
		bodyHeight:   bodyHeight,
		inputY:       inputY,
		inputHeight:  inputHeight,
	}
}

func (m Model) pointInChatPanel(x, y int, layout layoutMetrics) bool {
	if layout.compact {
		return m.Focus == focusChatList && x >= 0 && x < m.Width && y >= layout.bodyY+1 && y < layout.bodyY+1+layout.bodyHeight
	}
	return x >= 0 && x < layout.sidebarWidth && y >= layout.bodyY && y < layout.bodyY+layout.bodyHeight
}

func (m Model) pointInMessagePanel(x, y int, layout layoutMetrics) bool {
	if layout.compact {
		return m.Focus != focusChatList && x >= 0 && x < m.Width && y >= layout.bodyY+1 && y < layout.bodyY+1+layout.bodyHeight
	}
	left := layout.sidebarWidth
	right := layout.sidebarWidth + layout.messageWidth
	return x >= left && x < right && y >= layout.bodyY && y < layout.bodyY+layout.bodyHeight
}

func (m Model) pointInInputPanel(y int, layout layoutMetrics) bool {
	return y >= layout.inputY && y < layout.inputY+layout.inputHeight
}

func (m Model) chatIndexAt(y int, layout layoutMetrics) int {
	relY := y - layout.bodyY
	if layout.compact {
		relY--
	}
	if relY < m.ChatList.searchBlockSize {
		return -1
	}

	relY -= m.ChatList.searchBlockSize
	idx := relY / m.ChatList.itemBlockSize
	if relY%m.ChatList.itemBlockSize == m.ChatList.itemBlockSize-1 {
		return -1
	}

	filtered, _ := m.ChatList.filteredChats()
	if idx < 0 || idx >= len(filtered) {
		return -1
	}
	return idx
}

func (m Model) pointInActiveModal(x, y int) bool {
	var left, top, width, height int
	switch m.ActiveModal {
	case modalNewChat:
		left, top, width, height = m.NewChatModal.Bounds()
	case modalSettings:
		left, top, width, height = m.SettingsModal.Bounds()
	default:
		return false
	}
	return x >= left && x < left+width && y >= top && y < top+height
}
