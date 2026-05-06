package main

import (
	"os"

	"golang.org/x/term"
)

func interactiveTTYAvailable() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func runInteractiveMode(cfg Config) error {
	cli := NewTelegramCLIWithConfig(cfg)
	return cli.RunInteractive()
}
