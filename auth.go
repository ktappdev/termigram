package main

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
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
