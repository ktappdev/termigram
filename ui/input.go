package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputAreaModel renders the compose area at the bottom of the UI.
type InputAreaModel struct {
	Input      textinput.Model
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
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "Write a message..."
	ti.Focus()
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(TelegramDark.AccentBlue)

	return InputAreaModel{
		Input: ti,
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

// Update handles typing, send action, and shortcuts.
func (m InputAreaModel) Update(msg tea.Msg) InputAreaModel {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Input.Width = m.inputWidth()
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			text := strings.TrimSpace(m.Input.Value())
			if text != "" {
				m.Sent = append(m.Sent, text)
				m.Input.SetValue("")
				m.Typing = false
				m.TypingStep = 0
			}
			return m
		case "esc":
			m.ReplyTo = ""
			m.Input.SetValue("")
			m.Typing = false
		case "ctrl+o":
			if strings.TrimSpace(m.Input.Value()) != "" {
				m.Input.SetValue(m.Input.Value() + " ")
			}
			m.Input.SetValue(m.Input.Value() + "[attach]")
		case "ctrl+t":
			m.Typing = !m.Typing
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	_ = cmd
	m.Input.Width = m.inputWidth()

	if strings.TrimSpace(m.Input.Value()) != "" {
		m.Typing = true
	} else if !m.Typing {
		m.TypingStep = 0
	}

	return m
}

// View renders reply context, typing indicator, input row, and action hints.
func (m InputAreaModel) View() string {
	rows := []string{}

	if m.ReplyTo != "" {
		reply := m.ReplyStyle.Render(fmt.Sprintf("↪ Replying to %s (Esc to cancel)", m.ReplyTo))
		rows = append(rows, reply)
	}

	if m.Typing {
		rows = append(rows, m.TypingStyle.Render("typing"+typingDotsFrame(m.TypingStep)))
	}

	inputRow := m.ContainerStyle.Width(m.Width).Render(m.Input.View() + "   [Send ⏎]")
	rows = append(rows, inputRow)

	hints := m.HintStyle.Render(strings.Join([]string{"[Attach Ctrl+O]", "[Cancel Esc]", "[Typing Ctrl+T]"}, strings.Repeat(" ", spaceLg)))
	rows = append(rows, hints)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m InputAreaModel) inputWidth() int {
	if m.Width <= 0 {
		return 40
	}
	if m.Width <= 20 {
		return m.Width
	}
	width := m.Width - 20
	if width < 1 {
		return 1
	}
	return width
}
