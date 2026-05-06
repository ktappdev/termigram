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
	cli := NewTelegramCLI(1, "hash", filepath.Join(t.TempDir(), "session.json"))
	cli.ttyCheck = func() bool { return true }

	cli.envLookup = func(key string) string {
		switch key {
		case "TERM":
			return "xterm-kitty"
		default:
			return ""
		}
	}
	kitty := cli.detectInlineImageProtocol(inlineImageModeAuto, inlineImageProtocolNone)
	if kitty != inlineImageProtocolKitty {
		t.Fatalf("expected kitty protocol, got %q", kitty)
	}

	cli.envLookup = func(key string) string {
		switch key {
		case "TERM_PROGRAM":
			return "iTerm.app"
		default:
			return ""
		}
	}
	iterm := cli.detectInlineImageProtocol(inlineImageModeAuto, inlineImageProtocolNone)
	if iterm != inlineImageProtocolITerm2 {
		t.Fatalf("expected iTerm2 protocol, got %q", iterm)
	}
}

func TestDetectInlineImageProtocolDisablesAutoUnderTmux(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", filepath.Join(t.TempDir(), "session.json"))
	cli.ttyCheck = func() bool { return true }
	cli.envLookup = func(key string) string {
		switch key {
		case "TERM":
			return "xterm-kitty"
		case "TMUX":
			return "/tmp/tmux-1/default,123,0"
		default:
			return ""
		}
	}
	got := cli.detectInlineImageProtocol(inlineImageModeAuto, inlineImageProtocolNone)
	if got != inlineImageProtocolNone {
		t.Fatalf("expected auto mode to disable previews under tmux, got %q", got)
	}
}

func TestBuildInlinePreviewPNGProducesSmallPNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.png")
	writeLargePNG(t, path, 1600, 1200)

	data, rows, err := buildInlinePreviewPNG(path, 28, 10)
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

func TestRenderInlineImageBlockAlignsOutgoingRight(t *testing.T) {
	t.Setenv("TERMIGRAM_INLINE_IMAGES", "on")
	t.Setenv("TERMIGRAM_INLINE_IMAGE_PROTOCOL", "kitty")

	path := filepath.Join(t.TempDir(), "preview.png")
	writePNG(t, path)

	cli := NewTelegramCLI(1, "hash", filepath.Join(t.TempDir(), "session.json"))
	cli.ttyCheck = func() bool { return true }
	block, rows, ok := cli.renderInlineImageBlock("@alice", transcriptEntry{
		MessageID: 6,
		Outgoing:  true,
		Header:    "You",
		Body:      "[image #6] latest.png",
		Meta:      "10:01:00",
		Image: &ImageAttachment{
			Kind:       imageKind,
			Name:       "latest.png",
			MIMEType:   "image/png",
			CachedPath: path,
		},
	}, 100, cli.currentInlineImageConfig())
	if !ok || rows == 0 {
		t.Fatalf("expected inline image block to render")
	}
	if !strings.Contains(block, "\x1b_Ga=T,f=100,t=f") {
		t.Fatalf("expected kitty graphics sequence in inline block")
	}
	if !strings.HasPrefix(block, " ") {
		t.Fatalf("expected outgoing inline image to be indented to the right, got %q", block)
	}
}

func TestRenderActiveChatViewKeepsImagesInFlow(t *testing.T) {
	t.Setenv("TERMIGRAM_INLINE_IMAGES", "on")
	t.Setenv("TERMIGRAM_INLINE_IMAGE_PROTOCOL", "kitty")

	path := filepath.Join(t.TempDir(), "preview.png")
	writePNG(t, path)

	cli := NewTelegramCLI(1, "hash", filepath.Join(t.TempDir(), "session.json"))
	cli.ttyCheck = func() bool { return true }
	view := cli.renderActiveChatView("Alice", "@alice", []transcriptEntry{
		{
			MessageID: 1,
			Header:    "Alice",
			Body:      "before image",
			Meta:      "10:00:00",
		},
		{
			MessageID: 2,
			Outgoing:  true,
			Header:    "You",
			Body:      "[image #2] sent.png\ncaption",
			Meta:      "10:01:00  ✓",
			Image: &ImageAttachment{
				Kind:       imageKind,
				Name:       "sent.png",
				MIMEType:   "image/png",
				CachedPath: path,
			},
		},
		{
			MessageID: 3,
			Header:    "Alice",
			Body:      "after image",
			Meta:      "10:02:00",
		},
	}, 100, 30)

	beforeIndex := strings.Index(view, "before image")
	imageIndex := strings.Index(view, "\x1b_Ga=T,f=100,t=f")
	afterIndex := strings.Index(view, "after image")
	if beforeIndex < 0 || imageIndex < 0 || afterIndex < 0 {
		t.Fatalf("expected view to contain before text, kitty image block, and after text")
	}
	if !(beforeIndex < imageIndex && imageIndex < afterIndex) {
		t.Fatalf("expected image block to stay in message flow, got before=%d image=%d after=%d", beforeIndex, imageIndex, afterIndex)
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
