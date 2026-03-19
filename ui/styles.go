package ui

import "github.com/charmbracelet/lipgloss"

const (
	spaceXs = 0
	spaceSm = 1
	spaceMd = 2
	spaceLg = 3
)

// Styles holds base reusable style definitions for the TUI.
type Styles struct {
	App           lipgloss.Style
	Header        lipgloss.Style
	Sidebar       lipgloss.Style
	SidebarBorder lipgloss.Style
	Content       lipgloss.Style
	Input         lipgloss.Style
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
			Padding(spaceXs, spaceSm),
		Sidebar: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary),
		SidebarBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(TelegramDark.BgHover),
		Content: lipgloss.NewStyle().
			Background(TelegramDark.BgPrimary).
			Foreground(TelegramDark.TextPrimary),
		Input: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceXs, spaceSm),
	}
}
