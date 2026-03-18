package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focusArea int

const (
	focusChatList focusArea = iota
	focusMessages
	focusInput
)

const compactWidthThreshold = 60

type backendAuthMsg struct {
	Username string
	Err      error
}

type backendMessagesMsg struct {
	Target    string
	ChatTitle string
	Messages  []BackendMessage
	Err       error
}

type backendDialogsMsg struct {
	Dialogs []BackendDialog
	Err     error
}

type backendSendMsg struct {
	Text string
	Err  error
}

// Model composes all UI components into the full TUI layout.
type Model struct {
	Width  int
	Height int
	Styles Styles

	Header   HeaderModel
	ChatList ChatListModel
	Messages MessageViewModel
	Input    InputAreaModel

	Focus           focusArea
	ShowHelp        bool
	ActiveModal     activeModal
	compactLayout   bool
	animationActive bool
	lastSentSeen    int
	lastError       string
	lastNotice      string

	NewChatModal  NewChatModal
	SettingsModal SettingsModal

	ctx     context.Context
	backend Backend
}

// NewModel creates the root app model with component stubs plus optional backend.
func NewModel(ctx context.Context, backend Backend) Model {
	header := NewHeader()
	header.Status = StatusConnecting
	header.Username = "guest"

	input := NewInputArea()
	input.ReplyTo = "Alice"

	if ctx == nil {
		ctx = context.Background()
	}

	m := Model{
		Styles:          DefaultStyles(),
		Header:          header,
		ChatList:        NewChatList(),
		Messages:        NewMessageView(),
		Input:           input,
		Focus:           focusInput,
		animationActive: false,
		NewChatModal:    NewNewChatModal(),
		SettingsModal:   NewSettingsModal(),
		ctx:             ctx,
		backend:         backend,
	}
	m.syncHeaderChat()
	return m
}

// Init starts the Bubble Tea program.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	if m.backend == nil {
		m.Header.Status = StatusDisconnected
		m.lastError = "backend unavailable; running in stub mode"
	} else {
		cmds = append(cmds, m.loadAuthCmd(), m.loadMessagesCmd())
	}
	if m.needsAnimation() {
		cmds = append(cmds, animationTickCmd())
	}
	return tea.Batch(cmds...)
}

// Update routes key and resize events to the focused component.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animationTickMsg:
		if !m.needsAnimation() {
			m.animationActive = false
			return m, nil
		}
		m.Header.PulseStep = (m.Header.PulseStep + 1) % 12
		if m.Input.Typing {
			m.Input.TypingStep = (m.Input.TypingStep + 1) % 4
		}
		m.animationActive = true
		return m, animationTickCmd()
	case backendAuthMsg:
		if msg.Err != nil {
			m.Header.Status = StatusDisconnected
			m.lastError = msg.Err.Error()
			return m, m.maybeStartAnimationCmd()
		}
		m.Header.Status = StatusConnected
		if msg.Username != "" {
			m.Header.Username = msg.Username
		}
		m.lastError = ""
		return m, m.maybeStartAnimationCmd()
	case backendMessagesMsg:
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
			return m, nil
		}
		m.lastError = ""
		m.Messages.Messages = mapBackendMessages(msg.Messages, msg.ChatTitle, msg.Target)
		m.Messages.Scroll = m.Messages.maxScroll()
		return m, nil
	case backendSendMsg:
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
			return m, nil
		}
		m.lastError = ""
		m.Messages.Messages = append(m.Messages.Messages, Message{
			Text:     msg.Text,
			Time:     time.Now().Format("15:04"),
			Sender:   "You",
			Chat:     chatLabel(m.selectedChatTitle(), m.currentTarget(), ""),
			Outgoing: true,
			Read:     false,
		})
		m.Messages.Scroll = m.Messages.maxScroll()
		return m, nil
	case tea.WindowSizeMsg:
		prevCompact := m.compactLayout
		m.Width = msg.Width
		m.Height = msg.Height
		m.layoutComponents()
		if prevCompact != m.compactLayout {
			if m.compactLayout {
				m.lastNotice = "Compact layout enabled"
				if m.Focus == focusMessages || m.Focus == focusInput {
					m.Focus = focusMessages
				}
			} else {
				m.lastNotice = "Split layout enabled"
			}
		}

		headerSize := tea.WindowSizeMsg{Width: m.Width, Height: 1}
		chatSize := tea.WindowSizeMsg{Width: m.ChatList.Width, Height: m.ChatList.Height}
		messageSize := tea.WindowSizeMsg{Width: m.Messages.Width, Height: m.Messages.Height}
		inputSize := tea.WindowSizeMsg{Width: m.Width, Height: m.inputBlockHeight()}
		modalSize := tea.WindowSizeMsg{Width: m.Width, Height: m.Height}

		m.Header = m.Header.Update(headerSize)
		m.ChatList = m.ChatList.Update(chatSize)
		m.Messages = m.Messages.Update(messageSize)
		m.Input = m.Input.Update(inputSize)
		m.NewChatModal = m.NewChatModal.Update(modalSize)
		m.SettingsModal = m.SettingsModal.Update(modalSize)
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		key := msg.String()
		if m.ActiveModal != modalNone {
			return m.handleModalKey(msg)
		}
		switch key {
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			m.ShowHelp = !m.ShowHelp
			return m, nil
		}

		if m.ShowHelp {
			if key == "esc" {
				m.ShowHelp = false
			}
			return m, nil
		}

		switch key {
		case "tab":
			m.Focus = (m.Focus + 1) % 3
			m.lastNotice = "Focus: " + m.focusLabel()
			return m, nil
		case "shift+tab":
			m.Focus = (m.Focus + 2) % 3
			m.lastNotice = "Focus: " + m.focusLabel()
			return m, nil
		case "ctrl+n":
			m.ActiveModal = modalNewChat
			m.NewChatModal.Search.SetValue("")
			m.NewChatModal.Selected = 0
			m.lastNotice = "New chat"
			return m, nil
		case "ctrl+,":
			m.ActiveModal = modalSettings
			m.lastNotice = "Settings"
			return m, nil
		case "ctrl+f":
			m.Focus = focusChatList
			m.ChatList.SearchMode = true
			m.lastNotice = "Search chats"
			return m, nil
		case "r":
			m.Focus = focusInput
			m.Input.ReplyTo = m.selectedChatTitle()
			m.lastNotice = "Replying to " + m.Input.ReplyTo
			return m, nil
		case "d":
			return m.handleDeleteShortcut()
		case "enter":
			if m.Focus == focusChatList {
				m.Focus = focusMessages
				m.syncHeaderChat()
				m.lastNotice = "Opened " + m.selectedChatTitle()
				return m, m.loadMessagesCmd()
			}
		case "esc":
			if m.Focus != focusInput {
				m.Focus = focusChatList
				m.ChatList.SearchMode = false
				m.lastNotice = "Focus: chats"
				return m, nil
			}
		}

		switch m.Focus {
		case focusChatList:
			before := m.ChatList.SelectedIdx
			m.ChatList = m.ChatList.Update(msg)
			if m.ChatList.SelectedIdx != before {
				m.syncHeaderChat()
				return m, m.loadMessagesCmd()
			}
		case focusMessages:
			m.Messages = m.Messages.Update(msg)
		case focusInput:
			m.Input = m.Input.Update(msg)
			sendCmd := m.consumeSentMessagesCmd()
			return m, tea.Batch(sendCmd, m.maybeStartAnimationCmd())
		}
	}

	return m, nil
}

// View renders header, split panes, and input area.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return "Loading UI..."
	}

	header := m.Header
	header.CurrentChat = m.selectedChatTitle()
	headerView := header.View()
	inputView := m.Input.View()
	shortcutBar := m.renderShortcutBar()

	bodyHeight := m.bodyHeight()

	compact := m.isCompactLayout()
	sidebarWidth, messageWidth := m.computePaneWidths()
	body := ""
	if compact {
		modeLine := lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextSecondary).
			Padding(0, 1).
			Width(m.Width).
			Render(m.compactModeLabel())
		panelHeight := bodyHeight - 1
		if panelHeight < 1 {
			panelHeight = 1
		}
		if m.Focus == focusChatList {
			bodyPanel := m.ChatList.BaseStyle.Width(m.Width).Height(panelHeight).Render(m.ChatList.View())
			body = lipgloss.JoinVertical(lipgloss.Left, modeLine, bodyPanel)
		} else {
			bodyPanel := m.Messages.BaseStyle.Width(m.Width).Height(panelHeight).Render(m.Messages.View())
			body = lipgloss.JoinVertical(lipgloss.Left, modeLine, bodyPanel)
		}
	} else {
		chatPanel := m.ChatList.BaseStyle.Width(sidebarWidth).Height(bodyHeight).Render(m.ChatList.View())
		messagePanel := m.Messages.BaseStyle.Width(messageWidth).Height(bodyHeight).Render(m.Messages.View())
		body = lipgloss.JoinHorizontal(lipgloss.Top, chatPanel, messagePanel)
	}

	lines := []string{headerView, body, shortcutBar}
	if m.lastError != "" {
		status := lipgloss.NewStyle().
			Foreground(TelegramDark.AccentRed).
			MaxWidth(m.Width).
			Width(m.Width)
		lines = append(lines, status.Render("Error: "+m.lastError))
	} else if m.lastNotice != "" {
		status := lipgloss.NewStyle().
			Foreground(TelegramDark.TextMuted).
			MaxWidth(m.Width).
			Width(m.Width)
		lines = append(lines, status.Render(m.lastNotice))
	}
	lines = append(lines, inputView)
	base := lipgloss.JoinVertical(lipgloss.Left, lines...)

	if m.ActiveModal != modalNone {
		return m.renderModalOverlay(base)
	}
	if m.ShowHelp {
		return m.renderHelpOverlay(base)
	}
	return base
}

func (m *Model) layoutComponents() {
	m.compactLayout = m.isCompactLayout()

	bodyHeight := m.bodyHeight()
	sidebarWidth, messageWidth := m.computePaneWidths()

	m.Header.Width = m.Width
	m.ChatList.Width = sidebarWidth
	m.ChatList.Height = bodyHeight
	m.Messages.Width = messageWidth
	m.Messages.Height = bodyHeight
	m.Input.Width = m.Width
}

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
		text := m.Input.Sent[m.lastSentSeen]
		m.lastSentSeen++
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
	return tea.Batch(cmds...)
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
		if len(m.Messages.Messages) == 0 {
			m.lastNotice = "No message to delete"
			return m, nil
		}
		m.Messages.Messages = m.Messages.Messages[:len(m.Messages.Messages)-1]
		m.Messages.Scroll = m.Messages.maxScroll()
		m.lastNotice = "Deleted latest message"
		return m, nil
	default:
		m.lastNotice = "Delete works in chats/messages panels"
		return m, nil
	}
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

func (m Model) bodyHeight() int {
	bodyHeight := m.Height - 1 - m.inputBlockHeight() - 1 - m.statusLineHeight()
	if m.isCompactLayout() {
		bodyHeight-- // compact mode label row
	}
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	return bodyHeight
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

func (m Model) compactModeLabel() string {
	if m.Focus == focusChatList {
		return "Compact: Chats view (Enter/Tab to open messages)"
	}
	return "Compact: Messages view (Esc/Tab to show chats)"
}

func (m Model) renderShortcutBar() string {
	text := "↑/↓ Move  Enter Open chat  Esc Back  Ctrl+F Search  Ctrl+N New chat  Ctrl+, Settings  Ctrl+O Attach  R Reply  D Delete  ? Help"
	if m.isCompactLayout() {
		text = "Enter open • Esc chats • Ctrl+F search • ? help"
	}
	bar := lipgloss.NewStyle().
		Background(TelegramDark.BgSecondary).
		Foreground(TelegramDark.TextSecondary).
		Padding(0, 1)
	return bar.Width(m.Width).Render(text)
}

func (m Model) renderHelpOverlay(base string) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		"Keyboard Shortcuts",
		"",
		"Navigation",
		"  ↑ / ↓      Move in current list",
		"  Enter      Open selected chat",
		"  Esc        Close help / return to chats",
		"  Tab        Next focus panel",
		"  Shift+Tab  Previous focus panel",
		"",
		"Actions",
		"  Ctrl+N     Open new chat modal",
		"  Ctrl+,     Open settings modal",
		"  Ctrl+F     Search chats",
		"  Ctrl+O     Insert attach token in input",
		"  R          Reply to selected chat",
		"  D          Delete (chat/message)",
		"  ?          Toggle this help",
		"  Ctrl+C     Quit",
		"",
		"Mouse",
		"  Left click  Select chat / focus panel / click Send",
		"  Wheel       Scroll messages when hovering message pane",
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

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "esc" {
		m.ActiveModal = modalNone
		return m, nil
	}

	switch m.ActiveModal {
	case modalNewChat:
		m.NewChatModal = m.NewChatModal.Update(msg)
		if key == "enter" {
			user := m.NewChatModal.SelectedUser()
			if user != "" && user != "No matches" {
				title := strings.TrimPrefix(user, "@")
				if title == "" {
					title = "new chat"
				}
				m.ChatList.Chats = append([]ChatItem{{
					Title:       title,
					Target:      user,
					LastMessage: "",
					LastTime:    "now",
					Online:      true,
				}}, m.ChatList.Chats...)
				m.ChatList.SelectedIdx = 0
				m.ActiveModal = modalNone
				m.Focus = focusMessages
				m.syncHeaderChat()
				m.lastNotice = "Started chat with " + user
				return m, m.loadMessagesCmd()
			}
		}
	case modalSettings:
		m.SettingsModal = m.SettingsModal.Update(msg)
		m.lastNotice = "Theme " + onOff(m.SettingsModal.ThemeDark) + ", Notifications " + onOff(m.SettingsModal.Notifications)
	}

	return m, nil
}

func (m Model) renderModalOverlay(base string) string {
	switch m.ActiveModal {
	case modalNewChat:
		return m.NewChatModal.View()
	case modalSettings:
		return m.SettingsModal.View()
	default:
		return base
	}
}
