package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	interactiveResumeIdleThreshold = 15 * time.Minute
	interactiveResumeCooldown      = 15 * time.Second
	interactiveResumeSyncTimeout   = 20 * time.Second
	interactiveResumeDialogLimit   = 50
)

func (cli *TelegramCLI) interactiveResumeIdleLimit() time.Duration {
	if cli != nil && cli.config.InteractiveResumeIdleThreshold > 0 {
		return cli.config.InteractiveResumeIdleThreshold
	}
	return interactiveResumeIdleThreshold
}

func (cli *TelegramCLI) maybeResumeAfterIdle(ctx context.Context) {
	idle, shouldResume := cli.recordInteractiveUse(cli.nowFunc())
	if !shouldResume {
		return
	}

	if err := cli.resumeInteractiveSession(ctx, idle, false); err != nil {
		cli.writeOutput(fmt.Sprintf("[warn] Could not refresh Telegram after idle: %v", err))
	}
}

func (cli *TelegramCLI) recordInteractiveUse(now time.Time) (time.Duration, bool) {
	if now.IsZero() {
		now = cli.nowFunc()
	}

	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()

	previous := cli.chatState.lastInteractiveUse
	cli.chatState.lastInteractiveUse = now
	if previous.IsZero() {
		return 0, false
	}

	idle := now.Sub(previous)
	if idle < cli.interactiveResumeIdleLimit() {
		return idle, false
	}
	if !cli.chatState.lastResumeSync.IsZero() && now.Sub(cli.chatState.lastResumeSync) < interactiveResumeCooldown {
		return idle, false
	}

	cli.chatState.lastResumeSync = now
	return idle, true
}

func (cli *TelegramCLI) resumeInteractiveSession(ctx context.Context, idle time.Duration, force bool) error {
	if !force && idle < cli.interactiveResumeIdleLimit() {
		return nil
	}

	now := cli.nowFunc()
	if force {
		cli.markResumeSync(now)
	}

	resumeCtx, cancel := context.WithTimeout(ctx, interactiveResumeSyncTimeout)
	defer cancel()

	return cli.refreshInteractiveState(resumeCtx)
}

func (cli *TelegramCLI) markResumeSync(now time.Time) {
	if now.IsZero() {
		now = cli.nowFunc()
	}

	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	cli.chatState.lastResumeSync = now
}

func (cli *TelegramCLI) refreshInteractiveState(ctx context.Context) error {
	var errs []error

	target, label := cli.currentChat()
	if target != "" {
		if err := cli.transcriptStore.syncTranscriptContext(ctx, cli, target, label, transcriptResumeFetchLimit); err != nil {
			errs = append(errs, fmt.Errorf("refresh active chat: %w", err))
		} else {
			cli.clearChatUnreadCount(target)
		}
	}

	if _, err := cli.interactiveResumeDialogRefresher(ctx, cli, interactiveResumeDialogLimit); err != nil {
		errs = append(errs, fmt.Errorf("refresh dialogs: %w", err))
	}

	if target != "" && cli.transcriptStore.currentConsole() != nil {
		cli.redrawChatView()
	}

	return errors.Join(errs...)
}

func (cli *TelegramCLI) retryInteractiveRPC(ctx context.Context, fn func(context.Context) (int64, error)) (int64, error) {
	id, err := fn(ctx)
	if err == nil || ctx.Err() != nil || !isReconnectRetryable(err) {
		return id, err
	}

	resumeErr := cli.resumeInteractiveSession(ctx, cli.interactiveResumeIdleLimit(), true)
	retryID, retryErr := fn(ctx)
	if retryErr == nil {
		return retryID, nil
	}
	if resumeErr != nil {
		return retryID, fmt.Errorf("%w (resume refresh failed: %v)", retryErr, resumeErr)
	}
	return retryID, retryErr
}

func isReconnectRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}

	text := strings.ToLower(err.Error())
	retryableFragments := []string{
		"deadline exceeded",
		"timeout",
		"timed out",
		"eof",
		"broken pipe",
		"connection reset",
		"connection refused",
		"connection closed",
		"closed network connection",
		"transport",
		"reconnect",
		"rpc error code = unavailable",
	}
	for _, fragment := range retryableFragments {
		if strings.Contains(text, fragment) {
			return true
		}
	}

	return false
}
