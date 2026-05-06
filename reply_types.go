package main

import "strings"

// SendOptions carries optional send-time metadata shared by text and image sends.
type SendOptions struct {
	ReplyToMessageID int64
	Silent           bool
}

// ReplyReference is the human-readable reply target metadata exposed in JSON output
// and used for transcript rendering.
type ReplyReference struct {
	MessageID int64  `json:"message_id"`
	Sender    string `json:"sender,omitempty"`
	Preview   string `json:"preview,omitempty"`
}

type pendingReplyState struct {
	Target string
	Ref    ReplyReference
}

func cloneReplyReference(ref *ReplyReference) *ReplyReference {
	if ref == nil {
		return nil
	}
	clone := *ref
	return &clone
}

func normalizeReplyTarget(target string) string {
	return normalizeUsername(target)
}

func isValidReplyReference(ref *ReplyReference) bool {
	return ref != nil && ref.MessageID > 0
}

func normalizeReplyPreview(text string) string {
	text = strings.ReplaceAll(strings.TrimSpace(text), "\n", " ")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return text
}
