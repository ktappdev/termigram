package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

var (
	_ tea.Model
	_ = help.Model{}
	_ = lipgloss.NewStyle
)
