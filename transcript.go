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

func defaultTranscriptMessageLoader(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error) {
	backend := &UserBackend{cli: cli}
	return backend.GetMessages(ctx, target, limit)
}

func normalizeTranscriptTarget(target string) string {
	return normalizeUsername(target)
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

func transcriptEntriesFromMessages(target string, label string, messages []MessageOutput) []transcriptEntry {
	entries := make([]transcriptEntry, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		entries = append(entries, transcriptEntryFromMessageOutput(target, label, messages[i]))
	}
	return entries
}
