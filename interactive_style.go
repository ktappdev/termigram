package main

import (
	"fmt"
	"os"
)

const (
	ansiReset       = "\033[0m"
	ansiBold        = "\033[1m"
	ansiDim         = "\033[2m"
	ansiRed         = "\033[31m"
	ansiGreen       = "\033[32m"
	ansiYellow      = "\033[33m"
	ansiBlue        = "\033[34m"
	ansiCyan        = "\033[36m"
	ansiWhite       = "\033[37m"
	ansiBgSoftBlue  = "\033[48;2;31;63;91m"
	ansiBgSoftGreen = "\033[48;2;29;58;42m"
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
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	if normalizeReplyTarget(cli.chatState.currentChatTarget) != normalizeReplyTarget(target) {
		cli.chatState.pendingReply = nil
		cli.chatState.pendingReplyTarget = ""
	}
	cli.chatState.currentChatTarget = target
	cli.chatState.currentChatLabel = label
}

func (cli *TelegramCLI) currentChat() (target string, label string) {
	cli.chatState.mu.RLock()
	defer cli.chatState.mu.RUnlock()
	return cli.chatState.currentChatTarget, cli.chatState.currentChatLabel
}

func (cli *TelegramCLI) promptLabel() string {
	_, label := cli.currentChat()
	if label == "" {
		return fmt.Sprintf("%s %s ", green("●"), bold(">"))
	}
	return fmt.Sprintf("%s %s %s%s%s %s ", green("●"), yellow("📍"), dim("["), cyan(label), dim("]"), bold(">"))
}
