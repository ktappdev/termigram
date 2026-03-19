package main

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type legacyConsole struct {
	fd       int
	terminal *term.Terminal
	oldState *term.State
}

func newLegacyConsole(prompt string) (*legacyConsole, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	rw := struct {
		io.Reader
		io.Writer
	}{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}

	terminal := term.NewTerminal(rw, prompt)
	terminal.SetBracketedPasteMode(true)

	console := &legacyConsole{
		fd:       fd,
		terminal: terminal,
		oldState: oldState,
	}
	_ = console.Resize()
	return console, nil
}

func (c *legacyConsole) Close() error {
	if c == nil {
		return nil
	}
	c.terminal.SetBracketedPasteMode(false)
	return term.Restore(c.fd, c.oldState)
}

func (c *legacyConsole) ReadLine() (string, error) {
	return c.terminal.ReadLine()
}

func (c *legacyConsole) SetPrompt(prompt string) {
	c.terminal.SetPrompt(prompt)
}

func (c *legacyConsole) Resize() error {
	width, height, err := term.GetSize(c.fd)
	if err != nil {
		return err
	}
	return c.terminal.SetSize(width, height)
}

func (c *legacyConsole) WriteString(text string) error {
	if text == "" {
		return nil
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	_, err := c.terminal.Write([]byte(text))
	return err
}

func (c *legacyConsole) WriteBlock(text string) error {
	if text == "" {
		return nil
	}
	_, err := c.terminal.Write([]byte(text))
	return err
}
