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
	interactiveUITUI    interactiveUIMode = "tui"
	interactiveUILegacy interactiveUIMode = "legacy"
)

func parseInteractiveUIMode(value string) (interactiveUIMode, error) {
	mode := interactiveUIMode(strings.ToLower(strings.TrimSpace(value)))
	if mode == "" {
		return interactiveUIAuto, nil
	}
	switch mode {
	case interactiveUIAuto, interactiveUITUI, interactiveUILegacy:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported --ui value %q (expected auto, tui, or legacy)", value)
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
	case interactiveUILegacy:
		return cli.RunLegacy()
	case interactiveUITUI:
		if !interactiveTTYAvailable() {
			return fmt.Errorf("--ui=tui requires an interactive terminal")
		}
		return cli.RunTUI()
	default:
		return cli.RunLegacy()
	}
}
