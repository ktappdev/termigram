package main

import "testing"

import "github.com/gotd/td/tg"

func TestMessageFromShortUserUpdatePreservesReplyHeader(t *testing.T) {
	update := &tg.UpdateShortMessage{
		ID:      10,
		UserID:  42,
		Message: "reply",
		ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 7},
	}

	msg := messageFromShortUserUpdate(update)
	if msg == nil {
		t.Fatalf("expected message")
	}
	if replyToID, ok := replyMessageID(msg); !ok || replyToID != 7 {
		t.Fatalf("expected reply header to be preserved, got ok=%v id=%d", ok, replyToID)
	}
}

func TestSentMessageIDFromUpdatesHandlesShortSentMessage(t *testing.T) {
	updates := &tg.UpdateShortSentMessage{ID: 321}
	if got := sentMessageIDFromUpdates(updates); got != 321 {
		t.Fatalf("expected sent message id 321, got %d", got)
	}
}
