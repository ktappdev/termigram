package ui

import "github.com/charmbracelet/lipgloss"

// TelegramDark contains colors aligned with Telegram Desktop's dark theme.
var TelegramDark = struct {
	BgPrimary     lipgloss.Color
	BgSecondary   lipgloss.Color
	BgMessageIn   lipgloss.Color
	BgMessageOut  lipgloss.Color
	BgHover       lipgloss.Color
	BgSelected    lipgloss.Color
	TextPrimary   lipgloss.Color
	TextSecondary lipgloss.Color
	TextMuted     lipgloss.Color
	AccentBlue    lipgloss.Color
	AccentGreen   lipgloss.Color
	AccentYellow  lipgloss.Color
	AccentRed     lipgloss.Color
}{
	BgPrimary:     lipgloss.Color("#17212b"),
	BgSecondary:   lipgloss.Color("#0e1621"),
	BgMessageIn:   lipgloss.Color("#1d3a2a"),
	BgMessageOut:  lipgloss.Color("#1f3f5b"),
	BgHover:       lipgloss.Color("#202b36"),
	BgSelected:    lipgloss.Color("#2b5278"),
	TextPrimary:   lipgloss.Color("#ffffff"),
	TextSecondary: lipgloss.Color("#7a8b99"),
	TextMuted:     lipgloss.Color("#5c6c7a"),
	AccentBlue:    lipgloss.Color("#5ca0d3"),
	AccentGreen:   lipgloss.Color("#5dc97d"),
	AccentYellow:  lipgloss.Color("#faa619"),
	AccentRed:     lipgloss.Color("#e17076"),
}
