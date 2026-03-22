package main

import (
	"context"
	"fmt"
	"strings"
)

const (
	legacyTranscriptLimit             = 100
	legacyTranscriptHistoryFetchLimit = 20
	unreadTranscriptMinContextEntries = 2
)

type legacyTranscriptEntry struct {
	MessageID int64
	Outgoing  bool
	Header    string
	Body      string
	Meta      string
	Image     *ImageAttachment
}

var legacyTranscriptMessageLoader = func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
	backend := &UserBackend{cli: cli}
	return backend.GetMessages(ctx, target, limit)
}

func (cli *TelegramCLI) setLegacyConsole(console *legacyConsole) {
	cli.legacyMu.Lock()
	defer cli.legacyMu.Unlock()
	cli.legacyConsole = console
}

func (cli *TelegramCLI) currentLegacyConsole() *legacyConsole {
	cli.legacyMu.RLock()
	defer cli.legacyMu.RUnlock()
	return cli.legacyConsole
}

func normalizeLegacyTranscriptTarget(target string) string {
	return normalizeUsername(target)
}

func (cli *TelegramCLI) appendLegacyTranscriptEntry(target string, entry legacyTranscriptEntry) {
	normalized := normalizeLegacyTranscriptTarget(target)
	if normalized == "" {
		return
	}

	cli.legacyMu.Lock()
	defer cli.legacyMu.Unlock()

	existing := cli.legacyTranscripts[normalized]
	if entry.MessageID != 0 {
		for _, current := range existing {
			if current.MessageID == entry.MessageID {
				return
			}
		}
	}

	existing = append(existing, entry)
	if len(existing) > legacyTranscriptLimit {
		existing = existing[len(existing)-legacyTranscriptLimit:]
	}
	cli.legacyTranscripts[normalized] = existing
}

func (cli *TelegramCLI) legacyTranscriptSnapshot(target string) ([]legacyTranscriptEntry, bool) {
	normalized := normalizeLegacyTranscriptTarget(target)
	if normalized == "" {
		return nil, false
	}

	cli.legacyMu.RLock()
	defer cli.legacyMu.RUnlock()

	entries := append([]legacyTranscriptEntry(nil), cli.legacyTranscripts[normalized]...)
	return entries, cli.legacyLoaded[normalized]
}

func (cli *TelegramCLI) mergeLegacyTranscriptEntries(target string, fetched []legacyTranscriptEntry) {
	normalized := normalizeLegacyTranscriptTarget(target)
	if normalized == "" {
		return
	}

	cli.legacyMu.Lock()
	defer cli.legacyMu.Unlock()

	merged := mergeLegacyEntrySlices(fetched, cli.legacyTranscripts[normalized])
	if len(merged) > legacyTranscriptLimit {
		merged = merged[len(merged)-legacyTranscriptLimit:]
	}
	cli.legacyTranscripts[normalized] = merged
	cli.legacyLoaded[normalized] = true
}

func mergeLegacyEntrySlices(primary []legacyTranscriptEntry, secondary []legacyTranscriptEntry) []legacyTranscriptEntry {
	out := make([]legacyTranscriptEntry, 0, len(primary)+len(secondary))
	seenIDs := make(map[int64]struct{}, len(primary)+len(secondary))
	seenKeys := make(map[string]struct{}, len(primary)+len(secondary))

	appendUnique := func(entries []legacyTranscriptEntry) {
		for _, entry := range entries {
			if entry.MessageID != 0 {
				if _, exists := seenIDs[entry.MessageID]; exists {
					continue
				}
				seenIDs[entry.MessageID] = struct{}{}
				seenKeys[legacyEntryKey(entry)] = struct{}{}
			} else {
				key := legacyEntryKey(entry)
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

func legacyEntryKey(entry legacyTranscriptEntry) string {
	imageKey := ""
	if entry.Image != nil {
		imageKey = strings.TrimSpace(entry.Image.Kind) + "|" + strings.TrimSpace(entry.Image.Name) + "|" + strings.TrimSpace(entry.Image.MIMEType)
	}
	return fmt.Sprintf("%t|%s|%s|%s|%s", entry.Outgoing, strings.TrimSpace(entry.Header), strings.TrimSpace(entry.Body), strings.TrimSpace(entry.Meta), imageKey)
}

func (cli *TelegramCLI) ensureLegacyTranscript(ctx context.Context, target string, label string) error {
	return cli.ensureLegacyTranscriptContext(ctx, target, label, 1)
}

func (cli *TelegramCLI) ensureLegacyTranscriptContext(ctx context.Context, target string, label string, minEntries int) error {
	if ctx == nil {
		return nil
	}
	if minEntries < 1 {
		minEntries = 1
	}

	entries, loaded := cli.legacyTranscriptSnapshot(target)
	if loaded && len(entries) >= minEntries {
		return nil
	}

	messages, err := legacyTranscriptMessageLoader(ctx, cli, target, legacyTranscriptHistoryFetchLimit)
	if err != nil {
		return err
	}

	fetchedEntries := legacyEntriesFromMessages(target, label, messages)
	cli.mergeLegacyTranscriptEntries(target, fetchedEntries)
	return nil
}

func legacyEntriesFromMessages(target string, label string, messages []MessageOutput) []legacyTranscriptEntry {
	entries := make([]legacyTranscriptEntry, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		entries = append(entries, legacyTranscriptEntryFromMessageOutput(target, label, messages[i]))
	}
	return entries
}
