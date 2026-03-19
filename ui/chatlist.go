package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChatItem is a minimal conversation row shown in the chat sidebar.
type ChatItem struct {
	Title       string
	Target      string
	LastMessage string
	LastTime    string
	Online      bool
	UnreadCount int
}

// ChatListModel renders the left sidebar with search and chat rows.
type ChatListModel struct {
	Chats       []ChatItem
	SelectedIdx int
	HoveredIdx  int
	SearchMode  bool
	SearchQuery string
	Width       int
	Height      int

	BaseStyle       lipgloss.Style
	ItemStyle       lipgloss.Style
	HoverStyle      lipgloss.Style
	SelectedStyle   lipgloss.Style
	SeparatorStyle  lipgloss.Style
	searchBlockSize int
	itemBlockSize   int
}

// NewChatList creates a minimal sidebar model with theme-aligned styles.
func NewChatList() ChatListModel {
	return ChatListModel{
		Chats: []ChatItem{
			{Title: "Alice", Target: "@alice", LastMessage: "Hey, are you free?", LastTime: "10:42", Online: true, UnreadCount: 2},
			{Title: "Bob", Target: "@bob", LastMessage: "Meeting moved to 3pm", LastTime: "09:10", Online: false, UnreadCount: 0},
			{Title: "Team Group", Target: "@teamgroup", LastMessage: "Release checklist updated", LastTime: "Yesterday", Online: true, UnreadCount: 7},
		},
		HoveredIdx: -1,
		BaseStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary),
		ItemStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceSm, spaceSm),
		HoverStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgHover).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceSm, spaceSm),
		SelectedStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgSelected).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceSm, spaceSm),
		SeparatorStyle: lipgloss.NewStyle().
			Foreground(TelegramDark.BgHover).
			Padding(spaceXs, spaceSm),
		searchBlockSize: 2,
		itemBlockSize:   5,
	}
}

// Update handles keyboard navigation and search toggling.
func (m ChatListModel) Update(msg tea.Msg) ChatListModel {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			m.SearchMode = true
		case "esc":
			m.SearchMode = false
		case "up":
			if m.SelectedIdx > 0 {
				m.SelectedIdx--
			}
		case "down":
			chats, _ := m.filteredChats()
			if m.SelectedIdx < len(chats)-1 {
				m.SelectedIdx++
			}
		case "backspace":
			if m.SearchMode && len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
			}
		default:
			if m.SearchMode && len(msg.Runes) > 0 {
				m.SearchQuery += string(msg.Runes)
			}
		}
	}

	chats, _ := m.filteredChats()
	if len(chats) == 0 {
		m.SelectedIdx = 0
		m.HoveredIdx = -1
		return m
	}
	if m.SelectedIdx >= len(chats) {
		m.SelectedIdx = len(chats) - 1
	}
	if m.HoveredIdx >= len(chats) {
		m.HoveredIdx = -1
	}
	return m
}

// View renders the search row and chat list.
func (m ChatListModel) View() string {
	rows := []string{m.renderSearch()}
	chats, _ := m.filteredChats()
	for i, chat := range chats {
		rows = append(rows, m.renderItem(chat, i == m.SelectedIdx, i == m.HoveredIdx))
		if i < len(chats)-1 {
			rows = append(rows, m.renderSeparator())
		}
	}

	if len(chats) == 0 {
		empty := lipgloss.NewStyle().Foreground(TelegramDark.TextMuted).Padding(spaceXs, spaceSm).Render("No chats found")
		rows = append(rows, empty)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return m.BaseStyle.Width(m.Width).Height(m.Height).Render(content)
}

func (m ChatListModel) renderSearch() string {
	query := "Search"
	if m.SearchQuery != "" {
		query = m.SearchQuery
	}
	prefix := "🔍 "
	if m.SearchMode {
		prefix = "🔎 "
	}

	searchStyle := lipgloss.NewStyle().
		Background(TelegramDark.BgPrimary).
		Foreground(TelegramDark.TextMuted).
		Padding(spaceXs, spaceSm).
		MarginBottom(spaceSm)

	if m.SearchQuery != "" {
		searchStyle = searchStyle.Foreground(TelegramDark.TextPrimary)
	}

	return searchStyle.Width(m.Width).Render(prefix + query)
}

func (m ChatListModel) renderSeparator() string {
	if m.Width <= 2 {
		return ""
	}
	lineWidth := m.Width - 2
	if lineWidth < 1 {
		lineWidth = 1
	}
	line := strings.Repeat("─", lineWidth)
	return m.SeparatorStyle.Width(m.Width).Render(line)
}

func (m ChatListModel) filteredChats() ([]ChatItem, []int) {
	query := strings.TrimSpace(strings.ToLower(m.SearchQuery))
	if query == "" {
		indices := make([]int, len(m.Chats))
		for i := range m.Chats {
			indices[i] = i
		}
		return m.Chats, indices
	}

	filtered := make([]ChatItem, 0, len(m.Chats))
	indices := make([]int, 0, len(m.Chats))
	for i, chat := range m.Chats {
		haystack := strings.ToLower(chat.Title + " " + chat.LastMessage + " " + chat.Target)
		if strings.Contains(haystack, query) {
			filtered = append(filtered, chat)
			indices = append(indices, i)
		}
	}
	return filtered, indices
}

func (m ChatListModel) renderItem(chat ChatItem, selected bool, hovered bool) string {
	style := m.ItemStyle
	if selected {
		style = m.SelectedStyle
	} else if hovered {
		style = m.HoverStyle
	}

	statusDot := lipgloss.NewStyle().Foreground(TelegramDark.TextMuted).Render("○")
	if chat.Online {
		statusDot = lipgloss.NewStyle().Foreground(TelegramDark.AccentGreen).Render("●")
	}

	header := fmt.Sprintf("%s %s", statusDot, chat.Title)
	meta := lipgloss.NewStyle().Foreground(TelegramDark.TextSecondary).Render(chat.LastTime)
	line1 := header
	if m.Width > 0 {
		gap := m.Width - 6 - lipgloss.Width(header) - lipgloss.Width(meta)
		if gap > 1 {
			line1 = header + strings.Repeat(" ", gap) + meta
		}
	}

	preview := lipgloss.NewStyle().Foreground(TelegramDark.TextSecondary).Render(chat.LastMessage)
	if chat.UnreadCount > 0 {
		badge := lipgloss.NewStyle().
			Background(TelegramDark.AccentBlue).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceXs, spaceSm).
			Render(fmt.Sprintf("%d", chat.UnreadCount))
		preview = fmt.Sprintf("%s %s", preview, badge)
	}

	return style.Width(m.Width).Render(lipgloss.JoinVertical(lipgloss.Left, line1, preview))
}
