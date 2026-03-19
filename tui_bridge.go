package main

import tea "github.com/charmbracelet/bubbletea"

func (cli *TelegramCLI) attachTUIProgram(program *tea.Program) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.tuiProgram = program
}

func (cli *TelegramCLI) detachTUIProgram() {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.tuiProgram = nil
}

func (cli *TelegramCLI) sendTUIMessage(msg tea.Msg) bool {
	cli.mu.RLock()
	program := cli.tuiProgram
	cli.mu.RUnlock()
	if program == nil {
		return false
	}
	program.Send(msg)
	return true
}
