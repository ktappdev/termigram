package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a single rendered message row.
type Message struct {
	Text     string
	Time     string
	Sender   string
	Chat     string
	Outgoing bool
	Read     bool
}

// MessageViewModel renders the message pane with scroll support.
type MessageViewModel struct {
	Messages []Message
	Width    int
	Height   int
	Scroll   int

	BaseStyle     lipgloss.Style
	IncomingStyle lipgloss.Style
	OutgoingStyle lipgloss.Style
	MetaStyle     lipgloss.Style
}

// NewMessageView creates a minimal message pane with sample data.
func NewMessageView() MessageViewModel {
	return MessageViewModel{
		Messages: []Message{
			{Text: "Hey! How's it going?", Time: "10:30", Sender: "Alice", Chat: "@alice", Outgoing: false},
			{Text: "Pretty good — working on the UI components.", Time: "10:31", Sender: "You", Chat: "@alice", Outgoing: true, Read: true},
			{Text: "Nice. Can you share an update when done?", Time: "10:32", Sender: "Alice", Chat: "@alice", Outgoing: false},
			{Text: "Will do. Header and chat list are in.", Time: "10:33", Sender: "You", Chat: "@alice", Outgoing: true, Read: false},
		},
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

// Update handles window resize and scroll key input.
func (m MessageViewModel) Update(msg tea.Msg) MessageViewModel {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.clampScroll()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.Scroll--
		case "down", "j":
			m.Scroll++
		case "pgup":
			m.Scroll -= m.pageSize()
		case "pgdown":
			m.Scroll += m.pageSize()
		case "home":
			m.Scroll = 0
		case "end":
			m.Scroll = m.maxScroll()
		}
		m.clampScroll()
	}
	return m
}

// View renders visible messages, timestamps, and read receipts.
func (m MessageViewModel) View() string {
	start, end := m.visibleRange()
	if len(m.Messages) == 0 {
		return m.BaseStyle.Width(m.Width).Height(m.Height).Render("")
	}

	rows := make([]string, 0, (end-start)*2)
	if chat := m.chatContextLine(); chat != "" {
		rows = append(rows, chat, "")
	}
	for i, msg := range m.Messages[start:end] {
		rows = append(rows, m.renderMessage(msg))
		if i < end-start-1 {
			rows = append(rows, "")
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return m.BaseStyle.Width(m.Width).Height(m.Height).Render(content)
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
	case m.Width < 22:
		return style.Padding(spaceSm, spaceMd)
	case m.Width < 30:
		return style.Border(lipgloss.NormalBorder()).Padding(spaceSm, spaceMd)
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
	if maxOuter < 12 && available >= 12 {
		maxOuter = 12
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

func (m MessageViewModel) pageSize() int {
	if m.Height <= 1 {
		return 1
	}
	return m.Height / 2
}

func (m MessageViewModel) maxVisible() int {
	if m.Height <= 0 {
		return len(m.Messages)
	}
	// Rough estimate: padded bubble messages plus spacing tend to use ~6 lines.
	visible := m.Height / 6
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (m MessageViewModel) maxScroll() int {
	visible := m.maxVisible()
	if len(m.Messages) <= visible {
		return 0
	}
	return len(m.Messages) - visible
}

func (m *MessageViewModel) clampScroll() {
	if m.Scroll < 0 {
		m.Scroll = 0
	}
	max := m.maxScroll()
	if m.Scroll > max {
		m.Scroll = max
	}
}

func (m MessageViewModel) visibleRange() (int, int) {
	visible := m.maxVisible()
	start := m.Scroll
	end := start + visible
	if end > len(m.Messages) {
		end = len(m.Messages)
	}
	if start > end {
		start = end
	}
	return start, end
}
