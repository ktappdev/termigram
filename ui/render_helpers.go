package ui

import "github.com/charmbracelet/lipgloss"

func widthWithinStyle(total int, style lipgloss.Style) int {
	if total <= 0 {
		return 0
	}
	width := total - style.GetHorizontalFrameSize()
	if width < 1 {
		return 1
	}
	return width
}
