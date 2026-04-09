package main

import (
	"fmt"
	"strings"
)

func replyQuoteText(ref *ReplyReference) string {
	if !isValidReplyReference(ref) {
		return ""
	}
	sender := normalizeReplyPreview(ref.Sender)
	preview := normalizeReplyPreview(ref.Preview)
	switch {
	case sender != "" && preview != "":
		return fmt.Sprintf("↪ %s\n> %s", sender, preview)
	case sender != "":
		return fmt.Sprintf("↪ %s", sender)
	case preview != "":
		return fmt.Sprintf("↪ %s", preview)
	default:
		return fmt.Sprintf("↪ Message #%d", ref.MessageID)
	}
}

func composeMessageBody(base string, reply *ReplyReference) string {
	base = strings.TrimSpace(base)
	quote := replyQuoteText(reply)
	if quote == "" {
		return base
	}
	if base == "" {
		return quote
	}
	return quote + "\n\n" + base
}

func messageBodyText(messageID int64, text string, attachment *ImageAttachment, reply *ReplyReference) string {
	base := strings.TrimSpace(text)
	if attachment != nil {
		base = imagePlaceholderBody(messageID, attachment.Name, text)
	}
	return composeMessageBody(base, reply)
}

func messagePreviewText(text string, attachment *ImageAttachment) string {
	if attachment == nil {
		return strings.TrimSpace(text)
	}
	return imagePreviewText(attachment, text)
}

func entryPreviewText(entry transcriptEntry) string {
	if preview := normalizeReplyPreview(entry.Preview); preview != "" {
		return preview
	}
	if preview := normalizeReplyPreview(messagePreviewText(entry.Text, entry.Image)); preview != "" {
		return preview
	}
	if body := strings.TrimSpace(entry.Body); body != "" {
		lines := strings.Split(body, "\n")
		if len(lines) > 0 {
			return normalizeReplyPreview(lines[len(lines)-1])
		}
	}
	return "[message]"
}
