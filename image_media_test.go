package main

import (
	"strings"
	"testing"

	"github.com/gotd/td/tg"
)

func TestMessageOutputFromTGMessagePhotoWithoutCaption(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.cacheUser(&tg.User{ID: 42, FirstName: "Alice"})

	msg := &tg.Message{
		ID:      101,
		Date:    1_700_000_000,
		FromID:  &tg.PeerUser{UserID: 42},
		Message: "",
		Media: &tg.MessageMediaPhoto{
			Photo: &tg.Photo{
				ID:            7,
				AccessHash:    9,
				FileReference: []byte("ref"),
				Date:          1_700_000_000,
				DCID:          2,
				Sizes: []tg.PhotoSizeClass{
					&tg.PhotoSize{Type: "x", W: 20, H: 20, Size: 123},
				},
			},
		},
	}

	out := messageOutputFromTGMessage(cli, msg)
	if out.Image == nil {
		t.Fatalf("expected image metadata")
	}
	if !strings.Contains(out.Message, "[image #101]") {
		t.Fatalf("expected placeholder message, got %q", out.Message)
	}
}

func TestMessageOutputFromTGMessageImageDocumentWithCaption(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.cacheUser(&tg.User{ID: 99, FirstName: "Bob"})

	msg := &tg.Message{
		ID:      202,
		Date:    1_700_000_100,
		FromID:  &tg.PeerUser{UserID: 99},
		Message: "look at this",
		Media: &tg.MessageMediaDocument{
			Document: &tg.Document{
				ID:            11,
				AccessHash:    12,
				FileReference: []byte("ref"),
				Date:          1_700_000_100,
				DCID:          3,
				MimeType:      "image/webp",
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeFilename{FileName: "meme.webp"},
				},
			},
		},
	}

	out := messageOutputFromTGMessage(cli, msg)
	if out.Image == nil {
		t.Fatalf("expected image metadata")
	}
	if out.Image.MIMEType != "image/webp" {
		t.Fatalf("unexpected mime type: %q", out.Image.MIMEType)
	}
	if !strings.Contains(out.Message, "look at this") {
		t.Fatalf("expected caption in placeholder, got %q", out.Message)
	}
}
