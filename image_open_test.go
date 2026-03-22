package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotd/td/tg"
)

func TestOpenImageFromCurrentChatDefaultsToLast(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.setCurrentChat("@alice", "Alice")
	cli.legacyLoaded[normalizeLegacyTranscriptTarget("@alice")] = true

	path := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(path, []byte("jpg"), 0o644); err != nil {
		t.Fatalf("write cached image: %v", err)
	}

	cli.appendLegacyTranscriptEntry("@alice", legacyTranscriptEntry{
		MessageID: 5,
		Body:      "[image #5] photo.jpg",
		Image: &ImageAttachment{
			Kind:       imageKind,
			Name:       "photo.jpg",
			MIMEType:   "image/jpeg",
			CachedPath: path,
		},
	})

	originalOpen := openLocalPath
	defer func() { openLocalPath = originalOpen }()

	var opened string
	openLocalPath = func(path string) error {
		opened = path
		return nil
	}

	got, err := cli.openImageFromCurrentChat(context.Background(), "")
	if err != nil {
		t.Fatalf("openImageFromCurrentChat returned error: %v", err)
	}
	if got != path || opened != path {
		t.Fatalf("expected %q to be opened, got path=%q opened=%q", path, got, opened)
	}
}

func TestEnsureImageDownloadedUsesCacheAndReuse(t *testing.T) {
	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	entry := legacyTranscriptEntry{
		MessageID: 9,
		Image: &ImageAttachment{
			Kind:     imageKind,
			Name:     "photo.jpg",
			MIMEType: "image/jpeg",
			Location: &tg.InputDocumentFileLocation{},
		},
	}
	dir, err := mediaCacheDir()
	if err != nil {
		t.Fatalf("mediaCacheDir returned error: %v", err)
	}
	expectedPath := filepath.Join(dir, cachedImageFilename("@alice", entry))
	_ = os.Remove(expectedPath)

	originalDownload := imageDownloadFunc
	defer func() { imageDownloadFunc = originalDownload }()

	var calls int
	imageDownloadFunc = func(ctx context.Context, cli *TelegramCLI, entry legacyTranscriptEntry, path string) error {
		calls++
		return os.WriteFile(path, []byte("photo"), 0o644)
	}

	first, err := cli.ensureImageDownloaded(context.Background(), "@alice", entry)
	if err != nil {
		t.Fatalf("ensureImageDownloaded returned error: %v", err)
	}
	second, err := cli.ensureImageDownloaded(context.Background(), "@alice", legacyTranscriptEntry{
		MessageID: 9,
		Image: &ImageAttachment{
			Kind:       imageKind,
			Name:       "photo.jpg",
			MIMEType:   "image/jpeg",
			Location:   &tg.InputDocumentFileLocation{},
			CachedPath: first,
		},
	})
	if err != nil {
		t.Fatalf("ensureImageDownloaded second call returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one download call, got %d", calls)
	}
	if first != second {
		t.Fatalf("expected cache reuse, got %q and %q", first, second)
	}
}

func TestSendImageCachesRemoteSourceForReopen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		writePNGToWriter(t, w)
	}))
	defer server.Close()

	cli := NewTelegramCLI(1, "hash", t.TempDir()+"/session.json")
	cli.setCurrentChat("@alice", "Alice")

	originalSend := sendPreparedImageWithBackend
	originalOpen := openLocalPath
	defer func() {
		sendPreparedImageWithBackend = originalSend
		openLocalPath = originalOpen
	}()

	sendPreparedImageWithBackend = func(ctx context.Context, backend *UserBackend, target string, prepared preparedImageSource, caption string) error {
		return nil
	}

	var opened string
	openLocalPath = func(path string) error {
		opened = path
		return nil
	}

	if err := cli.sendImage(context.Background(), "@alice", "Alice", server.URL+"/meme.png", "hello"); err != nil {
		t.Fatalf("sendImage returned error: %v", err)
	}
	cli.legacyLoaded[normalizeLegacyTranscriptTarget("@alice")] = true

	got, err := cli.openImageFromCurrentChat(context.Background(), "last")
	if err != nil {
		t.Fatalf("openImageFromCurrentChat returned error: %v", err)
	}
	if got == "" || opened == "" {
		t.Fatalf("expected cached image path to be opened")
	}
	if got != opened {
		t.Fatalf("expected opened path %q to match returned path %q", opened, got)
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("expected cached outbound image to exist: %v", err)
	}
}
