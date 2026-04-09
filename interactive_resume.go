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

var interactiveResumeNow = time.Now
var interactiveResumeDialogRefresher = func(ctx context.Context, cli *TelegramCLI, limit int) ([]CachedChat, error) {
	return cli.fetchDialogs(ctx, limit, false)
}

func (cli *TelegramCLI) maybeResumeAfterIdle(ctx context.Context) {
	idle, shouldResume := cli.recordInteractiveUse(interactiveResumeNow())
	if !shouldResume {
		return
	}

	if err := cli.resumeInteractiveSession(ctx, idle, false); err != nil {
		cli.writeOutput(fmt.Sprintf("[warn] Could not refresh Telegram after idle: %v", err))
	}
}

func (cli *TelegramCLI) recordInteractiveUse(now time.Time) (time.Duration, bool) {
	if now.IsZero() {
		now = interactiveResumeNow()
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()

	previous := cli.lastInteractiveUse
	cli.lastInteractiveUse = now
	if previous.IsZero() {
		return 0, false
	}

	idle := now.Sub(previous)
	if idle < interactiveResumeIdleThreshold {
		return idle, false
	}
	if !cli.lastResumeSync.IsZero() && now.Sub(cli.lastResumeSync) < interactiveResumeCooldown {
		return idle, false
	}

	cli.lastResumeSync = now
	return idle, true
}

func (cli *TelegramCLI) resumeInteractiveSession(ctx context.Context, idle time.Duration, force bool) error {
	if !force && idle < interactiveResumeIdleThreshold {
		return nil
	}

	now := interactiveResumeNow()
	if force {
		cli.markResumeSync(now)
	}

	resumeCtx, cancel := context.WithTimeout(ensureContext(ctx), interactiveResumeSyncTimeout)
	defer cancel()

	return cli.refreshInteractiveState(resumeCtx)
}

func (cli *TelegramCLI) markResumeSync(now time.Time) {
	if now.IsZero() {
		now = interactiveResumeNow()
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.lastResumeSync = now
}

func (cli *TelegramCLI) refreshInteractiveState(ctx context.Context) error {
	var errs []error

	target, label := cli.currentChat()
	if target != "" {
		if err := cli.syncTranscriptContext(ctx, target, label, transcriptResumeFetchLimit); err != nil {
			errs = append(errs, fmt.Errorf("refresh active chat: %w", err))
		} else {
			cli.clearChatUnreadCount(target)
		}
	}

	if _, err := interactiveResumeDialogRefresher(ctx, cli, interactiveResumeDialogLimit); err != nil {
		errs = append(errs, fmt.Errorf("refresh dialogs: %w", err))
	}

	if target != "" && cli.currentConsole() != nil {
		cli.redrawChatView()
	}

	return errors.Join(errs...)
}

func (cli *TelegramCLI) retryInteractiveRPC(ctx context.Context, fn func(context.Context) (int64, error)) (int64, error) {
	id, err := fn(ctx)
	if err == nil || ctx.Err() != nil || !isReconnectRetryable(err) {
		return id, err
	}

	resumeErr := cli.resumeInteractiveSession(ctx, interactiveResumeIdleThreshold, true)
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
