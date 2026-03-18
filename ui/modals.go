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
	Users    []string
	Selected int
	Width    int
	Height   int
}

func NewNewChatModal() NewChatModal {
	search := textinput.New()
	search.Prompt = "Search: "
	search.Placeholder = "username"
	search.Focus()

	return NewChatModal{
		Search: search,
		Users:  []string{"@alice", "@bob", "@teamgroup", "@support", "@telegram"},
	}
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
			if m.Selected < len(m.filteredUsers())-1 {
				m.Selected++
			}
		}
	}
	var cmd tea.Cmd
	m.Search, cmd = m.Search.Update(msg)
	_ = cmd
	if m.Selected >= len(m.filteredUsers()) {
		m.Selected = len(m.filteredUsers()) - 1
	}
	if m.Selected < 0 {
		m.Selected = 0
	}
	return m
}

func (m NewChatModal) SelectedUser() string {
	users := m.filteredUsers()
	if len(users) == 0 || m.Selected < 0 || m.Selected >= len(users) {
		return ""
	}
	return users[m.Selected]
}

func (m NewChatModal) View() string {
	users := m.filteredUsers()
	if len(users) == 0 {
		users = []string{"No matches"}
	}

	rows := []string{"Start New Chat", "", m.Search.View(), ""}
	for i, user := range users {
		row := user
		if i == m.Selected {
			row = lipgloss.NewStyle().Foreground(TelegramDark.AccentBlue).Render("› " + user)
		}
		rows = append(rows, row)
	}
	rows = append(rows, "", "Enter to start • Esc to close")

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

func (m NewChatModal) filteredUsers() []string {
	q := strings.TrimSpace(strings.ToLower(m.Search.Value()))
	if q == "" {
		return m.Users
	}
	filtered := make([]string, 0, len(m.Users))
	for _, u := range m.Users {
		if strings.Contains(strings.ToLower(u), q) {
			filtered = append(filtered, u)
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
	users := m.filteredUsers()
	if len(users) == 0 {
		users = []string{"No matches"}
	}
	contentRows := len(users) + 6
	height = contentRows + 4 // padding + border
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
	height = contentRows + 4 // padding + border
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
