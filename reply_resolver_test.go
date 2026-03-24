package main

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/tg"
)

func TestResolveReplyReferencesForMessagesUsesBatchMessages(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.cacheUser(&tg.User{ID: 42, FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 99, FirstName: "Bob"})

	original := &tg.Message{
		ID:      41,
		FromID:  &tg.PeerUser{UserID: 42},
		Message: "hello there",
	}
	reply := &tg.Message{
		ID:      42,
		FromID:  &tg.PeerUser{UserID: 99},
		Message: "sounds good",
		ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 41},
	}

	refs := resolveReplyReferencesForMessages(context.Background(), cli, []*tg.Message{reply, original})
	ref := refs[42]
	if ref == nil {
		t.Fatalf("expected reply reference")
	}
	if ref.MessageID != 41 {
		t.Fatalf("expected reply message id 41, got %d", ref.MessageID)
	}
	if ref.Sender != "Alice" {
		t.Fatalf("expected sender Alice, got %q", ref.Sender)
	}
	if ref.Preview != "hello there" {
		t.Fatalf("expected preview from original message, got %q", ref.Preview)
	}
}

func TestResolveReplyReferencesForMessagesFallsBackToFetch(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.cacheUser(&tg.User{ID: 42, FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 99, FirstName: "Bob"})

	originalFetch := fetchReplyMessagesFunc
	defer func() { fetchReplyMessagesFunc = originalFetch }()
	fetchReplyMessagesFunc = func(ctx context.Context, cli *TelegramCLI, ids []int64) (map[int64]*tg.Message, error) {
		return map[int64]*tg.Message{
			55: {
				ID:      55,
				FromID:  &tg.PeerUser{UserID: 42},
				Message: "from fetch",
			},
		}, nil
	}

	reply := &tg.Message{
		ID:      77,
		FromID:  &tg.PeerUser{UserID: 99},
		Message: "later reply",
		ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 55},
	}

	refs := resolveReplyReferencesForMessages(context.Background(), cli, []*tg.Message{reply})
	ref := refs[77]
	if ref == nil {
		t.Fatalf("expected reply reference from fetch")
	}
	if ref.Sender != "Alice" || ref.Preview != "from fetch" {
		t.Fatalf("unexpected fetched ref: %+v", *ref)
	}
}

func TestFallbackReplyReferenceUsesQuoteText(t *testing.T) {
	originalFetch := fetchReplyMessagesFunc
	defer func() { fetchReplyMessagesFunc = originalFetch }()
	fetchReplyMessagesFunc = func(ctx context.Context, cli *TelegramCLI, ids []int64) (map[int64]*tg.Message, error) {
		return nil, errors.New("boom")
	}

	msg := &tg.Message{
		ID:      90,
		Message: "reply body",
		ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 12, QuoteText: "quoted bit"},
	}
	ref := fallbackReplyReference(msg)
	if ref == nil {
		t.Fatalf("expected fallback reply ref")
	}
	if ref.Preview != "quoted bit" {
		t.Fatalf("expected quote text preview, got %q", ref.Preview)
	}
}
