package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type interactiveUIMode string

const (
	interactiveUIAuto   interactiveUIMode = "auto"
	interactiveUILegacy interactiveUIMode = "legacy"
)

func parseInteractiveUIMode(value string) (interactiveUIMode, error) {
	mode := interactiveUIMode(strings.ToLower(strings.TrimSpace(value)))
	if mode == "" {
		return interactiveUIAuto, nil
	}
	switch mode {
	case interactiveUIAuto, interactiveUILegacy:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported --ui value %q (expected auto or legacy)", value)
	}
}

func interactiveTTYAvailable() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func runInteractiveMode(cfg Config, uiValue string) error {
	mode, err := parseInteractiveUIMode(uiValue)
	if err != nil {
		return err
	}

	cli := NewTelegramCLI(cfg.TelegramAppID, cfg.TelegramAppHash, cfg.SessionPath)
	switch mode {
	case interactiveUILegacy, interactiveUIAuto:
		return cli.RunLegacy()
	default:
		return cli.RunLegacy()
	}
}
