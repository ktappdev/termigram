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

var errPromptInterrupted = errors.New("prompt interrupted")

func (cli *TelegramCLI) readInteractiveLine(ctx context.Context, sigChan <-chan os.Signal) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		line, err := cli.reader.ReadString('\n')
		return strings.TrimSpace(line), err
	}

	console, err := newConsole(cli.promptLabel())
	if err != nil {
		line, readErr := cli.reader.ReadString('\n')
		if readErr != nil {
			return "", readErr
		}
		return strings.TrimSpace(line), nil
	}

	cli.setConsole(console)
	defer func() {
		cli.setConsole(nil)
		_ = console.Close()
	}()

	target, label := cli.currentChat()
	if target != "" {
		loadErr := cli.ensureTranscript(ctx, target, label)
		cli.redrawChatView()
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
				cli.handleChatResize()
				continue
			}
			return "", errPromptInterrupted
		case result := <-inputChan:
			if errors.Is(result.err, io.EOF) {
				return "", errPromptInterrupted
			}
			if result.err != nil {
				return "", result.err
			}
			return strings.TrimSpace(result.line), nil
		}
	}
}
