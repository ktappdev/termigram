package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	composerMinHeight = 3
	composerMaxHeight = 6
)

// InputAreaModel renders the compose area at the bottom of the UI.
type InputAreaModel struct {
	Input      textarea.Model
	Width      int
	ReplyTo    string
	Typing     bool
	TypingStep int
	Sent       []string

	ContainerStyle lipgloss.Style
	ReplyStyle     lipgloss.Style
	TypingStyle    lipgloss.Style
	HintStyle      lipgloss.Style
}

// NewInputArea creates a minimal themed input area.
func NewInputArea() InputAreaModel {
	composer := textarea.New()
	composer.Prompt = "┃ "
	composer.Placeholder = "Write a message..."
	composer.CharLimit = 0
	composer.ShowLineNumbers = false
	composer.SetHeight(composerMinHeight)
	composer.MaxHeight = composerMaxHeight
	composer.FocusedStyle.CursorLine = lipgloss.NewStyle()
	composer.BlurredStyle.CursorLine = lipgloss.NewStyle()
	composer.KeyMap.InsertNewline.SetEnabled(false)
	composer.Focus()
	composer.SetWidth(40)

	return InputAreaModel{
		Input: composer,
		ContainerStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgSecondary).
			Foreground(TelegramDark.TextPrimary).
			Padding(spaceSm, spaceMd),
		ReplyStyle: lipgloss.NewStyle().
			Foreground(TelegramDark.TextSecondary),
		TypingStyle: lipgloss.NewStyle().
			Foreground(TelegramDark.AccentBlue),
		HintStyle: lipgloss.NewStyle().
			Foreground(TelegramDark.TextMuted),
	}
}

func (m InputAreaModel) InitCmd() tea.Cmd {
	return textarea.Blink
}

// Update handles typing, send action, and shortcuts.
func (m InputAreaModel) Update(msg tea.Msg) InputAreaModel {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Input.SetWidth(m.inputWidth())
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			text := strings.TrimSpace(m.Input.Value())
			if text != "" {
				m.Sent = append(m.Sent, text)
				m.Input.Reset()
				m.Input.SetHeight(composerMinHeight)
				m.Typing = false
				m.TypingStep = 0
			}
			return m
		case "esc":
			m.ReplyTo = ""
			m.Input.Reset()
			m.Input.SetHeight(composerMinHeight)
			m.Typing = false
			return m
		case "ctrl+o":
			if strings.TrimSpace(m.Input.Value()) != "" {
				m.Input.InsertString(" ")
			}
			m.Input.InsertString("[attach]")
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	_ = cmd
	m.Input.SetWidth(m.inputWidth())
	m.adjustHeight()
	m.Typing = strings.TrimSpace(m.Input.Value()) != ""
	if !m.Typing {
		m.TypingStep = 0
	}
	return m
}

// View renders reply context, typing indicator, input row, and action hints.
func (m InputAreaModel) View() string {
	rows := make([]string, 0, 4)

	if m.ReplyTo != "" {
		reply := m.ReplyStyle.MaxWidth(m.Width).Width(m.Width).Render(fmt.Sprintf("↪ Replying to %s (Esc to cancel)", m.ReplyTo))
		rows = append(rows, reply)
	}

	if m.Typing {
		rows = append(rows, m.TypingStyle.MaxWidth(m.Width).Width(m.Width).Render("draft in progress"))
	}

	composer := m.ContainerStyle.Width(widthWithinStyle(m.Width, m.ContainerStyle)).Render(m.Input.View())
	rows = append(rows, composer)

	hints := []string{"[Send Enter]", "[Search Ctrl+F]", "[Attach Ctrl+O]", "[Cancel Esc]"}
	rows = append(rows, m.HintStyle.MaxWidth(m.Width).Width(m.Width).Render(strings.Join(hints, strings.Repeat(" ", spaceLg))))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m InputAreaModel) inputWidth() int {
	if m.Width <= 0 {
		return 40
	}
	available := m.Width - m.ContainerStyle.GetHorizontalFrameSize()
	if available < 12 {
		return 12
	}
	return available
}

func (m *InputAreaModel) adjustHeight() {
	height := m.Input.LineCount()
	if height < composerMinHeight {
		height = composerMinHeight
	}
	if height > composerMaxHeight {
		height = composerMaxHeight
	}
	m.Input.SetHeight(height)
}
