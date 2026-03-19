package ui

import "strings"

func mapBackendMessages(in []BackendMessage, chatTitle string, chatTarget string) []Message {
	out := make([]Message, 0, len(in))
	for _, msg := range in {
		out = append(out, Message{
			ID:       msg.ID,
			Text:     msg.Text,
			Time:     msg.Time,
			Sender:   msg.Sender,
			Chat:     chatLabel(chatTitle, chatTarget, msg.Chat),
			Outgoing: msg.Outgoing,
			Read:     msg.Read,
		})
	}
	return filterTranscriptEchoes(out)
}

func filterTranscriptEchoes(messages []Message) []Message {
	if len(messages) < 2 {
		return messages
	}

	filtered := make([]Message, 0, len(messages))
	for i, msg := range messages {
		payload, ok := transcriptEchoPayload(msg.Text)
		if ok && !msg.Outgoing && neighboringOutgoingMessage(messages, i, payload) {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func neighboringOutgoingMessage(messages []Message, idx int, payload string) bool {
	for _, neighbor := range []int{idx - 1, idx + 1} {
		if neighbor < 0 || neighbor >= len(messages) {
			continue
		}
		candidate := messages[neighbor]
		if candidate.Outgoing && strings.TrimSpace(candidate.Text) == payload {
			return true
		}
	}
	return false
}

func chatLabel(title string, target string, fallback string) string {
	switch {
	case title != "" && target != "":
		return title + " (" + target + ")"
	case title != "":
		return title
	case target != "":
		return target
	default:
		return fallback
	}
}
