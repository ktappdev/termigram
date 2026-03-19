package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) layoutComponents() {
	m.compactLayout = m.isCompactLayout()

	bodyHeight := m.bodyHeight()
	sidebarWidth, messageWidth := m.computePaneWidths()

	m.Header.Width = m.Width
	m.ChatList.Width = m.chatPanelContentWidth(sidebarWidth)
	m.ChatList.Height = bodyHeight
	m.Messages.Width = messageWidth
	m.Messages.Height = bodyHeight
	m.Input.Width = m.Width
}

func (m Model) needsAnimation() bool {
	if m.Input.Typing {
		return true
	}
	return m.Header.Status == StatusConnected || m.Header.Status == StatusConnecting
}

func (m *Model) maybeStartAnimationCmd() tea.Cmd {
	if m.animationActive || !m.needsAnimation() {
		return nil
	}
	m.animationActive = true
	return animationTickCmd()
}

func (m Model) isCompactLayout() bool {
	return m.Width > 0 && m.Width < compactWidthThreshold
}

func (m Model) statusLineHeight() int {
	if m.lastError != "" || m.lastNotice != "" {
		return 1
	}
	return 0
}

func (m Model) inputBlockHeight() int {
	height := m.Input.Input.Height() + m.Input.ContainerStyle.GetVerticalFrameSize() + 1
	if m.Input.ReplyTo != "" {
		height++
	}
	if m.Input.Typing {
		height++
	}
	if height < 4 {
		height = 4
	}
	return height
}

func (m Model) bodyHeight() int {
	bodyHeight := m.Height - m.headerBlockHeight() - m.inputBlockHeight() - 1 - m.statusLineHeight()
	if m.isCompactLayout() {
		bodyHeight--
	}
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	return bodyHeight
}

func (m Model) headerBlockHeight() int {
	header := m.Header
	header.Width = m.Width
	header.CurrentChat = m.headerCurrentChat()
	height := lipgloss.Height(header.View())
	if height < 2 {
		return 2
	}
	return height
}

func (m Model) computePaneWidths() (int, int) {
	if m.isCompactLayout() {
		return m.Width, m.Width
	}

	sidebarWidth := m.Width * 30 / 100
	if sidebarWidth < 24 {
		sidebarWidth = 24
	}
	if sidebarWidth > m.Width-20 {
		sidebarWidth = m.Width / 3
	}
	messageWidth := m.Width - sidebarWidth
	if messageWidth < 1 {
		messageWidth = 1
	}
	return sidebarWidth, messageWidth
}

func (m Model) chatPanelContentWidth(sidebarWidth int) int {
	if m.isCompactLayout() {
		return sidebarWidth
	}
	if sidebarWidth <= 1 {
		return 1
	}
	return sidebarWidth - 1
}

func (m Model) compactModeLabel() string {
	if m.Focus == focusChatList {
		return "Compact: Chats view (Enter to open, Tab to composer)"
	}
	return "Compact: Conversation view (Esc to chats)"
}

func (m Model) renderShortcutBar() string {
	text := "↑/↓ Move  Enter Open chat/send  Tab Focus  Esc Back  Ctrl+F Search  Ctrl+N New chat  Ctrl+, Settings  Ctrl+O Attach  R Reply  D Delete  ? Help"
	if m.isCompactLayout() {
		text = "Enter open/send • Esc chats • Ctrl+F search • Ctrl+N new chat • Ctrl+, settings • ? help"
	}
	bar := lipgloss.NewStyle().
		Background(TelegramDark.BgSecondary).
		Foreground(TelegramDark.TextSecondary).
		Padding(0, 1)
	return bar.Width(widthWithinStyle(m.Width, bar)).Render(text)
}

func (m Model) renderHelpOverlay(base string) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		"Keyboard Shortcuts",
		"",
		"Navigation",
		"  ↑ / ↓      Move in current list",
		"  Enter      Open selected chat / send message",
		"  Esc        Close help / return to chats",
		"  Tab        Cycle focus",
		"",
		"Actions",
		"  Ctrl+F     Search chats",
		"  Ctrl+N     Open new-chat selector",
		"  Ctrl+,     Open settings",
		"  Ctrl+O     Insert attach token",
		"  R          Reply to selected chat",
		"  D          Delete latest message/chat",
		"  ?          Toggle this help",
		"  Ctrl+C     Quit",
		"",
		"Mouse",
		"  Left click  Select chat or focus panel",
		"  Wheel       Scroll messages",
	)

	boxWidth := 56
	if m.Width > 0 && m.Width < 62 {
		boxWidth = m.Width - 4
		if boxWidth < 28 {
			boxWidth = 28
		}
	}

	box := lipgloss.NewStyle().
		Background(TelegramDark.BgPrimary).
		Foreground(TelegramDark.TextPrimary).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(TelegramDark.AccentBlue).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	overlay := lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box)
	if base == "" {
		return overlay
	}
	return overlay
}
