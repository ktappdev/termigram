package main

import "fmt"

func (cli *TelegramCLI) setPendingReply(target string, ref *ReplyReference) {
	normalized := normalizeReplyTarget(target)
	cli.mu.Lock()
	defer cli.mu.Unlock()
	if normalized == "" || !isValidReplyReference(ref) {
		cli.pendingReply = nil
		cli.pendingReplyTarget = ""
		return
	}
	cli.pendingReplyTarget = normalized
	cli.pendingReply = cloneReplyReference(ref)
}

func (cli *TelegramCLI) pendingReplyForTarget(target string) *ReplyReference {
	normalized := normalizeReplyTarget(target)
	if normalized == "" {
		return nil
	}

	cli.mu.RLock()
	defer cli.mu.RUnlock()
	if cli.pendingReply == nil || cli.pendingReplyTarget != normalized {
		return nil
	}
	return cloneReplyReference(cli.pendingReply)
}

func (cli *TelegramCLI) clearPendingReply() bool {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cleared := cli.pendingReply != nil || cli.pendingReplyTarget != ""
	cli.pendingReply = nil
	cli.pendingReplyTarget = ""
	return cleared
}

func (cli *TelegramCLI) consumePendingReply(target string) *ReplyReference {
	normalized := normalizeReplyTarget(target)
	if normalized == "" {
		return nil
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	if cli.pendingReply == nil || cli.pendingReplyTarget != normalized {
		return nil
	}
	ref := *cli.pendingReply
	cli.pendingReply = nil
	cli.pendingReplyTarget = ""
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
