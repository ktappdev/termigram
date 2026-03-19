package main

import (
	"fmt"
	"strings"
)

func (cli *TelegramCLI) writeLegacyOutput(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}

	if console := cli.currentLegacyConsole(); console != nil {
		_ = console.WriteString(text)
		return
	}

	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	fmt.Print(strings.ReplaceAll(text, "\n", "\r\n"))
}
