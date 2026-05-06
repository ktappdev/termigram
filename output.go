package main

import (
	"fmt"
	"os"
	"strings"
)

func (cli *TelegramCLI) writeOutput(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}

	if console := cli.transcriptStore.currentConsole(); console != nil {
		if err := console.WriteString(text); err != nil {
			fmt.Fprintf(os.Stderr, "termigram: write output: %v\n", err)
		}
		return
	}

	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	fmt.Print(strings.ReplaceAll(text, "\n", "\r\n"))
}
