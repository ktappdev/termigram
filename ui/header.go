package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionStatus represents the connection state shown in the header.
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusDisconnected ConnectionStatus = "disconnected"
)

// HeaderModel renders the top status bar for the app.
type HeaderModel struct {
	Title       string
	CurrentChat string
	Username    string
	Status      ConnectionStatus
	PulseStep   int
	Width       int
	Style       lipgloss.Style
}

// NewHeader creates a minimal header with Telegram-like defaults.
func NewHeader() HeaderModel {
	return HeaderModel{
		Title:       "termigram",
		CurrentChat: "",
		Username:    "user",
		Status:      StatusDisconnected,
		PulseStep:   0,
		Style: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceXs, spaceSm).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(TelegramDark.BgHover),
	}
}

// Update updates header dimensions from Bubble Tea messages.
func (h HeaderModel) Update(msg tea.Msg) HeaderModel {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.Width = msg.Width
	}
	return h
}

// View renders the header row and adapts to available width.
func (h HeaderModel) View() string {
	left := fmt.Sprintf(" %s", h.Title)
	if h.CurrentChat != "" && h.Title == "termigram" {
		left = fmt.Sprintf(" %s  💬 %s", h.Title, h.CurrentChat)
	}
	rightFull := fmt.Sprintf("%s %s | @%s ", h.statusBadge(), h.Status, h.Username)
	rightCompact := fmt.Sprintf("%s %s ", h.statusBadge(), h.Status)

	content := left + rightFull
	if h.Width > 0 {
		if lipgloss.Width(rightCompact) >= h.Width {
			return h.Style.Width(h.Width).Render(lipgloss.NewStyle().MaxWidth(h.Width).Render(rightCompact))
		}

		right := rightFull
		if lipgloss.Width(right) > h.Width/2 {
			right = rightCompact
		}

		gap := h.Width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap > 0 {
			content = left + strings.Repeat(" ", gap) + right
		} else {
			content = right
		}
	}

	return h.Style.Width(h.Width).Render(content)
}

func (h HeaderModel) statusBadge() string {
	return lipgloss.NewStyle().
		Foreground(connectionPulseColor(h.Status, h.PulseStep)).
		Render("●")
}
