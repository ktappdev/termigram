package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// UserAuthenticator implements auth.UserAuthenticator
type UserAuthenticator struct {
	phone string
	cli   *TelegramCLI
}

func (ua *UserAuthenticator) Phone(ctx context.Context) (string, error) {
	if ua.phone == "" {
		ua.phone = ua.cli.promptInput("Enter phone number (with country code, e.g. +1234567890): ")
	}
	return ua.phone, nil
}

func (ua *UserAuthenticator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return ua.cli.promptInput("Enter authentication code: "), nil
}

func (ua *UserAuthenticator) Password(ctx context.Context) (string, error) {
	return ua.cli.promptInput("Enter 2FA password: "), nil
}

func (ua *UserAuthenticator) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (ua *UserAuthenticator) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign up not supported")
}

// handleAuthError maps an authentication error to a user-friendly message.
// It handles common Telegram RPC errors encountered during the auth flow.
func handleAuthError(err error) string {
	if err == nil {
		return "Authentication error: unknown"
	}
	var rpcErr *tgerr.Error
	if errors.As(err, &rpcErr) {
		switch rpcErr.Type {
		case "AUTH_KEY_UNREGISTERED":
			return "Not signed in. Please authenticate first."
		case "SESSION_PASSWORD_NEEDED":
			return "2FA password required."
		case tg.ErrPhoneCodeInvalid, tg.ErrCodeInvalid:
			return "Invalid code. Please try again."
		case tg.ErrPhoneNumberUnoccupied:
			return "Phone number not registered."
		default:
			return fmt.Sprintf("Authentication error: %s", rpcErr.Message)
		}
	}
	return fmt.Sprintf("Authentication error: %s", err.Error())
}
