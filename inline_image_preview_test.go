package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectInlineImageProtocolAuto(t *testing.T) {
	originalTTYCheck := inlineImageTTYCheck
	inlineImageTTYCheck = func() bool { return true }
	defer func() { inlineImageTTYCheck = originalTTYCheck }()

	kitty := detectInlineImageProtocol(func(key string) string {
		switch key {
		case "TERM":
			return "xterm-kitty"
		default:
			return ""
		}
	}, inlineImageModeAuto, inlineImageProtocolNone)
	if kitty != inlineImageProtocolKitty {
		t.Fatalf("expected kitty protocol, got %q", kitty)
	}

	iterm := detectInlineImageProtocol(func(key string) string {
		switch key {
		case "TERM_PROGRAM":
			return "iTerm.app"
		default:
			return ""
		}
	}, inlineImageModeAuto, inlineImageProtocolNone)
	if iterm != inlineImageProtocolITerm2 {
		t.Fatalf("expected iTerm2 protocol, got %q", iterm)
	}
}

func TestDetectInlineImageProtocolDisablesAutoUnderTmux(t *testing.T) {
	originalTTYCheck := inlineImageTTYCheck
	inlineImageTTYCheck = func() bool { return true }
	defer func() { inlineImageTTYCheck = originalTTYCheck }()

	got := detectInlineImageProtocol(func(key string) string {
		switch key {
		case "TERM":
			return "xterm-kitty"
		case "TMUX":
			return "/tmp/tmux-1/default,123,0"
		default:
			return ""
		}
	}, inlineImageModeAuto, inlineImageProtocolNone)
	if got != inlineImageProtocolNone {
		t.Fatalf("expected auto mode to disable previews under tmux, got %q", got)
	}
}

func TestBuildInlinePreviewPNGProducesSmallPNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.png")
	writeLargePNG(t, path, 1600, 1200)

	data, rows, err := buildInlinePreviewPNG(path, inlineImageConfig{
		Protocol: inlineImageProtocolITerm2,
		Cols:     28,
		MaxRows:  10,
	})
	if err != nil {
		t.Fatalf("buildInlinePreviewPNG returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected preview bytes")
	}
	if len(data) > inlineImageMaxBytes {
		t.Fatalf("expected preview bytes <= %d, got %d", inlineImageMaxBytes, len(data))
	}
	if rows < 3 || rows > 10 {
		t.Fatalf("expected rows to be between 3 and 10, got %d", rows)
	}
}

func TestRenderInlineImagePreviewUsesLatestImage(t *testing.T) {
	originalTTYCheck := inlineImageTTYCheck
	inlineImageTTYCheck = func() bool { return true }
	defer func() { inlineImageTTYCheck = originalTTYCheck }()

	t.Setenv("TERMIGRAM_INLINE_IMAGES", "on")
	t.Setenv("TERMIGRAM_INLINE_IMAGE_PROTOCOL", "kitty")

	path := filepath.Join(t.TempDir(), "preview.png")
	writePNG(t, path)

	cli := NewTelegramCLI(1, "hash", filepath.Join(t.TempDir(), "session.json"))
	block, rows := cli.renderInlineImagePreview("@alice", []legacyTranscriptEntry{
		{
			MessageID: 5,
			Header:    "Alice",
			Body:      "[image #5] old.png",
			Meta:      "10:00:00",
			Image: &ImageAttachment{
				Kind:       imageKind,
				Name:       "old.png",
				MIMEType:   "image/png",
				CachedPath: path,
			},
		},
		{
			MessageID: 6,
			Header:    "Alice",
			Body:      "[image #6] latest.png",
			Meta:      "10:01:00",
			Image: &ImageAttachment{
				Kind:       imageKind,
				Name:       "latest.png",
				MIMEType:   "image/png",
				CachedPath: path,
			},
		},
	}, 100, 24)
	if rows == 0 {
		t.Fatalf("expected preview rows to be reserved")
	}
	if !strings.Contains(block, "Image preview: latest.png") {
		t.Fatalf("expected preview block for latest image, got %q", block)
	}
	if !strings.Contains(block, "\x1b_Ga=T,f=100,t=f") {
		t.Fatalf("expected kitty graphics sequence in preview block")
	}
}

func writeLargePNG(t *testing.T, path string, width int, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / maxInt(width-1, 1)),
				G: uint8((y * 255) / maxInt(height-1, 1)),
				B: 180,
				A: 255,
			})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create large png: %v", err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode large png: %v", err)
	}
}
