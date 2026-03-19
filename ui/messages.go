package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a single rendered message row.
type Message struct {
	ID       int64
	Text     string
	Time     string
	Sender   string
	Chat     string
	Outgoing bool
	Read     bool
}

// MessageViewModel renders the message pane with viewport-backed scrolling.
type MessageViewModel struct {
	Messages []Message
	Width    int
	Height   int
	Viewport viewport.Model
	ready    bool

	BaseStyle     lipgloss.Style
	IncomingStyle lipgloss.Style
	OutgoingStyle lipgloss.Style
	MetaStyle     lipgloss.Style
}

// NewMessageView creates a minimal message pane.
func NewMessageView() MessageViewModel {
	vp := viewport.New(1, 1)
	vp.MouseWheelEnabled = true
	return MessageViewModel{
		Messages: []Message{},
		Viewport: vp,
		BaseStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgPrimary).
			Foreground(TelegramDark.TextPrimary),
		IncomingStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgMessageIn).
			Foreground(TelegramDark.TextPrimary),
		OutgoingStyle: lipgloss.NewStyle().
			Background(TelegramDark.BgMessageOut).
			Foreground(TelegramDark.TextPrimary),
		MetaStyle: lipgloss.NewStyle().
			Foreground(TelegramDark.TextSecondary),
	}
}

// Update handles window resize and scroll input.
func (m MessageViewModel) Update(msg tea.Msg) MessageViewModel {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg.Width, msg.Height)
		return m
	case tea.KeyMsg, tea.MouseMsg:
		if !m.ready {
			m.Resize(m.Width, m.Height)
		}
		var cmd tea.Cmd
		m.Viewport, cmd = m.Viewport.Update(msg)
		_ = cmd
	}
	return m
}

// View renders visible messages, timestamps, and read receipts.
func (m MessageViewModel) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return m.BaseStyle.Width(m.Width).Height(m.Height).Render("")
	}
	if !m.ready {
		return m.BaseStyle.Width(m.Width).Height(m.Height).Render("")
	}
	if len(m.Messages) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(TelegramDark.TextMuted).
			Padding(1, 2).
			Render("No messages yet.")
		return m.BaseStyle.Width(m.Width).Height(m.Height).Render(empty)
	}
	return m.BaseStyle.Width(m.Width).Height(m.Height).Render(m.Viewport.View())
}

func (m *MessageViewModel) Resize(width int, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	wasAtBottom := m.ready && m.Viewport.AtBottom()
	prevOffset := m.Viewport.YOffset
	m.Width = width
	m.Height = height
	if !m.ready {
		m.Viewport = viewport.New(width, height)
		m.Viewport.MouseWheelEnabled = true
		m.ready = true
	} else {
		m.Viewport.Width = width
		m.Viewport.Height = height
	}
	m.syncViewport(wasAtBottom, prevOffset)
}

func (m *MessageViewModel) SetMessages(messages []Message) {
	m.Messages = append([]Message(nil), messages...)
	m.syncViewport(true, m.Viewport.YOffset)
}

func (m *MessageViewModel) AppendMessage(message Message, pinBottom bool) {
	if message.ID != 0 && m.hasMessageID(message.ID) {
		return
	}
	m.Messages = append(m.Messages, message)
	m.syncViewport(pinBottom, m.Viewport.YOffset)
}

func (m *MessageViewModel) RemoveLastMessage() bool {
	if len(m.Messages) == 0 {
		return false
	}
	m.Messages = m.Messages[:len(m.Messages)-1]
	m.syncViewport(true, m.Viewport.YOffset)
	return true
}

func (m MessageViewModel) hasMessageID(id int64) bool {
	if id == 0 {
		return false
	}
	for _, message := range m.Messages {
		if message.ID == id {
			return true
		}
	}
	return false
}

func (m *MessageViewModel) syncViewport(pinBottom bool, prevOffset int) {
	if !m.ready {
		return
	}
	m.Viewport.SetContent(m.renderTranscript())
	if pinBottom {
		m.Viewport.GotoBottom()
		return
	}
	m.Viewport.SetYOffset(prevOffset)
}

func (m MessageViewModel) renderTranscript() string {
	rows := make([]string, 0, len(m.Messages)*2+2)
	if chat := m.chatContextLine(); chat != "" {
		rows = append(rows, chat, "")
	}
	for i, msg := range m.Messages {
		rows = append(rows, m.renderMessage(msg))
		if i < len(m.Messages)-1 {
			rows = append(rows, "")
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m MessageViewModel) renderMessage(msg Message) string {
	style := m.bubbleStyle(msg.Outgoing)
	bubble := style.MaxWidth(m.bubbleWidth(style)).Render(m.messageText(msg))
	bubbleWidth := lipgloss.Width(bubble)

	parts := make([]string, 0, 3)
	if senderLabel := m.senderLabel(msg, bubbleWidth); senderLabel != "" {
		parts = append(parts, senderLabel)
	}
	parts = append(parts, bubble)
	parts = append(parts, m.metaLine(msg, bubbleWidth))

	align := lipgloss.Left
	if msg.Outgoing {
		align = lipgloss.Right
	}

	block := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.PlaceHorizontal(m.messagePaneWidth(), align, block)
}

func (m MessageViewModel) chatContextLine() string {
	if len(m.Messages) == 0 || m.Messages[0].Chat == "" {
		return ""
	}
	return m.MetaStyle.Render("Chat: " + m.Messages[0].Chat)
}

func (m MessageViewModel) bubbleStyle(outgoing bool) lipgloss.Style {
	style := m.IncomingStyle.BorderForeground(TelegramDark.AccentGreen)
	if outgoing {
		style = m.OutgoingStyle.BorderForeground(TelegramDark.AccentBlue)
	}

	switch {
	case m.Width < 18:
		return style.Padding(spaceSm, 1)
	case m.Width < 24:
		return style.Border(lipgloss.NormalBorder()).Padding(spaceSm, 1)
	default:
		return style.Border(lipgloss.RoundedBorder()).Padding(spaceSm, spaceMd)
	}
}

func (m MessageViewModel) bubbleWidth(style lipgloss.Style) int {
	available := m.messagePaneWidth() - 2
	if available < 1 {
		available = 1
	}

	maxOuter := available
	switch {
	case available >= 96:
		maxOuter = 72
	case available >= 72:
		maxOuter = (available * 2) / 3
	case available >= 40:
		maxOuter = (available * 7) / 10
	case available >= 24:
		maxOuter = (available * 4) / 5
	}

	if maxOuter > available {
		maxOuter = available
	}
	if maxOuter < 10 && available >= 10 {
		maxOuter = 10
	}

	innerWidth := maxOuter - style.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	return innerWidth
}

func (m MessageViewModel) messagePaneWidth() int {
	if m.Width <= 0 {
		return 60
	}
	return m.Width
}

func (m MessageViewModel) messageText(msg Message) string {
	if msg.Text == "" {
		return " "
	}
	return msg.Text
}

func (m MessageViewModel) senderLabel(msg Message, width int) string {
	if msg.Outgoing {
		return ""
	}

	senderText := strings.TrimSpace(msg.Sender)
	if senderText == "" || senderText == "Unknown" || m.isRedundantSenderLabel(senderText, msg.Chat) {
		return ""
	}

	if width < 1 {
		width = m.bubbleWidth(m.bubbleStyle(msg.Outgoing))
	}
	if width < 1 {
		width = 1
	}

	return lipgloss.NewStyle().
		Foreground(TelegramDark.AccentGreen).
		Bold(true).
		Width(width).
		MaxWidth(width).
		Render(senderText)
}

func (m MessageViewModel) isRedundantSenderLabel(sender string, chat string) bool {
	normalizedSender := normalizeLabel(sender)
	if normalizedSender == "" {
		return true
	}

	for _, candidate := range chatLabelCandidates(chat) {
		if normalizedSender == normalizeLabel(candidate) {
			return true
		}
	}
	return false
}

func chatLabelCandidates(chat string) []string {
	chat = strings.TrimSpace(chat)
	if chat == "" {
		return nil
	}

	candidates := []string{chat}
	if open := strings.LastIndex(chat, " ("); open != -1 && strings.HasSuffix(chat, ")") {
		title := strings.TrimSpace(chat[:open])
		target := strings.TrimSpace(chat[open+2 : len(chat)-1])
		if title != "" {
			candidates = append(candidates, title)
		}
		if target != "" {
			candidates = append(candidates, target)
		}
	}
	return candidates
}

func normalizeLabel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "@")
	return value
}

func (m MessageViewModel) metaLine(msg Message, width int) string {
	receipt := ""
	if msg.Outgoing {
		receipt = "✓"
		if msg.Read {
			receipt = "✓✓"
		}
	}

	meta := strings.TrimSpace(fmt.Sprintf("%s %s", msg.Time, receipt))
	align := lipgloss.Left
	if msg.Outgoing {
		align = lipgloss.Right
	}

	if width < 1 {
		width = 1
	}

	return m.MetaStyle.Width(width).Align(align).Render(meta)
}
