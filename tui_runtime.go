package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/ktappdev/termigram/ui"
)

func (cli *TelegramCLI) RunTUI() error {
	ctx, cancel := context.WithCancel(context.Background())
	cli.ctx = ctx
	cli.cancel = cancel

	return cli.client.Run(ctx, func(ctx context.Context) error {
		cli.api = cli.client.API()
		cli.sender = message.NewSender(cli.client.API())

		status, err := cli.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		if !status.Authorized {
			fmt.Println("Not authenticated. Starting auth flow...")
			authFlow := auth.NewFlow(&UserAuthenticator{cli: cli}, auth.SendCodeOptions{})
			if err := cli.client.Auth().IfNecessary(ctx, authFlow); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}
		}

		self, err := cli.client.Self(ctx)
		if err != nil {
			return fmt.Errorf("failed to get self: %w", err)
		}
		cli.cacheUser(self)

		if _, err := cli.fetchContacts(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load contacts for TUI: %v\n", err)
		}
		if _, err := cli.fetchDialogs(ctx, 50, false); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not preload dialogs for TUI: %v\n", err)
		}

		backend := &UserBackend{cli: cli}
		model := ui.NewModel(ctx, newUIBackendAdapter(backend))
		program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
		cli.attachTUIProgram(program)
		defer cli.detachTUIProgram()

		if _, err := program.Run(); err != nil {
			return fmt.Errorf("bubble tea UI failed: %w", err)
		}
		return nil
	})
}
