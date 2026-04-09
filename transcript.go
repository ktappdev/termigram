package main

import (
	"context"
	"fmt"
	"strings"
)

const (
	transcriptLimit                   = 100
	transcriptHistoryFetchLimit       = 20
	transcriptResumeFetchLimit        = 50
	unreadTranscriptMinContextEntries = 2
)

type transcriptEntry struct {
	MessageID int64
	Outgoing  bool
	Sender    string
	Header    string
	Body      string
	Meta      string
	Text      string
	Preview   string
	Reply     *ReplyReference
	Image     *ImageAttachment
}

var transcriptMessageLoader = func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
	backend := &UserBackend{cli: cli}
	return backend.GetMessages(ctx, target, limit)
}

func (cli *TelegramCLI) setConsole(console *console) {
	cli.transcriptMu.Lock()
	defer cli.transcriptMu.Unlock()
	cli.console = console
}

func (cli *TelegramCLI) currentConsole() *console {
	cli.transcriptMu.RLock()
	defer cli.transcriptMu.RUnlock()
	return cli.console
}

func normalizeTranscriptTarget(target string) string {
	return normalizeUsername(target)
}

func (cli *TelegramCLI) appendTranscriptEntry(target string, entry transcriptEntry) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" {
		return
	}

	cli.transcriptMu.Lock()
	defer cli.transcriptMu.Unlock()

	existing := cli.transcripts[normalized]
	if entry.MessageID != 0 {
		for _, current := range existing {
			if current.MessageID == entry.MessageID {
				return
			}
		}
	}

	existing = append(existing, entry)
	if len(existing) > transcriptLimit {
		existing = existing[len(existing)-transcriptLimit:]
	}
	cli.transcripts[normalized] = existing
}

func (cli *TelegramCLI) transcriptSnapshot(target string) ([]transcriptEntry, bool) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" {
		return nil, false
	}

	cli.transcriptMu.RLock()
	defer cli.transcriptMu.RUnlock()

	entries := append([]transcriptEntry(nil), cli.transcripts[normalized]...)
	return entries, cli.transcriptLoaded[normalized]
}

func (cli *TelegramCLI) mergeTranscriptEntries(target string, fetched []transcriptEntry) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" {
		return
	}

	cli.transcriptMu.Lock()
	defer cli.transcriptMu.Unlock()

	merged := mergeTranscriptEntrySlices(fetched, cli.transcripts[normalized])
	if len(merged) > transcriptLimit {
		merged = merged[len(merged)-transcriptLimit:]
	}
	cli.transcripts[normalized] = merged
	cli.transcriptLoaded[normalized] = true
}

func mergeTranscriptEntrySlices(primary []transcriptEntry, secondary []transcriptEntry) []transcriptEntry {
	out := make([]transcriptEntry, 0, len(primary)+len(secondary))
	seenIDs := make(map[int64]struct{}, len(primary)+len(secondary))
	seenKeys := make(map[string]struct{}, len(primary)+len(secondary))

	appendUnique := func(entries []transcriptEntry) {
		for _, entry := range entries {
			if entry.MessageID != 0 {
				if _, exists := seenIDs[entry.MessageID]; exists {
					continue
				}
				seenIDs[entry.MessageID] = struct{}{}
				seenKeys[transcriptEntryKey(entry)] = struct{}{}
			} else {
				key := transcriptEntryKey(entry)
				if _, exists := seenKeys[key]; exists {
					continue
				}
				seenKeys[key] = struct{}{}
			}
			out = append(out, entry)
		}
	}

	appendUnique(primary)
	appendUnique(secondary)
	return out
}

func transcriptEntryKey(entry transcriptEntry) string {
	imageKey := ""
	if entry.Image != nil {
		imageKey = strings.TrimSpace(entry.Image.Kind) + "|" + strings.TrimSpace(entry.Image.Name) + "|" + strings.TrimSpace(entry.Image.MIMEType)
	}
	return fmt.Sprintf("%t|%s|%s|%s|%s", entry.Outgoing, strings.TrimSpace(entry.Header), strings.TrimSpace(entry.Body), strings.TrimSpace(entry.Meta), imageKey)
}

func (cli *TelegramCLI) ensureTranscript(ctx context.Context, target string, label string) error {
	return cli.ensureTranscriptContext(ctx, target, label, 1)
}

func (cli *TelegramCLI) ensureTranscriptContext(ctx context.Context, target string, label string, minEntries int) error {
	if ctx == nil {
		return nil
	}
	if minEntries < 1 {
		minEntries = 1
	}

	entries, loaded := cli.transcriptSnapshot(target)
	if loaded && len(entries) >= minEntries {
		return nil
	}

	return cli.syncTranscriptContext(ctx, target, label, transcriptHistoryFetchLimit)
}

func (cli *TelegramCLI) syncTranscriptContext(ctx context.Context, target string, label string, limit int) error {
	if ctx == nil {
		return nil
	}
	if limit <= 0 {
		limit = transcriptHistoryFetchLimit
	}

	messages, err := transcriptMessageLoader(ctx, cli, target, limit)
	if err != nil {
		return err
	}

	fetchedEntries := transcriptEntriesFromMessages(target, label, messages)
	cli.mergeTranscriptEntries(target, fetchedEntries)
	return nil
}

func transcriptEntriesFromMessages(target string, label string, messages []MessageOutput) []transcriptEntry {
	entries := make([]transcriptEntry, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		entries = append(entries, transcriptEntryFromMessageOutput(target, label, messages[i]))
	}
	return entries
}
