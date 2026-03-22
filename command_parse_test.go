package main

import "testing"

func TestSplitCommandTokensSupportsQuotesAndEscapes(t *testing.T) {
	tokens, err := splitCommandTokens(`\image /Users/kentaylor/Pictures/my\ meme.png "hello there"`)
	if err != nil {
		t.Fatalf("splitCommandTokens returned error: %v", err)
	}
	want := []string{`\image`, `/Users/kentaylor/Pictures/my meme.png`, `hello there`}
	if len(tokens) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %#v", len(want), len(tokens), tokens)
	}
	for i := range want {
		if tokens[i] != want[i] {
			t.Fatalf("token %d mismatch: got %q want %q", i, tokens[i], want[i])
		}
	}
}

func TestSplitCommandTokensRejectsUnterminatedQuotes(t *testing.T) {
	if _, err := splitCommandTokens(`\image "unterminated`); err == nil {
		t.Fatalf("expected unterminated quote error")
	}
}
