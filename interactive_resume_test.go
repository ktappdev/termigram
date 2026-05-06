package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestSyncTranscriptContextUsesRequestedLimit(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")

	originalLoader := cli.transcriptStore.transcriptMessageLoader
	defer func() { cli.transcriptStore.transcriptMessageLoader = originalLoader }()

	var (
		gotTarget string
		gotLimit  int
	)
	cli.transcriptStore.transcriptMessageLoader = func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
		gotTarget = target
		gotLimit = limit
		return []MessageOutput{{
			ID:       101,
			FromName: "Alice",
			Message:  "caught up",
			Text:     "caught up",
			Preview:  "caught up",
			Date:     1710000000,
		}}, nil
	}

	if err := cli.transcriptStore.syncTranscriptContext(context.Background(), cli, "@alice", "Alice", transcriptResumeFetchLimit); err != nil {
		t.Fatalf("syncTranscriptContext error: %v", err)
	}

	if gotTarget != "@alice" {
		t.Fatalf("expected target @alice, got %q", gotTarget)
	}
	if gotLimit != transcriptResumeFetchLimit {
		t.Fatalf("expected limit %d, got %d", transcriptResumeFetchLimit, gotLimit)
	}

	entries, loaded := cli.transcriptStore.transcriptSnapshot("@alice")
	if !loaded {
		t.Fatalf("expected transcript to be marked loaded")
	}
	if len(entries) != 1 || entries[0].MessageID != 101 {
		t.Fatalf("expected synced transcript entry, got %#v", entries)
	}
}

func TestMaybeResumeAfterIdleRefreshesActiveChat(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.setCurrentChat("@alice", "Alice")
	cli.cacheUser(&tg.User{ID: 1, FirstName: "Alice", Username: "alice"})
	cli.chatState.chatUnreadCount[normalizeUsername("@alice")] = 3

	originalLoader := cli.transcriptStore.transcriptMessageLoader
	originalDialogRefresher := cli.interactiveResumeDialogRefresher
	defer func() {
		cli.transcriptStore.transcriptMessageLoader = originalLoader
		cli.interactiveResumeDialogRefresher = originalDialogRefresher
	}()

	base := time.Unix(1710000000, 0)
	cli.nowFunc = func() time.Time { return base.Add(interactiveResumeIdleThreshold + time.Minute) }

	cli.chatState.lastInteractiveUse = base

	var dialogRefreshes int
	cli.interactiveResumeDialogRefresher = func(ctx context.Context, cli *TelegramCLI, limit int) ([]CachedChat, error) {
		dialogRefreshes++
		if limit != interactiveResumeDialogLimit {
			t.Fatalf("expected dialog limit %d, got %d", interactiveResumeDialogLimit, limit)
		}
		return nil, nil
	}
	cli.transcriptStore.transcriptMessageLoader = func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
		return []MessageOutput{{
			ID:       202,
			FromName: "Alice",
			Message:  "missed while asleep",
			Text:     "missed while asleep",
			Preview:  "missed while asleep",
			Date:     base.Add(2 * time.Minute).Unix(),
		}}, nil
	}

	cli.maybeResumeAfterIdle(context.Background())

	if dialogRefreshes != 1 {
		t.Fatalf("expected one dialog refresh, got %d", dialogRefreshes)
	}
	if got := cli.chatState.chatUnreadCount[normalizeUsername("@alice")]; got != 0 {
		t.Fatalf("expected unread count cleared, got %d", got)
	}

	entries, _ := cli.transcriptStore.transcriptSnapshot("@alice")
	if len(entries) != 1 || entries[0].MessageID != 202 {
		t.Fatalf("expected refreshed transcript entry, got %#v", entries)
	}
}

func TestRetryInteractiveRPCRetriesReconnectErrors(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")

	originalDialogRefresher := cli.interactiveResumeDialogRefresher
	defer func() { cli.interactiveResumeDialogRefresher = originalDialogRefresher }()
	cli.interactiveResumeDialogRefresher = func(ctx context.Context, cli *TelegramCLI, limit int) ([]CachedChat, error) {
		return nil, nil
	}

	var attempts int
	id, err := cli.retryInteractiveRPC(context.Background(), func(ctx context.Context) (int64, error) {
		attempts++
		if attempts == 1 {
			return 0, fmt.Errorf("connection reset by peer")
		}
		return 303, nil
	})
	if err != nil {
		t.Fatalf("retryInteractiveRPC error: %v", err)
	}
	if id != 303 {
		t.Fatalf("expected retry result 303, got %d", id)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryInteractiveRPCSkipsPermanentErrors(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")

	var attempts int
	_, err := cli.retryInteractiveRPC(context.Background(), func(ctx context.Context) (int64, error) {
		attempts++
		return 0, fmt.Errorf("user not found")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
