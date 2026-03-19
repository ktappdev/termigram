package ui

// IncomingMessageMsg is sent by the host application when a live Telegram
// update arrives while the Bubble Tea UI is running.
type IncomingMessageMsg struct {
	Target    string
	ChatTitle string
	Message   BackendMessage
}
