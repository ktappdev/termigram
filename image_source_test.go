package main

import (
	"context"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareLocalImageSourceAcceptsPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writePNG(t, path)

	prepared, err := prepareImageSource(context.Background(), path)
	if err != nil {
		t.Fatalf("prepareImageSource returned error: %v", err)
	}
	if prepared.MIMEType != "image/png" {
		t.Fatalf("expected image/png, got %q", prepared.MIMEType)
	}
	if prepared.SendAsFile {
		t.Fatalf("expected png photo upload, got file upload")
	}
	if !prepared.Persistent {
		t.Fatalf("expected local files to be persistent")
	}
}

func TestPrepareLocalImageSourceRejectsGIF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.gif")
	writeGIF(t, path)

	if _, err := prepareImageSource(context.Background(), path); err == nil {
		t.Fatalf("expected gif to be rejected")
	}
}

func TestPrepareImageSourceAcceptsFileURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")
	writePNG(t, path)

	prepared, err := prepareImageSource(context.Background(), "file://"+path)
	if err != nil {
		t.Fatalf("prepareImageSource returned error: %v", err)
	}
	if prepared.Path != path {
		t.Fatalf("expected %q, got %q", path, prepared.Path)
	}
}

func TestPrepareRemoteImageSourceRejectsOversize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", "22000000")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if _, err := prepareImageSource(context.Background(), server.URL); err == nil || !strings.Contains(err.Error(), "20 MiB") {
		t.Fatalf("expected oversize error, got %v", err)
	}
}

func TestPrepareRemoteImageSourceDownloadsPNG(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		writePNGToWriter(t, w)
	}))
	defer server.Close()

	prepared, err := prepareImageSource(context.Background(), server.URL+"/meme")
	if err != nil {
		t.Fatalf("prepareImageSource returned error: %v", err)
	}
	defer prepared.Cleanup()
	if prepared.MIMEType != "image/png" {
		t.Fatalf("expected image/png, got %q", prepared.MIMEType)
	}
	if prepared.Persistent {
		t.Fatalf("expected remote download to be temporary")
	}
	if _, err := os.Stat(prepared.Path); err != nil {
		t.Fatalf("expected downloaded file to exist: %v", err)
	}
}

func TestPrepareRemoteImageSourceRejectsLyingImageHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = io.WriteString(w, "<html>not really an image</html>")
	}))
	defer server.Close()

	if _, err := prepareImageSource(context.Background(), server.URL+"/fake.png"); err == nil {
		t.Fatalf("expected non-image body to be rejected even with image/png header")
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer file.Close()
	writePNGToWriter(t, file)
}

func writePNGToWriter(t *testing.T, file interface{ Write([]byte) (int, error) }) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 1, color.RGBA{G: 255, A: 255})
	if err := png.Encode(writerAdapter{file}, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func writeGIF(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create gif: %v", err)
	}
	defer file.Close()

	img := image.NewPaletted(image.Rect(0, 0, 1, 1), []color.Color{color.Black, color.White})
	img.SetColorIndex(0, 0, 1)
	if err := gif.Encode(file, img, nil); err != nil {
		t.Fatalf("encode gif: %v", err)
	}
}

type writerAdapter struct {
	w interface{ Write([]byte) (int, error) }
}

func (w writerAdapter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}
