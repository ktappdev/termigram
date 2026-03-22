package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxRemoteImageBytes int64 = 20 << 20

type preparedImageSource struct {
	Path       string
	Name       string
	MIMEType   string
	SendAsFile bool
	Persistent bool
	Cleanup    func()
}

var remoteImageHTTPClient = &http.Client{Timeout: 30 * time.Second}

func normalizeImageSource(raw string) (string, error) {
	source := strings.TrimSpace(raw)
	if source == "" {
		return "", fmt.Errorf("image source is required")
	}

	if unquoted, err := unquoteIfNeeded(source); err == nil {
		source = unquoted
	}

	if strings.HasPrefix(source, "file://") {
		parsed, err := url.Parse(source)
		if err != nil {
			return "", fmt.Errorf("invalid file URL: %w", err)
		}
		if parsed.Path == "" {
			return "", fmt.Errorf("file URL is missing a path")
		}
		path, err := url.PathUnescape(parsed.Path)
		if err != nil {
			return "", fmt.Errorf("invalid file URL path: %w", err)
		}
		return path, nil
	}

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return source, nil
	}

	return unescapeShellPath(source), nil
}

func unquoteIfNeeded(value string) (string, error) {
	if len(value) < 2 {
		return value, nil
	}
	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1], nil
	}
	return value, nil
}

func unescapeShellPath(value string) string {
	runes := []rune(value)
	var builder strings.Builder
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '\\' || i == len(runes)-1 {
			builder.WriteRune(r)
			continue
		}

		next := runes[i+1]
		switch next {
		case ' ', '\t', '"', '\'', '\\':
			builder.WriteRune(next)
			i++
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func prepareImageSource(ctx context.Context, raw string) (preparedImageSource, error) {
	source, err := normalizeImageSource(raw)
	if err != nil {
		return preparedImageSource{}, err
	}

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return prepareRemoteImageSource(ctx, source)
	}
	return prepareLocalImageSource(source)
}

func prepareLocalImageSource(path string) (preparedImageSource, error) {
	if path == "" {
		return preparedImageSource{}, fmt.Errorf("image path is required")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return preparedImageSource{}, fmt.Errorf("resolve image path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return preparedImageSource{}, fmt.Errorf("read image path: %w", err)
	}
	if info.IsDir() {
		return preparedImageSource{}, fmt.Errorf("image path points to a directory")
	}
	mimeType, sendAsFile, name, err := inspectImageFile(absPath, "")
	if err != nil {
		return preparedImageSource{}, err
	}
	return preparedImageSource{
		Path:       absPath,
		Name:       name,
		MIMEType:   mimeType,
		SendAsFile: sendAsFile,
		Persistent: true,
		Cleanup:    func() {},
	}, nil
}

func prepareRemoteImageSource(ctx context.Context, rawURL string) (preparedImageSource, error) {
	req, err := http.NewRequestWithContext(ensureContext(ctx), http.MethodGet, rawURL, nil)
	if err != nil {
		return preparedImageSource{}, fmt.Errorf("build image request: %w", err)
	}

	resp, err := remoteImageHTTPClient.Do(req)
	if err != nil {
		return preparedImageSource{}, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return preparedImageSource{}, fmt.Errorf("download image: unexpected HTTP status %s", resp.Status)
	}
	if resp.ContentLength > maxRemoteImageBytes {
		return preparedImageSource{}, fmt.Errorf("download image: response exceeds 20 MiB limit")
	}

	tmpDir, err := os.MkdirTemp("", "termigram-image-*")
	if err != nil {
		return preparedImageSource{}, fmt.Errorf("create temp directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	name := remoteImageName(rawURL, resp.Header.Get("Content-Type"))
	targetPath := filepath.Join(tmpDir, name)
	file, err := os.Create(targetPath)
	if err != nil {
		cleanup()
		return preparedImageSource{}, fmt.Errorf("create temp image file: %w", err)
	}

	limited := io.LimitReader(resp.Body, maxRemoteImageBytes+1)
	written, copyErr := io.Copy(file, limited)
	closeErr := file.Close()
	if copyErr != nil {
		cleanup()
		return preparedImageSource{}, fmt.Errorf("download image: %w", copyErr)
	}
	if closeErr != nil {
		cleanup()
		return preparedImageSource{}, fmt.Errorf("close temp image file: %w", closeErr)
	}
	if written > maxRemoteImageBytes {
		cleanup()
		return preparedImageSource{}, fmt.Errorf("download image: response exceeds 20 MiB limit")
	}

	mimeType, sendAsFile, resolvedName, err := inspectImageFile(targetPath, resp.Header.Get("Content-Type"))
	if err != nil {
		cleanup()
		return preparedImageSource{}, err
	}
	finalPath := filepath.Join(tmpDir, resolvedName)
	if finalPath != targetPath {
		if err := os.Rename(targetPath, finalPath); err == nil {
			targetPath = finalPath
		}
	}

	return preparedImageSource{
		Path:       targetPath,
		Name:       resolvedName,
		MIMEType:   mimeType,
		SendAsFile: sendAsFile,
		Persistent: false,
		Cleanup:    cleanup,
	}, nil
}

func remoteImageName(rawURL string, contentType string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		if base := filepath.Base(parsed.Path); base != "." && base != "/" && base != "" {
			return base
		}
	}
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png":
		return "remote.png"
	case "image/webp":
		return "remote.webp"
	default:
		return "remote.jpg"
	}
}

func inspectImageFile(path string, headerType string) (mimeType string, sendAsFile bool, name string, err error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", false, "", fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	header := make([]byte, 512)
	n, readErr := file.Read(header)
	if readErr != nil && readErr != io.EOF {
		return "", false, "", fmt.Errorf("read image header: %w", readErr)
	}

	detectedType := http.DetectContentType(header[:n])
	if !isSupportedImageMIME(detectedType) {
		return "", false, "", fmt.Errorf("unsupported image format (supported: jpg, png, webp)")
	}

	name = filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(name))
	switch detectedType {
	case "image/jpeg":
		if ext != ".jpg" && ext != ".jpeg" {
			name = strings.TrimSuffix(name, filepath.Ext(name)) + ".jpg"
		}
	case "image/png":
		if ext != ".png" {
			name = strings.TrimSuffix(name, filepath.Ext(name)) + ".png"
		}
	case "image/webp":
		if ext != ".webp" {
			name = strings.TrimSuffix(name, filepath.Ext(name)) + ".webp"
		}
		sendAsFile = true
	}

	return detectedType, sendAsFile, name, nil
}
