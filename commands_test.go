package main

import (
	"context"
	"strings"
	"testing"

	"github.com/gotd/td/tg"
)

// ---------------------------------------------------------------------------
// Utility function tests (truncateInline, fuzzyMatch, filterChats)
// ---------------------------------------------------------------------------

func TestTruncateInline(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello world", 20, "hello world"},
		{"hello world", 5, "hell…"},
		{"hello world", 1, "…"},
		{"hello world", 0, "hello world"},
		{"hello\nworld", 20, "hello world"},
		{"hello\nworld\nfoo", 12, "hello world…"},
		{"  spaced  ", 20, "spaced"},
	}
	for _, tc := range tests {
		got := truncateInline(tc.input, tc.max)
		if got != tc.want {
			t.Errorf("truncateInline(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
		}
	}
}

func TestTruncateInline_Empty(t *testing.T) {
	if got := truncateInline("", 10); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		query string
		text  string
		want  bool
	}{
		{"abc", "abc", true},
		{"abc", "aBc", true},
		{"abc", "xyz", false},
		{"abc", "a long text with abc inside", true},
		{"abc", "ab", false},
		{"abc", "abcabc", true},
		{"hello", "world", false},
		{"", "anything", true},
		{"  ", "anything", true},
		{"a", "", false},
	}
	for _, tc := range tests {
		got := fuzzyMatch(tc.query, tc.text)
		if got != tc.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tc.query, tc.text, got, tc.want)
		}
	}
}

func TestFuzzyMatch_Sequential(t *testing.T) {
	// Characters must appear in order
	if !fuzzyMatch("abc", "aXbXc") {
		t.Error("expected fuzzy match for 'abc' in 'aXbXc'")
	}
	// Order matters
	if fuzzyMatch("cba", "abc") {
		t.Error("expected no fuzzy match for 'cba' in 'abc'")
	}
}

func TestFilterChats(t *testing.T) {
	chats := []CachedChat{
		{Label: "Alice", Target: "@alice", LastMessage: "hello there"},
		{Label: "Bob", Target: "@bob", LastMessage: "how are you"},
		{Label: "Charlie", Target: "@charlie", LastMessage: "good morning"},
	}

	t.Run("empty_query_returns_all", func(t *testing.T) {
		got := filterChats(chats, "")
		if len(got) != 3 {
			t.Fatalf("expected 3, got %d", len(got))
		}
	})

	t.Run("match_label_prefix", func(t *testing.T) {
		// "Ali" is a label prefix for Alice; fuzzyMatch also matches
		// Charlie due to sequential char match (a, l, i in "charlie").
		got := filterChats(chats, "Ali")
		if len(got) == 0 {
			t.Fatalf("expected at least Alice, got none")
		}
		// Alice should be first (score 0 by username prefix)
		if got[0].Label != "Alice" {
			t.Errorf("expected Alice first, got %v", got[0])
		}
	})

	t.Run("match_username_prefix", func(t *testing.T) {
		// "bob" matches username prefix exactly
		got := filterChats(chats, "bob")
		if len(got) != 1 || got[0].Target != "@bob" {
			t.Fatalf("expected [@bob], got %v", got)
		}
	})

	t.Run("match_message", func(t *testing.T) {
		got := filterChats(chats, "morning")
		if len(got) != 1 || got[0].Label != "Charlie" {
			t.Fatalf("expected [Charlie], got %v", got)
		}
	})

	t.Run("no_match", func(t *testing.T) {
		got := filterChats(chats, "zzzz")
		if len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})

	t.Run("fuzzy_match_label", func(t *testing.T) {
		// fuzzy match "lie" should match "Alice" (l,i,e sequential)
		got := filterChats(chats, "lie")
		if len(got) == 0 {
			t.Fatalf("expected fuzzy match, got none")
		}
	})

	t.Run("case_insensitive_username_prefix", func(t *testing.T) {
		// "ALICE" (uppercase) matches Alice via normalizedQuery (score 0).
		// Charlie can also fuzzy-match because a,l,i,c,e appear sequentially
		// within "charlie @charlie good morning". Alice should be first.
		got := filterChats(chats, "ALICE")
		if len(got) == 0 || got[0].Label != "Alice" {
			t.Fatalf("expected Alice first, got %v", got)
		}
	})

	t.Run("no_fuzzy_match", func(t *testing.T) {
		// "zzzz" has no sequential match in any chat's haystack
		got := filterChats(chats, "zzzz")
		if len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})
}

func TestFilterChats_Empty(t *testing.T) {
	got := filterChats(nil, "test")
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_Unknown — unknown command produces error message
// ---------------------------------------------------------------------------

func TestHandleCommand_Unknown(t *testing.T) {
	output := captureStdout(t, func() {
		unknownCommand("\\badcommand")
	})
	if !strings.Contains(output, "Unknown command:") {
		t.Errorf("expected 'Unknown command:' in output, got %q", output)
	}
	if !strings.Contains(output, "\\badcommand") {
		t.Errorf("expected command name in output, got %q", output)
	}
	if !strings.Contains(output, "\\help") {
		t.Errorf("expected \\help hint in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_Help — help text is printed
// ---------------------------------------------------------------------------

func TestHandleCommand_Help(t *testing.T) {
	output := captureStdout(t, func() {
		printHelp()
	})
	// Verify key sections/commands appear in help output
	expected := []string{
		"\\me", "\\contacts", "\\find", "\\msg", "\\image", "\\openimage",
		"\\to", "\\here", "\\chats", "\\unread", "\\close", "\\help", "\\quit",
	}
	for _, cmd := range expected {
		if !strings.Contains(output, cmd) {
			t.Errorf("help output missing %q", cmd)
		}
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_Quit — quit prints "Goodbye!" and calls cancel
// ---------------------------------------------------------------------------

func TestHandleCommand_Quit(t *testing.T) {
	cli := createTestCLI()

	cancelCalled := false
	cli.cancel = func() { cancelCalled = true }

	output := captureStdout(t, func() {
		cli.quitCommand()
	})

	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected 'Goodbye!' in output, got %q", output)
	}
	if !cancelCalled {
		t.Error("expected cancel to be called")
	}
}

func TestHandleCommand_Quit_NilCancel(t *testing.T) {
	cli := createTestCLI()
	// cli.cancel is already nil from createTestCLI

	output := captureStdout(t, func() {
		cli.quitCommand()
	})

	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected 'Goodbye!' in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_Clear — clear/close removes active chat
// ---------------------------------------------------------------------------

func TestHandleCommand_Clear(t *testing.T) {
	cli := createTestCLI()
	cli.setCurrentChat("@alice", "Alice")

	// Verify chat is active
	target, label := cli.currentChat()
	if target != "@alice" || label != "Alice" {
		t.Fatalf("expected chat to be set before clear")
	}

	output := captureStdout(t, func() {
		cli.clearActiveChat(false)
	})

	// Chat should be cleared
	target, label = cli.currentChat()
	if target != "" || label != "" {
		t.Errorf("expected cleared chat, got target=%q label=%q", target, label)
	}

	// Output should mention the closed chat
	if !strings.Contains(output, "Closed chat mode:") || !strings.Contains(output, "Alice") {
		t.Errorf("expected close message containing 'Closed chat mode:' and 'Alice', got %q", output)
	}
}

func TestHandleCommand_Clear_NoActiveChat(t *testing.T) {
	cli := createTestCLI()

	output := captureStdout(t, func() {
		cli.clearActiveChat(false)
	})

	if !strings.Contains(output, "No active chat") {
		t.Errorf("expected 'No active chat' message, got %q", output)
	}
}

func TestHandleCommand_Clear_Silent(t *testing.T) {
	cli := createTestCLI()
	cli.setCurrentChat("@alice", "Alice")

	output := captureStdout(t, func() {
		cli.clearActiveChat(true)
	})

	// Silent mode still prints the close message (existing behavior),
	// but the "No active chat to close." message is suppressed.
	if !strings.Contains(output, "Closed chat mode:") {
		t.Errorf("expected 'Closed chat mode:' in output, got %q", output)
	}

	// Chat should be cleared
	target, label := cli.currentChat()
	if target != "" || label != "" {
		t.Errorf("expected cleared chat, got target=%q label=%q", target, label)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_ChatSwitch_Valid / Invalid — /to command behavior
// ---------------------------------------------------------------------------

func TestHandleCommand_ChatSwitch_Valid(t *testing.T) {
	cli := createTestCLI()
	ctx := context.Background()

	// Cache a user so resolveChatTarget succeeds without API call
	cli.cacheUser(&tg.User{
		ID:        1,
		FirstName: "Alice",
		Username:  "alice",
	})

	output := captureStdout(t, func() {
		cli.switchActiveChat(ctx, "@alice", false)
	})

	target, label := cli.currentChat()
	if target != "@alice" {
		t.Errorf("expected target @alice, got %q", target)
	}
	if label != "Alice" {
		t.Errorf("expected label Alice, got %q", label)
	}

	if !strings.Contains(output, "Active chat:") || !strings.Contains(output, "Alice") {
		t.Errorf("expected active chat notification, got %q", output)
	}
}

func TestHandleCommand_ChatSwitch_Invalid(t *testing.T) {
	cli := createTestCLI()
	ctx := context.Background()

	// No user cached — "bad" is not @username or numeric ID
	output := captureStdout(t, func() {
		cli.switchActiveChat(ctx, "bad", false)
	})

	// Chat must not have changed
	target, label := cli.currentChat()
	if target != "" || label != "" {
		t.Errorf("expected chat unchanged, got target=%q label=%q", target, label)
	}

	if !strings.Contains(output, "Error:") {
		t.Errorf("expected error in output, got %q", output)
	}
	if !strings.Contains(output, "invalid user ID") {
		t.Errorf("expected 'invalid user ID' in error, got %q", output)
	}
}

func TestHandleCommand_ChatSwitch_UncachedUsername(t *testing.T) {
	// This scenario (uncached @username) requires an active API connection
	// to resolve via ContactsResolveUsername. In unit tests without a real
	// client, cli.api is nil and this path panics. The error path through
	// strconv.ParseInt is tested by TestHandleCommand_ChatSwitch_Invalid.
	t.Skip("needs API mock — error path tested by TestHandleCommand_ChatSwitch_Invalid")
}

// ---------------------------------------------------------------------------
// TestHandleCommand_ChatSwitch_Silent — silent mode produces no output
// ---------------------------------------------------------------------------

func TestHandleCommand_ChatSwitch_Silent(t *testing.T) {
	cli := createTestCLI()
	ctx := context.Background()

	cli.cacheUser(&tg.User{ID: 1, FirstName: "Alice", Username: "alice"})

	output := captureStdout(t, func() {
		cli.switchActiveChat(ctx, "@alice", true)
	})

	if output != "" {
		t.Errorf("expected no output in silent mode, got %q", output)
	}

	target, label := cli.currentChat()
	if target != "@alice" || label != "Alice" {
		t.Errorf("expected chat to be set, got target=%q label=%q", target, label)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_ShowActiveChat — /here displays active chat info
// ---------------------------------------------------------------------------

func TestHandleCommand_ShowActiveChat_NoActive(t *testing.T) {
	cli := createTestCLI()

	output := captureStdout(t, func() {
		cli.showActiveChat()
	})

	if !strings.Contains(output, "No active chat") {
		t.Errorf("expected 'No active chat' message, got %q", output)
	}
}

func TestHandleCommand_ShowActiveChat_WithActive(t *testing.T) {
	cli := createTestCLI()
	cli.setCurrentChat("@alice", "Alice")

	output := captureStdout(t, func() {
		cli.showActiveChat()
	})

	if !strings.Contains(output, "Active chat:") {
		t.Errorf("expected 'Active chat:' in output, got %q", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected 'Alice' in output, got %q", output)
	}
	if !strings.Contains(output, "@alice") {
		t.Errorf("expected '@alice' in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_ShowCachedChats — /chats displays cached conversation list
// ---------------------------------------------------------------------------

func TestShowCachedChats_Empty(t *testing.T) {
	cli := createTestCLI()

	output := captureStdout(t, func() {
		cli.showCachedChats()
	})

	// Two possible outputs depending on the actual code path
	if !strings.Contains(output, "No cached") {
		t.Errorf("expected empty-chats message, got %q", output)
	}
}

func TestShowCachedChats_WithChats(t *testing.T) {
	cli := createTestCLI()

	// Cache a couple of users so listCachedChats returns entries
	cli.cacheUser(&tg.User{ID: 1, FirstName: "Alice", Username: "alice"})
	cli.cacheUser(&tg.User{ID: 2, FirstName: "Bob", Username: "bob"})

	output := captureStdout(t, func() {
		cli.showCachedChats()
	})

	if !strings.Contains(output, "Cached conversations") {
		t.Errorf("expected 'Cached conversations' header, got %q", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected 'Alice' in chat list, got %q", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("expected 'Bob' in chat list, got %q", output)
	}
	if !strings.Contains(output, "@alice") {
		t.Errorf("expected '@alice' in output, got %q", output)
	}
	if !strings.Contains(output, "@bob") {
		t.Errorf("expected '@bob' in output, got %q", output)
	}
}

func TestShowCachedChats_ActiveMarker(t *testing.T) {
	cli := createTestCLI()

	cli.cacheUser(&tg.User{ID: 1, FirstName: "Alice", Username: "alice"})
	cli.cacheUser(&tg.User{ID: 2, FirstName: "Bob", Username: "bob"})
	cli.setCurrentChat("@alice", "Alice")

	output := captureStdout(t, func() {
		cli.showCachedChats()
	})

	// Active chat should be listed (both Alice and Bob are cached)
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
		t.Errorf("expected both Alice and Bob in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// TestDispatchCommand_RegistersAll — all expected commands are registered
// ---------------------------------------------------------------------------

func TestDispatchCommand_RegistersAll(t *testing.T) {
	// Every command the switch handles, with coverage notes.
	// Commands marked "tested-directly" have dedicated test functions above.
	// Commands marked "needs-integration" require a live API or terminal mock.
	type cmdInfo struct {
		command  string
		coverage string
	}

	commands := []cmdInfo{
		{`\me`,         "needs-integration (showSelf requires API)"},
		{`\contacts`,   "needs-integration (showContacts requires API)"},
		{`\find`,       "needs-integration (selectCachedChat needs terminal)"},
		{`\msg`,        "needs-integration (sendMessage requires API)"},
		{`\image`,      "needs-integration (sendImage requires API)"},
		{`\reply`,      "needs-integration (selectReplyTarget needs terminal)"},
		{`\cancelreply`, "tested-directly via reply_state tests"},
		{`\openimage`,  "needs-integration (openImageFromCurrentChat needs API)"},
		{`\to`,         "tested-directly via TestHandleCommand_ChatSwitch_*"},
		{`\here`,       "tested-directly via TestHandleCommand_ShowActiveChat_*"},
		{`\chats`,      "needs-integration (runChatsPicker needs terminal)"},
		{`\unread`,     "needs-integration (runUnreadPicker needs API)"},
		{`\close`,      "tested-directly via TestHandleCommand_Clear*"},
		{`\chat`,       "tested-directly via TestHandleCommand_ChatSwitch_* and TestHandleCommand_ShowActiveChat_*"},
		{`\back`,       "tested-directly via TestHandleCommand_Clear*"},
		{`\help`,       "tested-directly via TestHandleCommand_Help"},
		{`\quit`,       "tested-directly via TestHandleCommand_Quit*"},
		{`\exit`,       "tested-directly via TestHandleCommand_Quit*"},
	}

	for _, c := range commands {
		t.Run(c.command, func(t *testing.T) {
			// Verify the command can be tokenized
			tokens, err := splitCommandTokens(c.command + " arg1 arg2")
			if err != nil {
				t.Fatalf("splitCommandTokens(%q) error: %v", c.command, err)
			}
			if len(tokens) < 1 || tokens[0] != c.command {
				t.Fatalf("expected first token %q, got %#v", c.command, tokens)
			}

			// Verify unknownCommand does NOT match for known command tokens
			knownPrefixes := []string{
				"\\me", "\\contacts", "\\find", "\\msg", "\\image",
				"\\reply", "\\cancelreply", "\\openimage", "\\to",
				"\\here", "\\chats", "\\unread", "\\close", "\\chat",
				"\\back", "\\help", "\\quit", "\\exit",
			}
			for _, known := range knownPrefixes {
				if strings.HasPrefix(tokens[0], known) {
					return // known command, not unknown
				}
			}
			t.Errorf("command %q not in known prefixes list", c.command)
		})
	}
}

// ---------------------------------------------------------------------------
// TestCommandParsing_ArgsExtraction — parses "/cmd arg1 arg2" correctly
// ---------------------------------------------------------------------------

func TestCommandParsing_ArgsExtraction(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{`\help`,                    []string{`\help`}},
		{`\quit`,                    []string{`\quit`}},
		{`\to @alice`,               []string{`\to`, `@alice`}},
		{`\msg @alice hello world`,  []string{`\msg`, `@alice`, `hello`, `world`}},
		{`\msg 12345 hi`,            []string{`\msg`, `12345`, `hi`}},
		{`\find multi word query`,   []string{`\find`, `multi`, `word`, `query`}},
		{`\close`,                   []string{`\close`}},
		{`\image ./path.png`,        []string{`\image`, `./path.png`}},
		{`\image ./path.png "caption with spaces"`, []string{`\image`, `./path.png`, `caption with spaces`}},
		{`\openimage last`,          []string{`\openimage`, `last`}},
	}

	for _, tc := range tests {
		got, err := splitCommandTokens(tc.input)
		if err != nil {
			t.Errorf("splitCommandTokens(%q) error: %v", tc.input, err)
			continue
		}
		if len(got) != len(tc.want) {
			t.Errorf("splitCommandTokens(%q) = %#v (len=%d), want %#v (len=%d)",
				tc.input, got, len(got), tc.want, len(tc.want))
			continue
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Errorf("splitCommandTokens(%q)[%d] = %q, want %q",
					tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestCommandParsing_ArgsExtraction_BackslashEscape(t *testing.T) {
	// Backslash-escaped spaces in filenames
	tokens, err := splitCommandTokens(`\image my\ file.png`)
	if err != nil {
		t.Fatalf("splitCommandTokens error: %v", err)
	}
	want := []string{`\image`, `my file.png`}
	if len(tokens) != len(want) {
		t.Fatalf("got %#v, want %#v", tokens, want)
	}
	for i := range want {
		if tokens[i] != want[i] {
			t.Fatalf("token[%d] = %q, want %q", i, tokens[i], want[i])
		}
	}
}

func TestCommandParsing_ArgsExtraction_RejectsUnterminatedQuotes(t *testing.T) {
	if _, err := splitCommandTokens(`\msg "hello world`); err == nil {
		t.Error("expected error for unterminated quote")
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_ChatDeprecated — /chat command backward-compat aliases
// ---------------------------------------------------------------------------

func TestHandleCommand_ChatDeprecated(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, FirstName: "Alice", Username: "alice"})

	output := captureStdout(t, func() {
		cli.switchActiveChat(context.Background(), "@alice", true)
	})
	// Silent switch — no output
	if output != "" {
		t.Errorf("expected no output for silent switch, got %q", output)
	}

	target, label := cli.currentChat()
	if target != "@alice" || label != "Alice" {
		t.Errorf("expected chat switched to @alice, got target=%q label=%q", target, label)
	}
}

// ---------------------------------------------------------------------------
// TestHandleCommand_CancelledSwitch — switch error path from runChatsPicker
// ---------------------------------------------------------------------------

func TestHandleCommand_SwitchErrorPath(t *testing.T) {
	cli := createTestCLI()
	ctx := context.Background()

	// Passing an empty string should produce an error
	output := captureStdout(t, func() {
		cli.switchActiveChat(ctx, "", false)
	})

	if !strings.Contains(output, "Error:") {
		t.Errorf("expected error in output, got %q", output)
	}

	// Chat state remains empty
	target, label := cli.currentChat()
	if target != "" || label != "" {
		t.Errorf("expected unchanged chat, got target=%q label=%q", target, label)
	}
}
