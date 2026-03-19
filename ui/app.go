package ui

import (
	"context"
	"fmt"

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

	if ctx == nil {
		ctx = context.Background()
	}

	m := Model{
		Styles:          DefaultStyles(),
		Header:          header,
		ChatList:        NewChatList(),
		Messages:        NewMessageView(),
		Input:           NewInputArea(),
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
	cmds := []tea.Cmd{m.Input.InitCmd()}
	if m.backend == nil {
		m.Header.Status = StatusDisconnected
		m.lastError = "backend unavailable; running in stub mode"
	} else {
		cmds = append(cmds, m.loadAuthCmd())
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
		return m, tea.Batch(m.maybeStartAnimationCmd(), m.loadDialogsCmd())
	case backendDialogsMsg:
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
			return m, nil
		}
		m.lastError = ""
		return m, m.applyDialogs(msg.Dialogs)
	case backendMessagesMsg:
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
			return m, nil
		}
		m.lastError = ""
		m.Messages.SetMessages(mapBackendMessages(msg.Messages, msg.ChatTitle, msg.Target))
		m.markSelectedChatRead()
		return m, nil
	case backendSendMsg:
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
			return m, nil
		}
		m.lastError = ""
		m.appendOutgoingMessage(msg.Text)
		return m, nil
	case backendResolveChatMsg:
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
			return m, nil
		}
		m.lastError = ""
		if msg.Chat == nil {
			m.lastNotice = "No chat resolved"
			return m, nil
		}
		return m, m.openChat(ChatItem{Title: msg.Chat.Title, Target: msg.Chat.Target})
	case IncomingMessageMsg:
		m.handleIncomingMessage(msg)
		return m, nil
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg)
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	switch m.Focus {
	case focusMessages:
		m.Messages = m.Messages.Update(msg)
	case focusInput:
		m.Input = m.Input.Update(msg)
		return m, tea.Batch(m.consumeSentMessagesCmd(), m.maybeStartAnimationCmd())
	}
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "ctrl+f":
		m.Focus = focusChatList
		m.ChatList.SearchMode = true
		m.lastNotice = "Search chats"
		return m, nil
	case "ctrl+n":
		return m.openNewChatModal()
	case "ctrl+,":
		return m.openSettingsModal()
	case "r":
		m.Focus = focusInput
		m.Input.ReplyTo = m.selectedChatTitle()
		m.lastNotice = "Replying to " + m.Input.ReplyTo
		return m, nil
	case "d":
		return m.handleDeleteShortcut()
	case "enter":
		if m.Focus == focusChatList {
			m.Focus = focusInput
			m.markSelectedChatRead()
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
		before := m.selectedChatTarget()
		m.ChatList = m.ChatList.Update(msg)
		after := m.selectedChatTarget()
		if after != before {
			m.markSelectedChatRead()
			m.syncHeaderChat()
			return m, m.loadMessagesCmd()
		}
	case focusMessages:
		m.Messages = m.Messages.Update(msg)
	case focusInput:
		m.Input = m.Input.Update(msg)
		return m, tea.Batch(m.consumeSentMessagesCmd(), m.maybeStartAnimationCmd())
	}
	return m, nil
}

// View renders header, split panes, and input area.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return "Loading UI..."
	}

	header := m.Header
	header.CurrentChat = m.headerCurrentChat()
	headerView := header.View()
	inputView := m.Input.View()
	shortcutBar := m.renderShortcutBar()
	bodyHeight := m.bodyHeight()

	compact := m.isCompactLayout()
	sidebarWidth, _ := m.computePaneWidths()
	body := ""
	if compact {
		modeStyle := lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextSecondary).
			Padding(0, 1)
		modeLine := modeStyle.Width(widthWithinStyle(m.Width, modeStyle)).Render(m.compactModeLabel())
		panelHeight := bodyHeight - 1
		if panelHeight < 1 {
			panelHeight = 1
		}
		if m.Focus == focusChatList {
			bodyPanel := m.ChatList.View()
			body = lipgloss.JoinVertical(lipgloss.Left, modeLine, bodyPanel)
		} else {
			bodyPanel := m.Messages.View()
			body = lipgloss.JoinVertical(lipgloss.Left, modeLine, bodyPanel)
		}
	} else {
		chatPanel := m.Styles.SidebarBorder.
			Width(widthWithinStyle(sidebarWidth, m.Styles.SidebarBorder)).
			Height(bodyHeight).
			Render(m.ChatList.View())
		messagePanel := m.Messages.View()
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

func (m *Model) applyWindowSize(msg tea.WindowSizeMsg) {
	prevCompact := m.compactLayout
	m.Width = msg.Width
	m.Height = msg.Height
	m.Input = m.Input.Update(tea.WindowSizeMsg{Width: m.Width, Height: msg.Height})
	m.layoutComponents()
	if prevCompact != m.compactLayout {
		if m.compactLayout {
			m.lastNotice = "Compact layout enabled"
			if m.Focus == focusMessages {
				m.Focus = focusInput
			}
		} else {
			m.lastNotice = "Split layout enabled"
		}
	}

	headerSize := tea.WindowSizeMsg{Width: m.Width, Height: m.headerBlockHeight()}
	chatSize := tea.WindowSizeMsg{Width: m.ChatList.Width, Height: m.ChatList.Height}
	messageSize := tea.WindowSizeMsg{Width: m.Messages.Width, Height: m.Messages.Height}
	modalSize := tea.WindowSizeMsg{Width: m.Width, Height: m.Height}

	m.Header = m.Header.Update(headerSize)
	m.ChatList = m.ChatList.Update(chatSize)
	m.Messages = m.Messages.Update(messageSize)
	m.NewChatModal = m.NewChatModal.Update(modalSize)
	m.SettingsModal = m.SettingsModal.Update(modalSize)
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

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "esc" {
		m.ActiveModal = modalNone
		return m, nil
	}

	switch m.ActiveModal {
	case modalNewChat:
		if key == "enter" {
			if chat, ok := m.NewChatModal.SelectedChat(); ok {
				return m, m.openChat(chat)
			}
			query := m.NewChatModal.Query()
			if query == "" {
				m.lastNotice = "Enter a username or user id"
				return m, nil
			}
			if m.backend == nil {
				m.lastNotice = "No matching cached chat for " + query
				return m, nil
			}
			m.lastNotice = "Opening " + query
			return m, m.resolveNewChatCmd(query)
		}
		m.NewChatModal = m.NewChatModal.Update(msg)
	case modalSettings:
		m.SettingsModal = m.SettingsModal.Update(msg)
		m.lastNotice = fmt.Sprintf("Theme %s, Notifications %s", onOff(m.SettingsModal.ThemeDark), onOff(m.SettingsModal.Notifications))
	}

	return m, nil
}
