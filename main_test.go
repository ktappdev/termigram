package main

import "testing"

func TestParseRootFlags(t *testing.T) {
	remaining, mode, handled, err := parseRootFlags([]string{"--mode", "user"})
	if err != nil {
		t.Fatalf("parseRootFlags returned error: %v", err)
	}
	if handled {
		t.Fatalf("expected handled=false for regular flag parsing")
	}
	if mode != "user" {
		t.Fatalf("expected mode=user, got %q", mode)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no remaining args, got %v", remaining)
	}
}

func TestParseRootFlagsRejectsDeprecatedUIFlag(t *testing.T) {
	if _, _, _, err := parseRootFlags([]string{"--ui", "legacy"}); err == nil {
		t.Fatalf("expected removed --ui flag to return an error")
	}
}

func TestIsCLICommandIncludesSendImage(t *testing.T) {
	if !isCLICommand("send-image") {
		t.Fatalf("expected send-image to be recognized as a CLI command")
	}
}
