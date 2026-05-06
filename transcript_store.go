package main

import (
	"context"
	"strings"
	"sync"
)

type TranscriptStore struct {
	mu                      sync.RWMutex
	console                 *console
	transcripts             map[string][]transcriptEntry
	transcriptLoaded        map[string]bool
	config                  *Config
	transcriptMessageLoader func(ctx context.Context, cli *TelegramCLI, target string, limit int) ([]MessageOutput, error)
}

func newTranscriptStore(config *Config) *TranscriptStore {
	return &TranscriptStore{
		transcripts:             make(map[string][]transcriptEntry),
		transcriptLoaded:        make(map[string]bool),
		config:                  config,
		transcriptMessageLoader: defaultTranscriptMessageLoader,
	}
}

func (s *TranscriptStore) setConsole(console *console) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.console = console
}

func (s *TranscriptStore) currentConsole() *console {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.console
}

func (s *TranscriptStore) transcriptEntryLimit() int {
	if s != nil && s.config != nil && s.config.TranscriptLimit > 0 {
		return s.config.TranscriptLimit
	}
	return transcriptLimit
}

func (s *TranscriptStore) transcriptHistoryLimit() int {
	if s != nil && s.config != nil && s.config.TranscriptHistoryFetchLimit > 0 {
		return s.config.TranscriptHistoryFetchLimit
	}
	return transcriptHistoryFetchLimit
}

func (s *TranscriptStore) appendTranscriptEntry(target string, entry transcriptEntry) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" || s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.transcripts[normalized]
	if entry.MessageID != 0 {
		for _, current := range existing {
			if current.MessageID == entry.MessageID {
				return
			}
		}
	}

	entryLimit := s.transcriptEntryLimit()
	existing = append(existing, entry)
	if len(existing) > entryLimit {
		existing = existing[len(existing)-entryLimit:]
	}
	s.transcripts[normalized] = existing
}

func (s *TranscriptStore) transcriptSnapshot(target string) ([]transcriptEntry, bool) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" || s == nil {
		return nil, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := append([]transcriptEntry(nil), s.transcripts[normalized]...)
	return entries, s.transcriptLoaded[normalized]
}

func (s *TranscriptStore) mergeTranscriptEntries(target string, fetched []transcriptEntry) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" || s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entryLimit := s.transcriptEntryLimit()
	merged := mergeTranscriptEntrySlices(fetched, s.transcripts[normalized])
	if len(merged) > entryLimit {
		merged = merged[len(merged)-entryLimit:]
	}
	s.transcripts[normalized] = merged
	s.transcriptLoaded[normalized] = true
}

func (s *TranscriptStore) ensureTranscript(ctx context.Context, cli *TelegramCLI, target string, label string) error {
	return s.ensureTranscriptContext(ctx, cli, target, label, 1)
}

func (s *TranscriptStore) ensureTranscriptContext(ctx context.Context, cli *TelegramCLI, target string, label string, minEntries int) error {
	if ctx == nil {
		return nil
	}
	if minEntries < 1 {
		minEntries = 1
	}

	entries, loaded := s.transcriptSnapshot(target)
	if loaded && len(entries) >= minEntries {
		return nil
	}

	return s.syncTranscriptContext(ctx, cli, target, label, s.transcriptHistoryLimit())
}

func (s *TranscriptStore) syncTranscriptContext(ctx context.Context, cli *TelegramCLI, target string, label string, limit int) error {
	if ctx == nil {
		return nil
	}
	if limit <= 0 {
		limit = s.transcriptHistoryLimit()
	}

	messages, err := s.transcriptMessageLoader(ctx, cli, target, limit)
	if err != nil {
		return err
	}

	fetchedEntries := transcriptEntriesFromMessages(target, label, messages)
	s.mergeTranscriptEntries(target, fetchedEntries)
	return nil
}

func (s *TranscriptStore) updateImageCachePath(target string, messageID int64, path string) {
	normalized := normalizeTranscriptTarget(target)
	if normalized == "" || messageID == 0 || strings.TrimSpace(path) == "" || s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.transcripts[normalized]
	for i := range entries {
		if entries[i].MessageID == messageID && entries[i].Image != nil {
			entries[i].Image.CachedPath = path
		}
	}
	s.transcripts[normalized] = entries
}
