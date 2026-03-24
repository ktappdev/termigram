package main

import (
	"strings"
	"testing"
)

func TestRenderActiveLegacyChatViewShowsPendingReplyBanner(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.setCurrentChat("@alice", "Alice")
	cli.setPendingReply("@alice", &ReplyReference{MessageID: 5, Sender: "Alice", Preview: "hello"})

	view := cli.renderActiveLegacyChatView("Alice", "@alice", []legacyTranscriptEntry{{
		MessageID: 6,
		Outgoing:  true,
		Sender:    "You",
		Header:    "You",
		Body:      "ok",
		Meta:      "07:00:00  ✓",
		Text:      "ok",
		Preview:   "ok",
	}}, 80, 12)

	if !strings.Contains(view, "Replying to Alice: hello") {
		t.Fatalf("expected pending reply banner in view, got %q", view)
	}
}
