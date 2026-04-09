package main

import (
	"os"

	"golang.org/x/term"
)

func interactiveTTYAvailable() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func runInteractiveMode(cfg Config) error {
	cli := NewTelegramCLI(cfg.TelegramAppID, cfg.TelegramAppHash, cfg.SessionPath)
	return cli.RunInteractive()
}
