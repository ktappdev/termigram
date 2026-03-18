package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds base reusable style definitions for the TUI.
type Styles struct {
	App     lipgloss.Style
	Header  lipgloss.Style
	Sidebar lipgloss.Style
	Content lipgloss.Style
	Input   lipgloss.Style
}

// DefaultStyles returns minimal theme-aware styles for bootstrapping the UI package.
func DefaultStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Background(TelegramDark.BgPrimary).
			Foreground(TelegramDark.TextPrimary),
		Header: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary).
			Padding(0, 1),
		Sidebar: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary),
		Content: lipgloss.NewStyle().
			Background(TelegramDark.BgPrimary).
			Foreground(TelegramDark.TextPrimary),
		Input: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary).
			Padding(0, 1),
	}
}
