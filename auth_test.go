package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gotd/td/tgerr"
)

// handleAuthError tests

func TestHandleAuthError_SessionNotFound(t *testing.T) {
	err := tgerr.New(401, "AUTH_KEY_UNREGISTERED")
	msg := handleAuthError(err)
	expected := "Not signed in. Please authenticate first."
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_PasswordRequired(t *testing.T) {
	err := tgerr.New(401, "SESSION_PASSWORD_NEEDED")
	msg := handleAuthError(err)
	expected := "2FA password required."
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_CodeInvalid(t *testing.T) {
	// PHONE_CODE_INVALID
	err := tgerr.New(400, "PHONE_CODE_INVALID")
	msg := handleAuthError(err)
	expected := "Invalid code. Please try again."
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_CodeInvalid_CodeInvalid(t *testing.T) {
	// CODE_INVALID is also mapped
	err := tgerr.New(400, "CODE_INVALID")
	msg := handleAuthError(err)
	expected := "Invalid code. Please try again."
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_PhoneUnoccupied(t *testing.T) {
	err := tgerr.New(400, "PHONE_NUMBER_UNOCCUPIED")
	msg := handleAuthError(err)
	expected := "Phone number not registered."
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_Unknown(t *testing.T) {
	// An RPC error we don't have a specific mapping for
	err := tgerr.New(420, "FLOOD_WAIT_60")
	msg := handleAuthError(err)
	expected := "Authentication error: FLOOD_WAIT_60"
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}

	// Non-RPC error
	err2 := fmt.Errorf("network timeout")
	msg2 := handleAuthError(err2)
	expected2 := "Authentication error: network timeout"
	if msg2 != expected2 {
		t.Fatalf("expected %q, got %q", expected2, msg2)
	}
}

func TestHandleAuthError_WrappedError(t *testing.T) {
	// Verify errors.As works through wrapping (e.g. from fmt.Errorf(": %w"))
	inner := tgerr.New(401, "AUTH_KEY_UNREGISTERED")
	wrapped := fmt.Errorf("sign in failed: %w", inner)
	msg := handleAuthError(wrapped)
	expected := "Not signed in. Please authenticate first."
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_NilError(t *testing.T) {
	msg := handleAuthError(nil)
	expected := "Authentication error: unknown"
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

func TestHandleAuthError_StandardError(t *testing.T) {
	// A plain Go error that is not a tgerr
	msg := handleAuthError(errors.New("something went wrong"))
	expected := "Authentication error: something went wrong"
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

// TestHandleAuthError_PhoneCodeExpired checks another common code error
func TestHandleAuthError_PhoneCodeExpired(t *testing.T) {
	// PHONE_CODE_EXPIRED should fall through to default
	err := tgerr.New(400, "PHONE_CODE_EXPIRED")
	msg := handleAuthError(err)
	expected := "Authentication error: PHONE_CODE_EXPIRED"
	if msg != expected {
		t.Fatalf("expected %q, got %q", expected, msg)
	}
}

// Note: The full auth flow (auth.NewFlow, UserAuthenticator.Phone/Code/Password)
// requires network calls to the Telegram API and is NOT unit-testable without mocking the
// gotd client or creating a TelegramCLI with a real reader. The handleAuthError function
// above is the only pure, testable function in the auth package.
