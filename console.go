package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type console struct {
	fd       int
	terminal *term.Terminal
	oldState *term.State
}

func newConsole(prompt string) (*console, error) {
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

	console := &console{
		fd:       fd,
		terminal: terminal,
		oldState: oldState,
	}
	if err := console.Resize(); err != nil {
		fmt.Fprintf(os.Stderr, "termigram: initialize console size: %v\n", err)
	}
	return console, nil
}

func (c *console) Close() error {
	if c == nil {
		return nil
	}
	c.terminal.SetBracketedPasteMode(false)
	return term.Restore(c.fd, c.oldState)
}

func (c *console) ReadLine() (string, error) {
	return c.terminal.ReadLine()
}

func (c *console) SetPrompt(prompt string) {
	c.terminal.SetPrompt(prompt)
}

func (c *console) Resize() error {
	width, height, err := term.GetSize(c.fd)
	if err != nil {
		return err
	}
	return c.terminal.SetSize(width, height)
}

func (c *console) WriteString(text string) error {
	if text == "" {
		return nil
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	_, err := c.terminal.Write([]byte(text))
	return err
}

func (c *console) WriteBlock(text string) error {
	if text == "" {
		return nil
	}
	_, err := c.terminal.Write([]byte(text))
	return err
}
