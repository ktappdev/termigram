package main

import (
	"context"
	"testing"
)

func TestEnsureLegacyTranscriptContextFetchesWhenOnlyBufferedUnreadEntryExists(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.appendLegacyTranscriptEntry("@alice", legacyTranscriptEntry{
		MessageID: 99,
		Header:    "Alice",
		Body:      "buffered unread",
		Meta:      "09:05:00",
	})

	originalLoader := legacyTranscriptMessageLoader
	defer func() { legacyTranscriptMessageLoader = originalLoader }()

	calls := 0
	legacyTranscriptMessageLoader = func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
		calls++
		if target != "@alice" {
			t.Fatalf("expected target @alice, got %q", target)
		}
		if limit != legacyTranscriptHistoryFetchLimit {
			t.Fatalf("expected history fetch limit %d, got %d", legacyTranscriptHistoryFetchLimit, limit)
		}
		return []MessageOutput{
			{ID: 1, FromName: "Alice", Message: "older context", Date: 100},
			{ID: 2, FromName: "Alice", Message: "newer context", Date: 200},
		}, nil
	}

	if err := cli.ensureLegacyTranscriptContext(context.Background(), "@alice", "Alice", unreadTranscriptMinContextEntries); err != nil {
		t.Fatalf("ensureLegacyTranscriptContext returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one history fetch, got %d", calls)
	}

	snapshot, loaded := cli.legacyTranscriptSnapshot("@alice")
	if !loaded {
		t.Fatalf("expected transcript to be marked loaded after fetch")
	}
	if got := len(snapshot); got != 3 {
		t.Fatalf("expected merged transcript of 3 entries, got %d", got)
	}
	if snapshot[0].Body != "newer context" || snapshot[1].Body != "older context" || snapshot[2].Body != "buffered unread" {
		t.Fatalf("expected fetched history to be merged ahead of buffered unread entry, got %#v", snapshot)
	}
}

func TestEnsureLegacyTranscriptContextSkipsFetchWhenLoadedTranscriptAlreadyHasContext(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.mergeLegacyTranscriptEntries("@alice", []legacyTranscriptEntry{
		{MessageID: 1, Body: "older"},
		{MessageID: 2, Body: "newer"},
	})

	originalLoader := legacyTranscriptMessageLoader
	defer func() { legacyTranscriptMessageLoader = originalLoader }()

	calls := 0
	legacyTranscriptMessageLoader = func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
		calls++
		return nil, nil
	}

	if err := cli.ensureLegacyTranscriptContext(context.Background(), "@alice", "Alice", unreadTranscriptMinContextEntries); err != nil {
		t.Fatalf("ensureLegacyTranscriptContext returned error: %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected no history fetch when loaded transcript already has context, got %d", calls)
	}
}
