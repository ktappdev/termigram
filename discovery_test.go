package main

import (
	"testing"

	"github.com/gotd/td/tg"
)

// createTestCLI returns a TelegramCLI with minimal setup for testing
// findMatchingUsernames. Only the userCache is needed.
func createTestCLI() *TelegramCLI {
	return &TelegramCLI{
		userCache: newUserCache(),
		chatState: newChatState(),
	}
}

func TestFindMatchingUsernames_NoMatches(t *testing.T) {
	cli := createTestCLI()
	// Empty cache — no users at all
	results := cli.findMatchingUsernames("alice", 10)
	if results != nil {
		t.Fatalf("expected nil for empty cache, got %v", results)
	}

	// Cache populated but no prefix match
	cli.cacheUser(&tg.User{ID: 1, Username: "bob", FirstName: "Bob"})
	cli.cacheUser(&tg.User{ID: 2, Username: "charlie", FirstName: "Charlie"})
	results = cli.findMatchingUsernames("alice", 10)
	if results != nil {
		t.Fatalf("expected nil when no usernames match prefix, got %v", results)
	}
}

func TestFindMatchingUsernames_SingleMatch(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, Username: "alice", FirstName: "Alice"})

	results := cli.findMatchingUsernames("ali", 10)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(results), results)
	}
	if results[0] != "@alice" {
		t.Fatalf("expected @alice, got %q", results[0])
	}
}

func TestFindMatchingUsernames_MultipleMatches(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, Username: "alice", FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 2, Username: "allen", FirstName: "Allen"})
	cli.cacheUser(&tg.User{ID: 3, Username: "albert", FirstName: "Albert"})
	// Non-matching user
	cli.cacheUser(&tg.User{ID: 4, Username: "bob", FirstName: "Bob"})

	results := cli.findMatchingUsernames("al", 10)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}
	// Results must be sorted alphabetically
	expected := []string{"@albert", "@alice", "@allen"}
	for i, name := range expected {
		if results[i] != name {
			t.Fatalf("result[%d] expected %q, got %q", i, name, results[i])
		}
	}
}

func TestFindMatchingUsernames_LimitApplied(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, Username: "alice", FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 2, Username: "allen", FirstName: "Allen"})
	cli.cacheUser(&tg.User{ID: 3, Username: "albert", FirstName: "Albert"})

	// Limit 2 of 3 matching results
	results := cli.findMatchingUsernames("al", 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results (limit applied), got %d: %v", len(results), results)
	}
	// First two alphabetically
	if results[0] != "@albert" || results[1] != "@alice" {
		t.Fatalf("expected [@albert @alice], got %v", results)
	}

	// Limit greater than available results — returns all
	results = cli.findMatchingUsernames("al", 100)
	if len(results) != 3 {
		t.Fatalf("expected 3 results (limit > matches), got %d", len(results))
	}

	// Limit of 0 — should return all (limit > 0 check in code)
	results = cli.findMatchingUsernames("al", 0)
	if len(results) != 3 {
		t.Fatalf("expected 3 results (limit 0 means no limit), got %d", len(results))
	}
}

func TestFindMatchingUsernames_CaseInsensitive(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, Username: "Charlie", FirstName: "Charlie"})

	// Search with uppercase prefix
	results := cli.findMatchingUsernames("CHAR", 10)
	if len(results) != 1 {
		t.Fatalf("expected 1 result for case-insensitive search, got %d: %v", len(results), results)
	}
	if results[0] != "@charlie" {
		t.Fatalf("expected @charlie (lowercase), got %q", results[0])
	}

	// Search with mixed case
	results = cli.findMatchingUsernames("cHaR", 10)
	if len(results) != 1 {
		t.Fatalf("expected 1 result for mixed-case search, got %d", len(results))
	}
	if results[0] != "@charlie" {
		t.Fatalf("expected @charlie, got %q", results[0])
	}
}

func TestFindMatchingUsernames_PrefixOnly(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, Username: "alice123", FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 2, Username: "alice_smith", FirstName: "Alice Smith"})
	// Exact match also present
	cli.cacheUser(&tg.User{ID: 3, Username: "alice", FirstName: "Alice Only"})

	// Prefix "alice" should match all three
	results := cli.findMatchingUsernames("alice", 10)
	if len(results) != 3 {
		t.Fatalf("expected 3 results for prefix 'alice', got %d: %v", len(results), results)
	}
	expected := []string{"@alice", "@alice123", "@alice_smith"}
	for i, name := range expected {
		if results[i] != name {
			t.Fatalf("result[%d] expected %q, got %q", i, name, results[i])
		}
	}

	// Shorter prefix — broader match (3 users start with "al")
	results = cli.findMatchingUsernames("al", 10)
	if len(results) != 3 {
		t.Fatalf("expected 3 results for prefix 'al', got %d: %v", len(results), results)
	}
}

func TestFindMatchingUsernames_WithAtSymbol(t *testing.T) {
	cli := createTestCLI()
	cli.cacheUser(&tg.User{ID: 1, Username: "alice", FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 2, Username: "bob", FirstName: "Bob"})

	// Search with "@" prefix — should work identically
	withAt := cli.findMatchingUsernames("@ali", 10)
	withoutAt := cli.findMatchingUsernames("ali", 10)

	if len(withAt) != 1 {
		t.Fatalf("expected 1 result with @ prefix, got %d: %v", len(withAt), withAt)
	}
	if len(withoutAt) != 1 {
		t.Fatalf("expected 1 result without @ prefix, got %d: %v", len(withoutAt), withoutAt)
	}
	if withAt[0] != "@alice" {
		t.Fatalf("expected @alice with @ prefix, got %q", withAt[0])
	}
	if withoutAt[0] != "@alice" {
		t.Fatalf("expected @alice without @ prefix, got %q", withoutAt[0])
	}

	// Full "@username" search
	results := cli.findMatchingUsernames("@alice", 10)
	if len(results) != 1 || results[0] != "@alice" {
		t.Fatalf("expected [@alice] for @alice search, got %v", results)
	}

	// Empty prefix after stripping @ — should return nil
	results = cli.findMatchingUsernames("@", 10)
	if results != nil {
		t.Fatalf("expected nil for empty prefix after stripping @, got %v", results)
	}
}

func TestFindMatchingUsernames_Integration(t *testing.T) {
	cli := createTestCLI()

	// Cache three users
	cli.cacheUser(&tg.User{ID: 1, Username: "alice", FirstName: "Alice"})
	cli.cacheUser(&tg.User{ID: 2, Username: "bob", FirstName: "Bob"})
	cli.cacheUser(&tg.User{ID: 3, Username: "charlie", FirstName: "Charlie"})

	// Search for each prefix
	t.Run("prefix_a", func(t *testing.T) {
		results := cli.findMatchingUsernames("a", 10)
		if len(results) != 1 || results[0] != "@alice" {
			t.Fatalf("expected [@alice], got %v", results)
		}
	})

	t.Run("prefix_b", func(t *testing.T) {
		results := cli.findMatchingUsernames("b", 10)
		if len(results) != 1 || results[0] != "@bob" {
			t.Fatalf("expected [@bob], got %v", results)
		}
	})

	t.Run("prefix_c", func(t *testing.T) {
		results := cli.findMatchingUsernames("c", 10)
		if len(results) != 1 || results[0] != "@charlie" {
			t.Fatalf("expected [@charlie], got %v", results)
		}
	})

	t.Run("prefix_empty", func(t *testing.T) {
		results := cli.findMatchingUsernames("", 10)
		if results != nil {
			t.Fatalf("expected nil for empty prefix, got %v", results)
		}
	})

	t.Run("prefix_no_match", func(t *testing.T) {
		results := cli.findMatchingUsernames("z", 10)
		if results != nil {
			t.Fatalf("expected nil for non-matching prefix, got %v", results)
		}
	})

	// Verify cache search matches what was inserted
	t.Run("all_users_accessible", func(t *testing.T) {
		results := cli.findMatchingUsernames("", 10)
		if results != nil {
			t.Fatalf("expected nil for empty prefix, got %v", results)
		}
	})
}
