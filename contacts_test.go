package main

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestProcessContactsResponse_ContactsNotModified(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.ContactsContactsNotModified{}

	out, err := processContactsResponse(resp, cli)
	if err != nil {
		t.Fatalf("expected nil error for ContactsNotModified, got: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output for ContactsNotModified, got: %v", out)
	}
}

func TestProcessContactsResponse_ContactsWithUsers(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.ContactsContacts{
		Users: []tg.UserClass{
			&tg.User{
				ID:        1001,
				FirstName: "Alice",
				LastName:  "Smith",
				Username:  "alice",
				Phone:     "12345",
			},
			&tg.User{
				ID:        1002,
				FirstName: "Bob",
				LastName:  "Jones",
				Username:  "bob",
				Phone:     "67890",
			},
		},
	}

	out, err := processContactsResponse(resp, cli)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 contacts, got %d: %v", len(out), out)
	}

	// Verify first contact output
	if out[0].UserID != 1001 {
		t.Fatalf("expected UserID 1001, got %d", out[0].UserID)
	}
	if out[0].FirstName != "Alice" {
		t.Fatalf("expected FirstName Alice, got %q", out[0].FirstName)
	}
	if out[0].LastName != "Smith" {
		t.Fatalf("expected LastName Smith, got %q", out[0].LastName)
	}
	if out[0].Username != "alice" {
		t.Fatalf("expected Username alice, got %q", out[0].Username)
	}
	if out[0].Phone != "12345" {
		t.Fatalf("expected Phone 12345, got %q", out[0].Phone)
	}

	// Verify second contact
	if out[1].UserID != 1002 {
		t.Fatalf("expected UserID 1002, got %d", out[1].UserID)
	}
	if out[1].Username != "bob" {
		t.Fatalf("expected Username bob, got %q", out[1].Username)
	}

	// Verify users were cached
	u, found := cli.getUserByID(1001)
	if !found {
		t.Fatal("expected user 1001 to be cached")
	}
	if u.FirstName != "Alice" {
		t.Fatalf("expected cached user 1001 FirstName Alice, got %q", u.FirstName)
	}

	u, found = cli.getUserByID(1002)
	if !found {
		t.Fatal("expected user 1002 to be cached")
	}
	if u.FirstName != "Bob" {
		t.Fatalf("expected cached user 1002 FirstName Bob, got %q", u.FirstName)
	}
}

func TestProcessContactsResponse_EmptyContactsList(t *testing.T) {
	cli := createTestCLI()
	resp := &tg.ContactsContacts{
		Users: []tg.UserClass{},
	}

	out, err := processContactsResponse(resp, cli)
	if err != nil {
		t.Fatalf("expected nil error for empty contacts, got: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil empty slice for empty contacts, got nil")
	}
	if len(out) != 0 {
		t.Fatalf("expected 0 contacts, got %d", len(out))
	}
}

func TestProcessContactsResponse_UnexpectedType(t *testing.T) {
	cli := createTestCLI()
	// nil is not a valid *tg.ContactsContacts or *tg.ContactsContactsNotModified
	var resp tg.ContactsContactsClass = nil

	out, err := processContactsResponse(resp, cli)
	if err == nil {
		t.Fatal("expected error for nil response, got nil")
	}
	if err.Error() != "unexpected contacts response type: <nil>" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
	if out != nil {
		t.Fatalf("expected nil output on error, got: %v", out)
	}
}

func TestProcessContactsResponse_NonUserClassInUsers(t *testing.T) {
	cli := createTestCLI()
	// Include a UserClass that is not *tg.User — should be silently skipped.
	resp := &tg.ContactsContacts{
		Users: []tg.UserClass{
			&tg.User{
				ID:        2001,
				FirstName: "Charlie",
				Username:  "charlie",
			},
			// UserEmpty is a valid UserClass that is not *tg.User
			&tg.UserEmpty{ID: 9999},
			&tg.User{
				ID:        2002,
				FirstName: "Diana",
				Username:  "diana",
			},
		},
	}

	out, err := processContactsResponse(resp, cli)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 contacts (skipping non-*tg.User), got %d: %v", len(out), out)
	}
	if out[0].UserID != 2001 {
		t.Fatalf("expected first contact UserID 2001, got %d", out[0].UserID)
	}
	if out[1].UserID != 2002 {
		t.Fatalf("expected second contact UserID 2002, got %d", out[1].UserID)
	}

	// UserEmpty should NOT be cached; only actual *tg.User entries are cached.
	if _, found := cli.getUserByID(9999); found {
		t.Fatal("UserEmpty (ID 9999) should not have been cached")
	}
}
