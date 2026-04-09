package main

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestRenderChatViewFitsTerminalBounds(t *testing.T) {
	entries := []transcriptEntry{
		{
			MessageID: 1,
			Header:    "Ken's Butler",
			Body:      "This is a long incoming message that should wrap cleanly inside the chat view when the terminal width changes.",
			Meta:      "07:28:10",
		},
		{
			Header:   "You",
			Body:     "This is a test message that should stay visible once without overflowing the render width.",
			Meta:     "07:29:48  ✓",
			Outgoing: true,
		},
	}

	for _, width := range []int{120, 90, 60, 40, 24} {
		view := renderChatView("Ken's Butler", "@Ken592Bot", entries, width, 20)
		for _, line := range strings.Split(view, "\n") {
			if got := runewidth.StringWidth(stripANSI(line)); got > width {
				t.Fatalf("rendered line width %d exceeded width %d: %q", got, width, line)
			}
		}
	}
}

func TestMergeTranscriptEntrySlicesDedupesRepeatedMessages(t *testing.T) {
	fetched := []transcriptEntry{
		{MessageID: 10, Header: "Ken", Body: "hello", Meta: "07:28:10"},
	}
	buffered := []transcriptEntry{
		{MessageID: 10, Header: "Ken", Body: "hello", Meta: "07:28:10"},
		{Header: "You", Body: "draft send", Meta: "07:29:48  ✓", Outgoing: true},
		{Header: "You", Body: "draft send", Meta: "07:29:48  ✓", Outgoing: true},
	}

	merged := mergeTranscriptEntrySlices(fetched, buffered)
	if got := len(merged); got != 2 {
		t.Fatalf("expected 2 merged entries after dedupe, got %d", got)
	}
}
