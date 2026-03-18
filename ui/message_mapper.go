package ui

func mapBackendMessages(in []BackendMessage, chatTitle string, chatTarget string) []Message {
	out := make([]Message, 0, len(in))
	for _, msg := range in {
		out = append(out, Message{
			Text:     msg.Text,
			Time:     msg.Time,
			Sender:   msg.Sender,
			Chat:     chatLabel(chatTitle, chatTarget, msg.Chat),
			Outgoing: msg.Outgoing,
			Read:     msg.Read,
		})
	}
	return out
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
