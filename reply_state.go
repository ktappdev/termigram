package main

import "fmt"

func (cli *TelegramCLI) setPendingReply(target string, ref *ReplyReference) {
	normalized := normalizeReplyTarget(target)
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	if normalized == "" || !isValidReplyReference(ref) {
		cli.chatState.pendingReply = nil
		cli.chatState.pendingReplyTarget = ""
		return
	}
	cli.chatState.pendingReplyTarget = normalized
	cli.chatState.pendingReply = cloneReplyReference(ref)
}

func (cli *TelegramCLI) pendingReplyForTarget(target string) *ReplyReference {
	normalized := normalizeReplyTarget(target)
	if normalized == "" {
		return nil
	}

	cli.chatState.mu.RLock()
	defer cli.chatState.mu.RUnlock()
	if cli.chatState.pendingReply == nil || cli.chatState.pendingReplyTarget != normalized {
		return nil
	}
	return cloneReplyReference(cli.chatState.pendingReply)
}

func (cli *TelegramCLI) clearPendingReply() bool {
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	cleared := cli.chatState.pendingReply != nil || cli.chatState.pendingReplyTarget != ""
	cli.chatState.pendingReply = nil
	cli.chatState.pendingReplyTarget = ""
	return cleared
}

func (cli *TelegramCLI) consumePendingReply(target string) *ReplyReference {
	normalized := normalizeReplyTarget(target)
	if normalized == "" {
		return nil
	}

	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	if cli.chatState.pendingReply == nil || cli.chatState.pendingReplyTarget != normalized {
		return nil
	}
	ref := *cli.chatState.pendingReply
	cli.chatState.pendingReply = nil
	cli.chatState.pendingReplyTarget = ""
	return &ref
}

func (cli *TelegramCLI) pendingReplyBanner(target string) string {
	ref := cli.pendingReplyForTarget(target)
	if !isValidReplyReference(ref) {
		return ""
	}
	if preview := normalizeReplyPreview(ref.Preview); preview != "" {
		if sender := normalizeReplyPreview(ref.Sender); sender != "" {
			return fmt.Sprintf("Replying to %s: %s", sender, preview)
		}
		return fmt.Sprintf("Replying to message #%d: %s", ref.MessageID, preview)
	}
	if sender := normalizeReplyPreview(ref.Sender); sender != "" {
		return fmt.Sprintf("Replying to %s", sender)
	}
	return fmt.Sprintf("Replying to message #%d", ref.MessageID)
}

func pendingReplySummary(ref *ReplyReference) string {
	if !isValidReplyReference(ref) {
		return ""
	}
	if preview := normalizeReplyPreview(ref.Preview); preview != "" {
		if sender := normalizeReplyPreview(ref.Sender); sender != "" {
			return fmt.Sprintf("%s: %s", sender, preview)
		}
		return fmt.Sprintf("Message #%d: %s", ref.MessageID, preview)
	}
	if sender := normalizeReplyPreview(ref.Sender); sender != "" {
		return sender
	}
	return fmt.Sprintf("Message #%d", ref.MessageID)
}
