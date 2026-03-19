package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type activeModal int

const (
	modalNone activeModal = iota
	modalNewChat
	modalSettings
)

type NewChatModal struct {
	Search   textinput.Model
	Chats    []ChatItem
	Selected int
	Width    int
	Height   int
}

func NewNewChatModal() NewChatModal {
	search := textinput.New()
	search.Prompt = "Search: "
	search.Placeholder = "@username or user id"
	search.Focus()

	return NewChatModal{Search: search}
}

func (m NewChatModal) SetChats(chats []ChatItem) NewChatModal {
	m.Chats = append([]ChatItem(nil), chats...)
	if m.Selected >= len(m.filteredChats()) {
		m.Selected = len(m.filteredChats()) - 1
	}
	if m.Selected < 0 {
		m.Selected = 0
	}
	return m
}

func (m NewChatModal) Update(msg tea.Msg) NewChatModal {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.Selected > 0 {
				m.Selected--
			}
		case "down":
			if m.Selected < len(m.filteredChats())-1 {
				m.Selected++
			}
		}
	}
	var cmd tea.Cmd
	m.Search, cmd = m.Search.Update(msg)
	_ = cmd
	if m.Selected >= len(m.filteredChats()) {
		m.Selected = len(m.filteredChats()) - 1
	}
	if m.Selected < 0 {
		m.Selected = 0
	}
	return m
}

func (m NewChatModal) Query() string {
	return strings.TrimSpace(m.Search.Value())
}

func (m NewChatModal) SelectedChat() (ChatItem, bool) {
	chats := m.filteredChats()
	if len(chats) == 0 || m.Selected < 0 || m.Selected >= len(chats) {
		return ChatItem{}, false
	}
	return chats[m.Selected], true
}

func (m NewChatModal) View() string {
	chats := m.filteredChats()
	rows := []string{"Start New Chat", "", m.Search.View(), ""}

	if len(chats) == 0 {
		query := m.Query()
		switch {
		case query != "":
			rows = append(rows,
				lipgloss.NewStyle().Foreground(TelegramDark.AccentBlue).Render("Press Enter to open "+query),
				dimText("No cached chats match yet."),
			)
		default:
			rows = append(rows, dimText("Type a username or user id to open a chat."))
		}
	} else {
		for i, chat := range chats {
			label := chat.Title
			if strings.TrimSpace(label) == "" {
				label = strings.TrimSpace(chat.Target)
			}
			row := fmt.Sprintf("%s %s", label, dimText("("+chat.Target+")"))
			if chat.UnreadCount > 0 {
				row += " " + lipgloss.NewStyle().Foreground(TelegramDark.AccentBlue).Render(fmt.Sprintf("[%d unread]", chat.UnreadCount))
			}
			if i == m.Selected {
				row = lipgloss.NewStyle().Foreground(TelegramDark.AccentBlue).Render("› " + row)
			}
			rows = append(rows, row)
		}
	}
	rows = append(rows, "", "Enter to open • Esc to close")

	box := lipgloss.NewStyle().
		Background(TelegramDark.BgSecondary).
		Foreground(TelegramDark.TextPrimary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(TelegramDark.AccentBlue).
		Padding(1, 2).
		Width(m.modalWidth()).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box)
}

func (m NewChatModal) filteredChats() []ChatItem {
	q := strings.TrimSpace(strings.ToLower(m.Search.Value()))
	if q == "" {
		return append([]ChatItem(nil), m.Chats...)
	}
	filtered := make([]ChatItem, 0, len(m.Chats))
	for _, chat := range m.Chats {
		haystack := strings.ToLower(chat.Title + " " + chat.Target + " " + chat.LastMessage)
		if strings.Contains(haystack, q) {
			filtered = append(filtered, chat)
		}
	}
	return filtered
}

func (m NewChatModal) modalWidth() int {
	width := 48
	if m.Width > 0 && m.Width < width+6 {
		width = m.Width - 6
		if width < 26 {
			width = 26
		}
	}
	return width
}

func (m NewChatModal) Bounds() (left, top, width, height int) {
	width = m.modalWidth()
	contentRows := len(m.filteredChats()) + 6
	if contentRows < 7 {
		contentRows = 7
	}
	height = contentRows + 4
	if m.Width > 0 {
		left = (m.Width - width) / 2
	}
	if m.Height > 0 {
		top = (m.Height - height) / 2
	}
	return left, top, width, height
}

type SettingsModal struct {
	ThemeDark     bool
	Notifications bool
	Selected      int
	Width         int
	Height        int
}

func NewSettingsModal() SettingsModal {
	return SettingsModal{ThemeDark: true, Notifications: true}
}

func (m SettingsModal) Update(msg tea.Msg) SettingsModal {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.Selected > 0 {
				m.Selected--
			}
		case "down":
			if m.Selected < 1 {
				m.Selected++
			}
		case "enter", " ":
			m = m.toggleSelected()
		}
	}
	return m
}

func (m SettingsModal) View() string {
	rows := []string{"Settings", ""}
	options := []string{
		fmt.Sprintf("Theme: %s", onOff(m.ThemeDark)),
		fmt.Sprintf("Notifications: %s", onOff(m.Notifications)),
	}
	for i, opt := range options {
		if i == m.Selected {
			rows = append(rows, lipgloss.NewStyle().Foreground(TelegramDark.AccentBlue).Render("› "+opt))
		} else {
			rows = append(rows, "  "+opt)
		}
	}
	rows = append(rows, "", "Enter/Space to toggle • Esc to close")

	box := lipgloss.NewStyle().
		Background(TelegramDark.BgSecondary).
		Foreground(TelegramDark.TextPrimary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(TelegramDark.AccentBlue).
		Padding(1, 2).
		Width(m.modalWidth()).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box)
}

func (m SettingsModal) toggleSelected() SettingsModal {
	if m.Selected == 0 {
		m.ThemeDark = !m.ThemeDark
	} else {
		m.Notifications = !m.Notifications
	}
	return m
}

func (m SettingsModal) modalWidth() int {
	width := 44
	if m.Width > 0 && m.Width < width+6 {
		width = m.Width - 6
		if width < 24 {
			width = 24
		}
	}
	return width
}

func (m SettingsModal) Bounds() (left, top, width, height int) {
	width = m.modalWidth()
	contentRows := 6
	height = contentRows + 4
	if m.Width > 0 {
		left = (m.Width - width) / 2
	}
	if m.Height > 0 {
		top = (m.Height - height) / 2
	}
	return left, top, width, height
}

func onOff(v bool) string {
	if v {
		return "On"
	}
	return "Off"
}

func dimText(v string) string {
	return lipgloss.NewStyle().Foreground(TelegramDark.TextMuted).Render(v)
}
