package main

import (
	"fmt"
	"os"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
)

var ansiEnabled = os.Getenv("NO_COLOR") == ""

func colorize(style string, text string) string {
	if !ansiEnabled {
		return text
	}
	return style + text + ansiReset
}

func bold(text string) string   { return colorize(ansiBold, text) }
func dim(text string) string    { return colorize(ansiDim, text) }
func red(text string) string    { return colorize(ansiRed, text) }
func green(text string) string  { return colorize(ansiGreen, text) }
func yellow(text string) string { return colorize(ansiYellow, text) }
func blue(text string) string   { return colorize(ansiBlue, text) }
func cyan(text string) string   { return colorize(ansiCyan, text) }

func (cli *TelegramCLI) setCurrentChat(target string, label string) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.currentChatTarget = target
	cli.currentChatLabel = label
}

func (cli *TelegramCLI) currentChat() (target string, label string) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()
	return cli.currentChatTarget, cli.currentChatLabel
}

func (cli *TelegramCLI) promptLabel() string {
	_, label := cli.currentChat()
	if label == "" {
		return fmt.Sprintf("%s %s ", green("●"), bold(">"))
	}
	return fmt.Sprintf("%s %s %s%s%s %s ", green("●"), yellow("📍"), dim("["), cyan(label), dim("]"), bold(">"))
}
