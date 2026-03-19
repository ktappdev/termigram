package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	homeTitle    = "Termigram"
	homeSubtitle = "Telegram for nerds and AI agents"
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
		Title:       homeTitle,
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
	rightFull := fmt.Sprintf("%s %s | @%s ", h.statusBadge(), h.Status, h.Username)
	rightCompact := fmt.Sprintf("%s %s ", h.statusBadge(), h.Status)
	if h.Username == "" {
		rightFull = rightCompact
	}

	if strings.TrimSpace(h.CurrentChat) == "" {
		return h.renderHomeHeader(rightFull, rightCompact)
	}
	return h.renderChatHeader(rightFull, rightCompact)
}

func (h HeaderModel) renderHomeHeader(rightFull string, rightCompact string) string {
	title := lipgloss.NewStyle().
		Foreground(TelegramDark.AccentBlue).
		Bold(true).
		Render(homeTitle)
	subtitleStyle := lipgloss.NewStyle().Foreground(TelegramDark.TextSecondary)

	if h.Width > 0 && lipgloss.Width(rightCompact) >= h.Width {
		return h.Style.Width(h.Width).Render(lipgloss.NewStyle().MaxWidth(h.Width).Render(rightCompact))
	}

	right := rightFull
	if h.Width > 0 && lipgloss.Width(rightFull) > h.Width/2 {
		right = rightCompact
	}

	leftAvailable := h.Width - lipgloss.Width(right)
	if leftAvailable <= 0 {
		leftAvailable = lipgloss.Width(title)
	}

	left := h.homeBrandBlock(title, subtitleStyle, leftAvailable)
	return h.renderAlignedLine(left, right)
}

func (h HeaderModel) renderChatHeader(rightFull string, rightCompact string) string {
	left := fmt.Sprintf(" %s  💬 %s", h.Title, h.CurrentChat)
	if h.Width > 0 {
		if lipgloss.Width(rightCompact) >= h.Width {
			return h.Style.Width(h.Width).Render(lipgloss.NewStyle().MaxWidth(h.Width).Render(rightCompact))
		}

		right := rightFull
		if lipgloss.Width(right) > h.Width/2 {
			right = rightCompact
		}

		maxLeft := h.Width - lipgloss.Width(right) - 1
		if maxLeft < 1 {
			maxLeft = 1
		}
		left = truncateWithEllipsis(left, maxLeft)
		return h.renderAlignedLine(left, right)
	}

	return h.Style.Width(h.Width).Render(left + rightFull)
}

func (h HeaderModel) homeBrandBlock(title string, subtitleStyle lipgloss.Style, available int) string {
	compactTitle := title
	if available <= lipgloss.Width(title)+2 {
		return compactTitle
	}

	subtitle := " · " + homeSubtitle
	if available < lipgloss.Width(title)+lipgloss.Width(subtitle)+1 {
		subtitle = " · " + truncateWithEllipsis(homeSubtitle, available-lipgloss.Width(title)-3)
	}
	if strings.TrimSpace(strings.TrimPrefix(subtitle, " · ")) == "" {
		return compactTitle
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		title,
		subtitleStyle.Render(subtitle),
	)
}

func (h HeaderModel) renderAlignedLine(left string, right string) string {
	content := left + right
	if h.Width > 0 {
		gap := h.Width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap > 0 {
			content = left + strings.Repeat(" ", gap) + right
		} else if lipgloss.Width(left) > h.Width-lipgloss.Width(right) {
			content = truncateWithEllipsis(left, h.Width-lipgloss.Width(right)-1) + " " + right
		}
	}
	return h.Style.Width(h.Width).Render(content)
}

func truncateWithEllipsis(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= maxWidth {
		return text
	}
	if maxWidth == 1 {
		return "…"
	}

	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes) + "…"
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return "…"
}

func (h HeaderModel) statusBadge() string {
	return lipgloss.NewStyle().
		Foreground(connectionPulseColor(h.Status, h.PulseStep)).
		Render("●")
}
