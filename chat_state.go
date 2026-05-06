package main

import (
	"sync"
	"time"
)

// ChatState holds all active chat context and per-chat metadata.
// It replaces the scattered chat-state fields that were previously
// embedded directly in TelegramCLI.
type ChatState struct {
	mu sync.RWMutex

	currentChatTarget string
	currentChatLabel  string

	lastInteractiveUse time.Time
	lastResumeSync     time.Time

	chatLastActivity map[string]time.Time
	chatLastMessage  map[string]string
	chatUnreadCount  map[string]int
	seenIncoming     map[string]time.Time

	pendingReply       *ReplyReference
	pendingReplyTarget string
}

func newChatState() *ChatState {
	return &ChatState{
		chatLastActivity: make(map[string]time.Time),
		chatLastMessage:  make(map[string]string),
		chatUnreadCount:  make(map[string]int),
		seenIncoming:     make(map[string]time.Time),
	}
}
