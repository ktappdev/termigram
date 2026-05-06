package main

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestProcessDialogsResponse_Success(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 3,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				TopMessage:  2,
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 1, Message: "Hello there", Date: 1700000000},
			&tg.Message{ID: 2, Message: "Hi back", Date: 1700000100},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Alice", LastName: "Smith", Username: "alice"},
			&tg.User{ID: 1002, FirstName: "Bob", Username: "bob"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 dialogs, got %d", len(out))
	}

	// First dialog: Alice
	if out[0].Label != "Alice Smith" {
		t.Fatalf("expected Label 'Alice Smith', got %q", out[0].Label)
	}
	if out[0].Target != "@alice" {
		t.Fatalf("expected Target '@alice', got %q", out[0].Target)
	}
	if out[0].UnreadCount != 3 {
		t.Fatalf("expected UnreadCount 3, got %d", out[0].UnreadCount)
	}
	if out[0].LastMessage != "Hello there" {
		t.Fatalf("expected LastMessage 'Hello there', got %q", out[0].LastMessage)
	}
	if out[0].LastActivity.Unix() != 1700000000 {
		t.Fatalf("expected LastActivity 1700000000, got %d", out[0].LastActivity.Unix())
	}

	// Second dialog: Bob
	if out[1].Label != "Bob" {
		t.Fatalf("expected Label 'Bob', got %q", out[1].Label)
	}
	if out[1].Target != "@bob" {
		t.Fatalf("expected Target '@bob', got %q", out[1].Target)
	}
	if out[1].UnreadCount != 0 {
		t.Fatalf("expected UnreadCount 0, got %d", out[1].UnreadCount)
	}

	// Users were cached
	u, found := cli.getUserByID(1001)
	if !found {
		t.Fatal("expected user 1001 to be cached")
	}
	if u.FirstName != "Alice" {
		t.Fatalf("expected cached user 1001 FirstName 'Alice', got %q", u.FirstName)
	}
}

func TestProcessDialogsResponse_EmptyDialogs(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs:  []tg.DialogClass{},
		Messages: []tg.MessageClass{},
		Users:    []tg.UserClass{},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error for empty dialogs, got: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(out) != 0 {
		t.Fatalf("expected 0 dialogs, got %d", len(out))
	}
}

func TestProcessDialogsResponse_UnreadOnly_FiltersRead(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 0, // read — should be filtered
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				TopMessage:  2,
				UnreadCount: 5, // unread — should be kept
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1003},
				TopMessage:  3,
				UnreadCount: 0, // read — should be filtered
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 1, Message: "read msg", Date: 1700000000},
			&tg.Message{ID: 2, Message: "unread msg", Date: 1700000100},
			&tg.Message{ID: 3, Message: "another read", Date: 1700000200},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Alice", Username: "alice"},
			&tg.User{ID: 1002, FirstName: "Bob", Username: "bob"},
			&tg.User{ID: 1003, FirstName: "Charlie", Username: "charlie"},
		},
	}

	out, err := processDialogsResponse(resp, cli, true)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 unread dialog, got %d", len(out))
	}
	if out[0].Target != "@bob" {
		t.Fatalf("expected target '@bob' (the unread dialog), got %q", out[0].Target)
	}
	if out[0].UnreadCount != 5 {
		t.Fatalf("expected UnreadCount 5, got %d", out[0].UnreadCount)
	}
}

func TestProcessDialogsResponse_UnreadOnly_KeepsAllWhenAllUnread(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				UnreadCount: 2,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				UnreadCount: 7,
			},
		},
		Messages: []tg.MessageClass{},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Alice", Username: "alice"},
			&tg.User{ID: 1002, FirstName: "Bob", Username: "bob"},
		},
	}

	out, err := processDialogsResponse(resp, cli, true)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 dialogs (both unread), got %d", len(out))
	}
}

func TestProcessDialogsResponse_AllDialogs_IncludesRead(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				UnreadCount: 0,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				UnreadCount: 3,
			},
		},
		Messages: []tg.MessageClass{},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Alice", Username: "alice"},
			&tg.User{ID: 1002, FirstName: "Bob", Username: "bob"},
		},
	}

	// unreadOnly=false — should include all dialogs
	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 dialogs (including read ones), got %d", len(out))
	}
}

func TestProcessDialogsResponse_UnsupportedType_ReturnsError(t *testing.T) {
	cli := createTestCLI()
	// MessagesDialogsNotModified does NOT implement dialogsEnvelope (no GetDialogs/GetMessages/GetUsers)
	resp := &tg.MessagesDialogsNotModified{Count: 0}

	out, err := processDialogsResponse(resp, cli, false)
	if err == nil {
		t.Fatal("expected error for unsupported response type, got nil")
	}
	if err.Error() != "unsupported dialogs response type: *tg.MessagesDialogsNotModified" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
	if out != nil {
		t.Fatalf("expected nil output on error, got: %v", out)
	}
}

func TestProcessDialogsResponse_DialogFolderSkipped(t *testing.T) {
	cli := createTestCLI()
	// DialogFolder is a DialogClass but not *tg.Dialog — should be silently skipped.
	// Only the *tg.Dialog with user 1002 should appear in output.
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.DialogFolder{
				Peer:       &tg.PeerUser{UserID: 1001},
				TopMessage: 0,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				TopMessage:  1,
				UnreadCount: 1,
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 1, Message: "actual dialog", Date: 1700000000},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "FolderUser", Username: "folder"},
			&tg.User{ID: 1002, FirstName: "RealUser", Username: "real"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog (DialogFolder skipped), got %d", len(out))
	}
	if out[0].Target != "@real" {
		t.Fatalf("expected target '@real', got %q", out[0].Target)
	}

	// DialogFolder's user should still be cached even though the dialog was skipped
	_, found := cli.getUserByID(1001)
	if !found {
		t.Fatal("expected user 1001 (from DialogFolder) to still be cached via cacheUsersFromClasses")
	}
}

func TestProcessDialogsResponse_NonUserPeerSkipped(t *testing.T) {
	cli := createTestCLI()
	// Dialogs with PeerChat and PeerChannel peers should be skipped (only PeerUser handled).
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerChat{ChatID: 2001},
				TopMessage:  1,
				UnreadCount: 2,
			},
			&tg.Dialog{
				Peer:        &tg.PeerChannel{ChannelID: 3001},
				TopMessage:  2,
				UnreadCount: 5,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  3,
				UnreadCount: 1,
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 3, Message: "user msg", Date: 1700000000},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Diana", Username: "diana"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog (PeerUser only), got %d", len(out))
	}
	if out[0].Target != "@diana" {
		t.Fatalf("expected target '@diana', got %q", out[0].Target)
	}
}

func TestProcessDialogsResponse_UserMissingFromCache_FallbackLabel(t *testing.T) {
	cli := createTestCLI()
	// User 1001 not in cache and not in Users list — fallback to "User N" label.
	// User 1002 IS in Users list and should be resolved normally.
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 9999},
				TopMessage:  1,
				UnreadCount: 0,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  2,
				UnreadCount: 2,
			},
		},
		Messages: []tg.MessageClass{},
		Users: []tg.UserClass{
			// Only user 1001 is in the response users
			&tg.User{ID: 1001, FirstName: "Eve", Username: "eve"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 dialogs, got %d", len(out))
	}

	// User 9999 not found — fallback label
	if out[0].Label != "User 9999" {
		t.Fatalf("expected fallback Label 'User 9999', got %q", out[0].Label)
	}
	if out[0].Target != "9999" {
		t.Fatalf("expected fallback Target '9999', got %q", out[0].Target)
	}

	// User 1001 found via cache
	if out[1].Label != "Eve" {
		t.Fatalf("expected Label 'Eve', got %q", out[1].Label)
	}
	if out[1].Target != "@eve" {
		t.Fatalf("expected Target '@eve', got %q", out[1].Target)
	}
}

func TestProcessDialogsResponse_TopMessageAttached(t *testing.T) {
	cli := createTestCLI()
	// Dialog references TopMessage=42 which exists in Messages map.
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  42,
				UnreadCount: 0,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				TopMessage:  99, // no message with ID 99 — should not set LastMessage/LastActivity
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 42, Message: "Last text here ", Date: 1700000123},
			// Note: trailing space in "Last text here " to test strings.TrimSpace
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Frank", Username: "frank"},
			&tg.User{ID: 1002, FirstName: "Grace", Username: "grace"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 dialogs, got %d", len(out))
	}

	// Frank: has top message 42
	if out[0].LastMessage != "Last text here" {
		t.Fatalf("expected LastMessage 'Last text here' (trimmed), got %q", out[0].LastMessage)
	}
	if out[0].LastActivity.Unix() != 1700000123 {
		t.Fatalf("expected LastActivity 1700000123, got %d", out[0].LastActivity.Unix())
	}

	// Grace: no message with ID 99
	if out[1].LastMessage != "" {
		t.Fatalf("expected empty LastMessage for unmatched TopMessage, got %q", out[1].LastMessage)
	}
	if !out[1].LastActivity.IsZero() {
		t.Fatalf("expected zero LastActivity for unmatched TopMessage, got %v", out[1].LastActivity)
	}
}

func TestProcessDialogsResponse_MessageClassNonMessageSkipped(t *testing.T) {
	cli := createTestCLI()
	// Messages list contains MessageEmpty and MessageService types that should be skipped
	// when building the messagesByID map (only *tg.Message is stored).
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{
			&tg.MessageEmpty{ID: 1},    // not *tg.Message — won't be in messagesByID
			// Note: MessageService is also a MessageClass but has no Message field.
			// The code only looks for *tg.Message, so this is fine.
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Helen", Username: "helen"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(out))
	}

	// Top message ID is 1, but only MessageEmpty with ID 1 exists — not a *tg.Message,
	// so LastMessage/LastActivity should remain empty.
	if out[0].LastMessage != "" {
		t.Fatalf("expected empty LastMessage (MessageEmpty skipped), got %q", out[0].LastMessage)
	}
}

func TestProcessDialogsResponse_SliceType(t *testing.T) {
	cli := createTestCLI()
	// MessagesDialogsSlice also implements dialogsEnvelope — should work identically.
	resp := &tg.MessagesDialogsSlice{
		Count: 1,
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 4,
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 1, Message: "from slice", Date: 1700000500},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Ivan", Username: "ivan"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error for MessagesDialogsSlice, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(out))
	}
	if out[0].Label != "Ivan" {
		t.Fatalf("expected Label 'Ivan', got %q", out[0].Label)
	}
	if out[0].Target != "@ivan" {
		t.Fatalf("expected Target '@ivan', got %q", out[0].Target)
	}
	if out[0].UnreadCount != 4 {
		t.Fatalf("expected UnreadCount 4, got %d", out[0].UnreadCount)
	}
	if out[0].LastMessage != "from slice" {
		t.Fatalf("expected LastMessage 'from slice', got %q", out[0].LastMessage)
	}
}

// TestProcessDialogsResponse_MessagesSliceAndDialogsSlice_SameOrder confirms that
// dialogs and their corresponding messages are correctly matched by ID, even when
// the order differs between Dialogs and Messages slices.
func TestProcessDialogsResponse_MessagesCrossReference(t *testing.T) {
	cli := createTestCLI()
	// Dialogs reference top messages by ID. The messages slice order differs from
	// the dialogs slice order — matching should be by ID, not position.
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  5,
				UnreadCount: 0,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				TopMessage:  3,
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{
			&tg.Message{ID: 3, Message: "third msg", Date: 1700000300},
			&tg.Message{ID: 5, Message: "fifth msg", Date: 1700000500},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Jack", Username: "jack"},
			&tg.User{ID: 1002, FirstName: "Kate", Username: "kate"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 dialogs, got %d", len(out))
	}

	// Jack has top message 5
	if out[0].LastMessage != "fifth msg" {
		t.Fatalf("expected Jack's LastMessage 'fifth msg', got %q", out[0].LastMessage)
	}
	if out[0].LastActivity.Unix() != 1700000500 {
		t.Fatalf("expected Jack's LastActivity 1700000500, got %d", out[0].LastActivity.Unix())
	}

	// Kate has top message 3
	if out[1].LastMessage != "third msg" {
		t.Fatalf("expected Kate's LastMessage 'third msg', got %q", out[1].LastMessage)
	}
	if out[1].LastActivity.Unix() != 1700000300 {
		t.Fatalf("expected Kate's LastActivity 1700000300, got %d", out[1].LastActivity.Unix())
	}
}

// TestProcessDialogsResponse_EmptyUsersList ensures that when the users list is empty,
// dialogs still get processed with fallback labels.
func TestProcessDialogsResponse_EmptyUsersList(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{},
		Users:    []tg.UserClass{}, // no users in response — user won't be cached
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(out))
	}

	// User not cached — fallback label
	if out[0].Label != "User 1001" {
		t.Fatalf("expected fallback Label 'User 1001', got %q", out[0].Label)
	}
	if out[0].Target != "1001" {
		t.Fatalf("expected fallback Target '1001', got %q", out[0].Target)
	}
}

// TestProcessDialogsResponse_NilResponse ensures a nil MessagesDialogsClass
// (which cannot be cast to dialogsEnvelope) returns an error.
func TestProcessDialogsResponse_NilResponse(t *testing.T) {
	cli := createTestCLI()

	out, err := processDialogsResponse(nil, cli, false)
	if err == nil {
		t.Fatal("expected error for nil response, got nil")
	}
	if err.Error() != "unsupported dialogs response type: <nil>" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
	if out != nil {
		t.Fatalf("expected nil output on error, got: %v", out)
	}
}

// TestProcessDialogsResponse_LastActivityZeroForMissingMessage confirms that
// when there's no message for a dialog's TopMessage, LastActivity is zero.
func TestProcessDialogsResponse_LastActivityZeroForMissingMessage(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{}, // no messages at all
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Leo", Username: "leo"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(out))
	}
	if !out[0].LastActivity.IsZero() {
		t.Fatalf("expected zero LastActivity when no message matches, got %v", out[0].LastActivity)
	}
}

// TestProcessDialogsResponse_UnreadOnlyWithNoUnreadDialogs returns empty when
// all dialogs are read and unreadOnly=true.
func TestProcessDialogsResponse_UnreadOnlyWithNoUnreadDialogs_ReturnsEmpty(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				UnreadCount: 0,
			},
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1002},
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Mike", Username: "mike"},
			&tg.User{ID: 1002, FirstName: "Nina", Username: "nina"},
		},
	}

	out, err := processDialogsResponse(resp, cli, true)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected 0 dialogs (all read), got %d", len(out))
	}
}

// TestProcessDialogsResponse_CachePersistsAcrossCalls verifies that users cached
// from one response are available for resolving dialogs in a subsequent response.
func TestProcessDialogsResponse_CachePersistsAcrossCalls(t *testing.T) {
	cli := createTestCLI()

	// First call: cache a user via a dialog setup
	resp1 := &tg.MessagesDialogs{
		Dialogs:  []tg.DialogClass{},
		Messages: []tg.MessageClass{},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "Oscar", Username: "oscar"},
		},
	}
	_, err := processDialogsResponse(resp1, cli, false)
	if err != nil {
		t.Fatalf("expected nil error on first call, got: %v", err)
	}

	// Second call: dialog referencing user 1001 should use cached user
	resp2 := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 3,
			},
		},
		Messages: []tg.MessageClass{},
		Users:    []tg.UserClass{}, // no users in this response
	}
	out, err := processDialogsResponse(resp2, cli, false)
	if err != nil {
		t.Fatalf("expected nil error on second call, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(out))
	}
	if out[0].Label != "Oscar" {
		t.Fatalf("expected Label 'Oscar' (from cache), got %q", out[0].Label)
	}
	if out[0].Target != "@oscar" {
		t.Fatalf("expected Target '@oscar' (from cache), got %q", out[0].Target)
	}
	if out[0].UnreadCount != 3 {
		t.Fatalf("expected UnreadCount 3, got %d", out[0].UnreadCount)
	}
}

// TestProcessDialogsResponse_UserWithOnlyUsername verifies that a user with
// only a username (no first/last name) gets the username as label.
func TestProcessDialogsResponse_UserWithOnlyUsername(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 1001},
				TopMessage:  1,
				UnreadCount: 0,
			},
		},
		Messages: []tg.MessageClass{},
		Users: []tg.UserClass{
			&tg.User{ID: 1001, FirstName: "", LastName: "", Username: "justauser"},
		},
	}

	out, err := processDialogsResponse(resp, cli, false)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(out))
	}
	// buildChatLabel falls back to Username when FirstName+LastName is empty
	if out[0].Label != "justauser" {
		t.Fatalf("expected Label 'justauser' (fallback to username), got %q", out[0].Label)
	}
	if out[0].Target != "@justauser" {
		t.Fatalf("expected Target '@justauser', got %q", out[0].Target)
	}
}


