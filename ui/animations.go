package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const animationInterval = 120 * time.Millisecond

type animationTickMsg time.Time

func animationTickCmd() tea.Cmd {
	return tea.Tick(animationInterval, func(t time.Time) tea.Msg {
		return animationTickMsg(t)
	})
}

func typingDotsFrame(step int) string {
	switch step % 4 {
	case 1:
		return "."
	case 2:
		return ".."
	case 3:
		return "..."
	default:
		return ""
	}
}

func connectionPulseColor(status ConnectionStatus, step int) lipgloss.Color {
	switch status {
	case StatusConnected:
		palette := []lipgloss.Color{TelegramDark.AccentBlue, TelegramDark.AccentGreen, TelegramDark.AccentBlue, TelegramDark.TextSecondary}
		return palette[step%len(palette)]
	case StatusConnecting:
		palette := []lipgloss.Color{TelegramDark.AccentYellow, TelegramDark.TextSecondary, TelegramDark.AccentYellow}
		return palette[step%len(palette)]
	default:
		return TelegramDark.AccentRed
	}
}
