package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

type replyCLITestBackend struct {
	sendTarget   string
	sendText     string
	sendOpts     SendOptions
	imageTarget  string
	imageSource  string
	imageCaption string
	imageOpts    SendOptions
	messages     []MessageOutput
}

func (b *replyCLITestBackend) IsAuthorized(ctx context.Context) (bool, error) {
	return true, nil
}

func (b *replyCLITestBackend) GetSelf(ctx context.Context) (*UserOutput, error) {
	return &UserOutput{ID: 1}, nil
}

func (b *replyCLITestBackend) SendMessage(ctx context.Context, target string, text string, opts SendOptions) error {
	b.sendTarget = target
	b.sendText = text
	b.sendOpts = opts
	return nil
}

func (b *replyCLITestBackend) SendImage(ctx context.Context, target string, source string, caption string, opts SendOptions) error {
	b.imageTarget = target
	b.imageSource = source
	b.imageCaption = caption
	b.imageOpts = opts
	return nil
}

func (b *replyCLITestBackend) GetMessages(ctx context.Context, target string, limit int) ([]MessageOutput, error) {
	return b.messages, nil
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()
	_ = w.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	return string(data)
}

func TestRunCLICommandSendPassesReplyTo(t *testing.T) {
	backend := &replyCLITestBackend{}
	err := RunCLICommand(context.Background(), backend, CLICommand{
		Name:             "send",
		Args:             []string{"@alice", "hello"},
		ReplyToMessageID: 77,
	})
	if err != nil {
		t.Fatalf("RunCLICommand returned error: %v", err)
	}
	if backend.sendOpts.ReplyToMessageID != 77 {
		t.Fatalf("expected reply-to 77, got %d", backend.sendOpts.ReplyToMessageID)
	}
}

func TestRunCLICommandSendImagePassesReplyTo(t *testing.T) {
	backend := &replyCLITestBackend{}
	err := RunCLICommand(context.Background(), backend, CLICommand{
		Name:             "send-image",
		Args:             []string{"@alice", "./meme.png", "caption"},
		ReplyToMessageID: 88,
	})
	if err != nil {
		t.Fatalf("RunCLICommand returned error: %v", err)
	}
	if backend.imageOpts.ReplyToMessageID != 88 {
		t.Fatalf("expected reply-to 88, got %d", backend.imageOpts.ReplyToMessageID)
	}
}

func TestCmdGetJSONIncludesReplyMetadata(t *testing.T) {
	backend := &replyCLITestBackend{
		messages: []MessageOutput{{
			ID:       9,
			FromName: "Bob",
			Message:  "↪ Alice\n> hi\n\nhello",
			Date:     1,
			Reply:    &ReplyReference{MessageID: 5, Sender: "Alice", Preview: "hi"},
		}},
	}
	output := captureStdout(t, func() {
		err := RunCLICommand(context.Background(), backend, CLICommand{
			Name:  "get",
			Args:  []string{"@alice"},
			JSON:  true,
			Limit: 10,
		})
		if err != nil {
			t.Fatalf("RunCLICommand returned error: %v", err)
		}
	})
	if !strings.Contains(output, `"reply"`) || !strings.Contains(output, `"message_id": 5`) {
		t.Fatalf("expected reply metadata in JSON output, got %s", output)
	}
}
