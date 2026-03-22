package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram/downloader"
)

var imageDownloadFunc = downloadImageAttachment
var openLocalPath = defaultOpenLocalPath
var cacheOutboundImageFunc = cacheOutboundImageCopy

func (cli *TelegramCLI) ensureImageDownloaded(ctx context.Context, target string, entry legacyTranscriptEntry) (string, error) {
	if entry.Image == nil {
		return "", fmt.Errorf("message does not contain an image")
	}

	if fileExists(entry.Image.CachedPath) {
		return entry.Image.CachedPath, nil
	}
	if entry.Image.Location == nil {
		if fileExists(entry.Image.CachedPath) {
			return entry.Image.CachedPath, nil
		}
		return "", fmt.Errorf("image is not downloadable")
	}

	dir, err := mediaCacheDir()
	if err != nil {
		return "", err
	}
	filename := cachedImageFilename(target, entry)
	path := filepath.Join(dir, filename)
	if fileExists(path) {
		entry.Image.CachedPath = path
		cli.updateLegacyImageCachePath(target, entry.MessageID, path)
		return path, nil
	}

	if err := imageDownloadFunc(ensureContext(ctx), cli, entry, path); err != nil {
		return "", err
	}
	cli.updateLegacyImageCachePath(target, entry.MessageID, path)
	return path, nil
}

func mediaCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = filepath.Join(os.TempDir(), "termigram-media")
	} else {
		base = filepath.Join(base, "termigram", "media")
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("create media cache directory: %w", err)
	}
	return base, nil
}

func cachedImageFilename(target string, entry legacyTranscriptEntry) string {
	name := "image"
	if entry.Image != nil && strings.TrimSpace(entry.Image.Name) != "" {
		name = filepath.Base(entry.Image.Name)
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	base = sanitizeFilename(base)
	if base == "" {
		base = "image"
	}
	if ext == "" {
		ext = ".jpg"
	}

	target = sanitizeFilename(strings.TrimPrefix(normalizeUsername(target), "@"))
	if target == "" {
		target = "chat"
	}
	if entry.MessageID == 0 {
		return fmt.Sprintf("%s-%s%s", target, base, ext)
	}
	return fmt.Sprintf("%s-%d-%s%s", target, entry.MessageID, base, ext)
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "\n", "-", "\r", "-", "\t", "-", " ", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, ".-")
	return value
}

func downloadImageAttachment(ctx context.Context, cli *TelegramCLI, entry legacyTranscriptEntry, path string) error {
	if entry.Image == nil || entry.Image.Location == nil {
		return fmt.Errorf("image is not downloadable")
	}
	_, err := downloader.NewDownloader().Download(cli.api, entry.Image.Location).ToPath(ctx, path)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}
	return nil
}

func defaultOpenLocalPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		return fmt.Errorf("opening images is not supported on %s", runtime.GOOS)
	}
	return cmd.Start()
}

func parseMessageID(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("message id is required")
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid message id %q", raw)
	}
	return id, nil
}

func cacheOutboundImageCopy(target string, attachment *ImageAttachment, sourcePath string) (string, error) {
	if attachment == nil {
		return "", fmt.Errorf("image attachment is required")
	}
	dir, err := mediaCacheDir()
	if err != nil {
		return "", err
	}
	entry := legacyTranscriptEntry{
		Image: attachment,
	}
	filename := cachedImageFilename(target, entry)
	path := filepath.Join(dir, filename)

	src, err := os.Open(filepath.Clean(sourcePath))
	if err != nil {
		return "", fmt.Errorf("open outbound image cache source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create outbound image cache file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("cache outbound image: %w", err)
	}
	return path, nil
}
