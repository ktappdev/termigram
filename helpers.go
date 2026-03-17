package main

import (
	"fmt"
	"strings"
)

func (cli *TelegramCLI) promptInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := cli.reader.ReadString('\n')
	return strings.TrimSpace(input)
}
