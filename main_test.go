package main

import "testing"

func TestParseRootFlagsDefaultUI(t *testing.T) {
	remaining, mode, uiMode, handled, err := parseRootFlags([]string{"--mode", "user"})
	if err != nil {
		t.Fatalf("parseRootFlags returned error: %v", err)
	}
	if handled {
		t.Fatalf("expected handled=false for regular flag parsing")
	}
	if mode != "user" {
		t.Fatalf("expected mode=user, got %q", mode)
	}
	if uiMode != "auto" {
		t.Fatalf("expected default uiMode=auto, got %q", uiMode)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no remaining args, got %v", remaining)
	}
}

func TestParseInteractiveUIModeValidation(t *testing.T) {
	for _, candidate := range []string{"auto", "legacy", "AUTO"} {
		if _, err := parseInteractiveUIMode(candidate); err != nil {
			t.Fatalf("expected %q to be accepted, got error %v", candidate, err)
		}
	}
	for _, candidate := range []string{"tui", "broken"} {
		if _, err := parseInteractiveUIMode(candidate); err == nil {
			t.Fatalf("expected invalid ui mode %q to return an error", candidate)
		}
	}
}
