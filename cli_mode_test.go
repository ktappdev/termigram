package main

import (
	"context"
	"errors"
	"testing"
)

// mockBackend implements TelegramBackend and all optional sub-interfaces
// via function fields. Each test sets only the fields it needs.
type mockBackend struct {
	isAuthorizedFunc  func(ctx context.Context) (bool, error)
	getSelfFunc       func(ctx context.Context) (*UserOutput, error)
	sendMessageFunc   func(ctx context.Context, target string, text string, opts SendOptions) error
	sendImageFunc     func(ctx context.Context, target string, source string, caption string, opts SendOptions) error
	getMessagesFunc   func(ctx context.Context, target string, limit int) ([]MessageOutput, error)
	getContactsFunc   func(ctx context.Context) ([]ContactOutput, error)
	findUsernamesFunc func(prefix string, limit int) []string
	resolveTargetFunc func(ctx context.Context, target string) (*UserOutput, error)
}

// TelegramBackend
func (m *mockBackend) IsAuthorized(ctx context.Context) (bool, error) {
	if m.isAuthorizedFunc != nil {
		return m.isAuthorizedFunc(ctx)
	}
	return true, nil
}

func (m *mockBackend) GetSelf(ctx context.Context) (*UserOutput, error) {
	if m.getSelfFunc != nil {
		return m.getSelfFunc(ctx)
	}
	return &UserOutput{ID: 1, FirstName: "Test", LastName: "User", Username: "testuser"}, nil
}

func (m *mockBackend) SendMessage(ctx context.Context, target string, text string, opts SendOptions) error {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, target, text, opts)
	}
	return nil
}

// ImageSenderBackend
func (m *mockBackend) SendImage(ctx context.Context, target string, source string, caption string, opts SendOptions) error {
	if m.sendImageFunc != nil {
		return m.sendImageFunc(ctx, target, source, caption, opts)
	}
	return nil
}

// HistoryBackend
func (m *mockBackend) GetMessages(ctx context.Context, target string, limit int) ([]MessageOutput, error) {
	if m.getMessagesFunc != nil {
		return m.getMessagesFunc(ctx, target, limit)
	}
	return nil, nil
}

// ContactsBackend
func (m *mockBackend) GetContacts(ctx context.Context) ([]ContactOutput, error) {
	if m.getContactsFunc != nil {
		return m.getContactsFunc(ctx)
	}
	return nil, nil
}

// UsernameDiscoveryBackend
func (m *mockBackend) FindMatchingUsernames(prefix string, limit int) []string {
	if m.findUsernamesFunc != nil {
		return m.findUsernamesFunc(prefix, limit)
	}
	return nil
}

func (m *mockBackend) FindUserByUsername(ctx context.Context, username string) (*UserOutput, error) {
	return nil, nil
}

// TargetResolverBackend
func (m *mockBackend) ResolveTarget(ctx context.Context, target string) (*UserOutput, error) {
	if m.resolveTargetFunc != nil {
		return m.resolveTargetFunc(ctx, target)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// cmdMe
// ---------------------------------------------------------------------------

func TestCmdMe_Success(t *testing.T) {
	backend := &mockBackend{
		getSelfFunc: func(ctx context.Context) (*UserOutput, error) {
			return &UserOutput{
				ID: 123, FirstName: "Alice", LastName: "Smith",
				Username: "alice", Phone: "+1234567890",
			}, nil
		},
	}
	err := cmdMe(context.Background(), backend, false)
	if err != nil {
		t.Fatalf("cmdMe returned error: %v", err)
	}
}

func TestCmdMe_Error(t *testing.T) {
	expectedErr := errors.New("connection failed")
	backend := &mockBackend{
		getSelfFunc: func(ctx context.Context) (*UserOutput, error) {
			return nil, expectedErr
		},
	}
	err := cmdMe(context.Background(), backend, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error wrapping %v, got %v", expectedErr, err)
	}
}

func TestCmdMe_JSON(t *testing.T) {
	backend := &mockBackend{
		getSelfFunc: func(ctx context.Context) (*UserOutput, error) {
			return &UserOutput{
				ID: 456, FirstName: "Bob", LastName: "Jones",
				Username: "bobjones", Phone: "+9876543210",
			}, nil
		},
	}
	err := cmdMe(context.Background(), backend, true)
	if err != nil {
		t.Fatalf("cmdMe JSON returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// cmdContacts
// ---------------------------------------------------------------------------

func TestCmdContacts_Success(t *testing.T) {
	backend := &mockBackend{
		getContactsFunc: func(ctx context.Context) ([]ContactOutput, error) {
			return []ContactOutput{
				{UserID: 1, FirstName: "Alice", LastName: "A", Username: "alice"},
				{UserID: 2, FirstName: "Bob", LastName: "B", Username: "bob"},
			}, nil
		},
	}
	err := cmdContacts(context.Background(), backend, 10, 0, false)
	if err != nil {
		t.Fatalf("cmdContacts returned error: %v", err)
	}
}

func TestCmdContacts_Error(t *testing.T) {
	expectedErr := errors.New("no contacts backend")
	backend := &mockBackend{
		getContactsFunc: func(ctx context.Context) ([]ContactOutput, error) {
			return nil, expectedErr
		},
	}
	err := cmdContacts(context.Background(), backend, 10, 0, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error wrapping %v, got %v", expectedErr, err)
	}
}

// telegramOnlyBackend implements only TelegramBackend — no optional sub-interfaces.
type telegramOnlyBackend struct{}

func (b *telegramOnlyBackend) IsAuthorized(ctx context.Context) (bool, error) { return true, nil }
func (b *telegramOnlyBackend) GetSelf(ctx context.Context) (*UserOutput, error) { return nil, nil }
func (b *telegramOnlyBackend) SendMessage(ctx context.Context, _ string, _ string, _ SendOptions) error {
	return nil
}

func TestCmdContacts_BackendNotSupported(t *testing.T) {
	unsupported := &telegramOnlyBackend{}
	err := cmdContacts(context.Background(), unsupported, 10, 0, false)
	if err == nil {
		t.Fatal("expected error for non-ContactsBackend")
	}
	if err.Error() != "backend does not support contacts" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCmdContacts_EmptyList(t *testing.T) {
	backend := &mockBackend{
		getContactsFunc: func(ctx context.Context) ([]ContactOutput, error) {
			return []ContactOutput{}, nil
		},
	}
	err := cmdContacts(context.Background(), backend, 10, 0, false)
	if err != nil {
		t.Fatalf("cmdContacts with empty list returned error: %v", err)
	}
}

func TestCmdContacts_JSON(t *testing.T) {
	backend := &mockBackend{
		getContactsFunc: func(ctx context.Context) ([]ContactOutput, error) {
			return []ContactOutput{
				{UserID: 1, FirstName: "Alice", Username: "alice"},
			}, nil
		},
	}
	err := cmdContacts(context.Background(), backend, 10, 0, true)
	if err != nil {
		t.Fatalf("cmdContacts JSON returned error: %v", err)
	}
}

func TestCmdContacts_Pagination(t *testing.T) {
	contacts := make([]ContactOutput, 25)
	for i := range contacts {
		contacts[i] = ContactOutput{
			UserID:    int64(i + 1),
			FirstName: "User",
			LastName:  string(rune('A' + i%26)),
		}
	}
	backend := &mockBackend{
		getContactsFunc: func(ctx context.Context) ([]ContactOutput, error) {
			return contacts, nil
		},
	}

	// First page (default 20)
	err := cmdContacts(context.Background(), backend, 20, 0, false)
	if err != nil {
		t.Fatalf("cmdContacts page 1 returned error: %v", err)
	}

	// Second page (offset 20)
	err = cmdContacts(context.Background(), backend, 20, 20, false)
	if err != nil {
		t.Fatalf("cmdContacts page 2 returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// cmdFind
// ---------------------------------------------------------------------------

func TestCmdFind_NoArgs_ReturnsError(t *testing.T) {
	backend := &mockBackend{}
	err := cmdFind(context.Background(), backend, []string{}, false)
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
	if err.Error() != "usage: find <prefix>" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCmdFind_Success(t *testing.T) {
	backend := &mockBackend{
		findUsernamesFunc: func(prefix string, limit int) []string {
			return []string{"@alice", "@allen", "@albert"}
		},
	}
	err := cmdFind(context.Background(), backend, []string{"al"}, false)
	if err != nil {
		t.Fatalf("cmdFind returned error: %v", err)
	}
}

func TestCmdFind_Error(t *testing.T) {
	// telegramOnlyBackend does NOT implement UsernameDiscoveryBackend
	backend := &telegramOnlyBackend{}
	err := cmdFind(context.Background(), backend, []string{"al"}, false)
	if err == nil {
		t.Fatal("expected error for non-discovery backend, got nil")
	}
	if err.Error() != "backend does not support username prefix matching" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCmdFind_NoMatches(t *testing.T) {
	backend := &mockBackend{
		findUsernamesFunc: func(prefix string, limit int) []string {
			return nil
		},
	}
	err := cmdFind(context.Background(), backend, []string{"zzz"}, false)
	if err != nil {
		t.Fatalf("cmdFind with no matches returned error: %v", err)
	}
}

func TestCmdFind_JSON(t *testing.T) {
	backend := &mockBackend{
		findUsernamesFunc: func(prefix string, limit int) []string {
			return []string{"@alice"}
		},
	}
	err := cmdFind(context.Background(), backend, []string{"al"}, true)
	if err != nil {
		t.Fatalf("cmdFind JSON returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// parseSendOptions
// ---------------------------------------------------------------------------

func TestParseSendOptions_NoFlags(t *testing.T) {
	args := []string{"@alice", "hello"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ReplyToMessageID != 0 {
		t.Fatalf("expected ReplyToMessageID=0, got %d", opts.ReplyToMessageID)
	}
	if opts.Silent {
		t.Fatal("expected Silent=false")
	}
	if len(remaining) != 2 || remaining[0] != "@alice" || remaining[1] != "hello" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseSendOptions_ReplyTo(t *testing.T) {
	args := []string{"--reply-to", "42", "@alice", "hello"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ReplyToMessageID != 42 {
		t.Fatalf("expected ReplyToMessageID=42, got %d", opts.ReplyToMessageID)
	}
	if len(remaining) != 2 || remaining[0] != "@alice" || remaining[1] != "hello" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseSendOptions_Silent(t *testing.T) {
	args := []string{"--silent", "@bob", "quiet message"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.Silent {
		t.Fatal("expected Silent=true")
	}
	if len(remaining) != 2 || remaining[0] != "@bob" || remaining[1] != "quiet message" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseSendOptions_ReplyToAndSilent(t *testing.T) {
	args := []string{"--silent", "--reply-to", "7", "@charlie", "msg"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ReplyToMessageID != 7 {
		t.Fatalf("expected ReplyToMessageID=7, got %d", opts.ReplyToMessageID)
	}
	if !opts.Silent {
		t.Fatal("expected Silent=true")
	}
	if len(remaining) != 2 || remaining[0] != "@charlie" || remaining[1] != "msg" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseSendOptions_UnknownFlag(t *testing.T) {
	args := []string{"--unknown", "@alice", "hello"}
	_, _, err := parseSendOptions(args)
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
	if err.Error() != "unknown flag: --unknown" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestParseSendOptions_UnknownFlagAfterPositional(t *testing.T) {
	args := []string{"@alice", "--bogus", "hello"}
	_, _, err := parseSendOptions(args)
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
	if err.Error() != "unknown flag: --bogus" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestParseSendOptions_InvalidReplyToValue(t *testing.T) {
	args := []string{"--reply-to", "notanumber", "@alice"}
	_, _, err := parseSendOptions(args)
	if err == nil {
		t.Fatal("expected error for invalid --reply-to value, got nil")
	}
}

func TestParseSendOptions_MissingReplyToValue(t *testing.T) {
	// --reply-to at end of args without a value — should be caught gracefully
	args := []string{"--reply-to"}
	_, _, err := parseSendOptions(args)
	if err == nil {
		t.Fatal("expected error for missing --reply-to value, got nil")
	}
}

func TestParseSendOptions_PositionalArgsPreserved(t *testing.T) {
	args := []string{"--silent", "--reply-to", "99", "@dave", "multi", "word", "message"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ReplyToMessageID != 99 {
		t.Fatalf("expected ReplyToMessageID=99, got %d", opts.ReplyToMessageID)
	}
	if !opts.Silent {
		t.Fatal("expected Silent=true")
	}
	expected := []string{"@dave", "multi", "word", "message"}
	if len(remaining) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, remaining)
	}
	for i := range expected {
		if remaining[i] != expected[i] {
			t.Fatalf("remaining[%d]: expected %q, got %q", i, expected[i], remaining[i])
		}
	}
}

func TestParseSendOptions_ReplyToZeroValue(t *testing.T) {
	args := []string{"--reply-to", "0", "@eve", "hello"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0 is a valid integer; it means "no reply" — store it as-is
	if opts.ReplyToMessageID != 0 {
		t.Fatalf("expected ReplyToMessageID=0, got %d", opts.ReplyToMessageID)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 remaining args, got %d: %v", len(remaining), remaining)
	}
}

func TestParseSendOptions_OnlyFlags(t *testing.T) {
	args := []string{"--silent", "--reply-to", "5"}
	remaining, opts, err := parseSendOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ReplyToMessageID != 5 {
		t.Fatalf("expected ReplyToMessageID=5, got %d", opts.ReplyToMessageID)
	}
	if !opts.Silent {
		t.Fatal("expected Silent=true")
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no remaining args, got %d: %v", len(remaining), remaining)
	}
}

// ---------------------------------------------------------------------------
// RunCLICommand (dispatch)
// ---------------------------------------------------------------------------

func TestRunCLICommand_Unknown(t *testing.T) {
	cmd := CLICommand{Name: "nonexistent"}
	err := RunCLICommand(context.Background(), &mockBackend{}, cmd)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if err.Error() != "unknown command: nonexistent" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// chronologicalMessages
// ---------------------------------------------------------------------------

func TestChronologicalMessages(t *testing.T) {
	msgs := []MessageOutput{
		{ID: 2, Date: 200, FromName: "second"},
		{ID: 1, Date: 100, FromName: "first"},
		{ID: 3, Date: 100, FromName: "first-b"},
	}
	ordered := chronologicalMessages(msgs)
	if len(ordered) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(ordered))
	}
	// First two have same date, ordered by ID
	if ordered[0].ID != 1 || ordered[1].ID != 3 || ordered[2].ID != 2 {
		t.Fatalf("unexpected order: %+v", ordered)
	}
}

func TestChronologicalMessages_Empty(t *testing.T) {
	// Nil input produces nil output (append(nil, nil...) returns nil)
	result := chronologicalMessages(nil)
	if result != nil {
		t.Fatal("expected nil slice for nil input")
	}
	// Empty slice input also produces nil (same append behavior)
	result = chronologicalMessages([]MessageOutput{})
	if result != nil {
		t.Fatal("expected nil slice for empty slice input")
	}
}

func TestChronologicalMessages_AlreadyOrdered(t *testing.T) {
	msgs := []MessageOutput{
		{ID: 1, Date: 100, FromName: "alpha"},
		{ID: 2, Date: 200, FromName: "beta"},
	}
	result := chronologicalMessages(msgs)
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[0].ID != 1 || result[1].ID != 2 {
		t.Fatalf("order changed unexpectedly: %+v", result)
	}
}

// ---------------------------------------------------------------------------
// parseLimitArg
// ---------------------------------------------------------------------------

func TestParseLimitArg_NoLimit(t *testing.T) {
	args := []string{"@alice"}
	remaining, limit := parseLimitArg(args)
	if limit != 10 {
		t.Fatalf("expected default limit 10, got %d", limit)
	}
	if len(remaining) != 1 || remaining[0] != "@alice" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseLimitArg_WithLimit(t *testing.T) {
	args := []string{"--limit", "25", "@alice"}
	remaining, limit := parseLimitArg(args)
	if limit != 25 {
		t.Fatalf("expected limit 25, got %d", limit)
	}
	if len(remaining) != 1 || remaining[0] != "@alice" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseLimitArg_LimitAtEnd(t *testing.T) {
	args := []string{"@alice", "--limit", "5"}
	remaining, limit := parseLimitArg(args)
	if limit != 5 {
		t.Fatalf("expected limit 5, got %d", limit)
	}
	if len(remaining) != 1 || remaining[0] != "@alice" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}

func TestParseLimitArg_InvalidLimitIgnored(t *testing.T) {
	args := []string{"--limit", "notanumber", "@alice"}
	remaining, limit := parseLimitArg(args)
	if limit != 10 {
		t.Fatalf("expected default limit 10 when value invalid, got %d", limit)
	}
	// --limit and its arg are not removed when the value is invalid
	if len(remaining) != 3 {
		t.Fatalf("expected 3 remaining args (--limit, notanumber, @alice), got %d: %v", len(remaining), remaining)
	}
}

func TestParseLimitArg_ZeroLimitReturnsDefault(t *testing.T) {
	args := []string{"--limit", "0", "@alice"}
	remaining, limit := parseLimitArg(args)
	if limit != 10 {
		t.Fatalf("expected default limit 10 for limit=0, got %d", limit)
	}
	// 0 is valid parse but fails limit > 0 check, so --limit 0 stays in args
	if len(remaining) != 3 {
		t.Fatalf("expected 3 remaining args, got %d: %v", len(remaining), remaining)
	}
}

func TestParseLimitArg_NegativeLimitReturnsDefault(t *testing.T) {
	args := []string{"--limit", "-5", "@alice"}
	remaining, limit := parseLimitArg(args)
	if limit != 10 {
		t.Fatalf("expected default limit 10 for negative value, got %d", limit)
	}
	if len(remaining) != 3 {
		t.Fatalf("expected 3 remaining args, got %d: %v", len(remaining), remaining)
	}
}

func TestParseLimitArg_MultipleLimitsFirstWins(t *testing.T) {
	// parseLimitArg returns the first match it finds, removing just that pair
	args := []string{"--limit", "5", "--limit", "15", "@alice"}
	remaining, limit := parseLimitArg(args)
	if limit != 5 {
		t.Fatalf("expected limit 5 (first wins), got %d", limit)
	}
	// First --limit 5 removed, but --limit 15 remains
	if len(remaining) != 3 {
		t.Fatalf("expected 3 remaining args, got %d: %v", len(remaining), remaining)
	}
	if remaining[0] != "--limit" || remaining[1] != "15" || remaining[2] != "@alice" {
		t.Fatalf("unexpected remaining: %v", remaining)
	}
}
