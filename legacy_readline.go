package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

var errLegacyPromptInterrupted = errors.New("legacy prompt interrupted")

func (cli *TelegramCLI) readInteractiveLine(ctx context.Context, sigChan <-chan os.Signal) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		line, err := cli.reader.ReadString('\n')
		return strings.TrimSpace(line), err
	}

	console, err := newLegacyConsole(cli.promptLabel())
	if err != nil {
		line, readErr := cli.reader.ReadString('\n')
		if readErr != nil {
			return "", readErr
		}
		return strings.TrimSpace(line), nil
	}

	cli.setLegacyConsole(console)
	defer func() {
		cli.setLegacyConsole(nil)
		_ = console.Close()
	}()

	target, label := cli.currentChat()
	if target != "" {
		loadErr := cli.ensureLegacyTranscript(ctx, target, label)
		cli.redrawLegacyChatView()
		if loadErr != nil {
			_ = console.WriteString(fmt.Sprintf("Warning: could not load recent chat history: %v", loadErr))
		}
	}

	inputChan := make(chan inputReadResult, 1)
	go func() {
		line, err := console.ReadLine()
		inputChan <- inputReadResult{line: line, err: err}
	}()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case sig := <-sigChan:
			if sig == syscall.SIGWINCH {
				cli.handleLegacyResize()
				continue
			}
			return "", errLegacyPromptInterrupted
		case result := <-inputChan:
			if errors.Is(result.err, io.EOF) {
				return "", errLegacyPromptInterrupted
			}
			if result.err != nil {
				return "", result.err
			}
			return strings.TrimSpace(result.line), nil
		}
	}
}
