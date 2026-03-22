package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "golang.org/x/image/webp"
)

type inlineImageProtocol string

const (
	inlineImageProtocolNone   inlineImageProtocol = ""
	inlineImageProtocolITerm2 inlineImageProtocol = "iterm2"
	inlineImageProtocolKitty  inlineImageProtocol = "kitty"
)

const (
	inlineImageModeAuto    = "auto"
	inlineImageModeOn      = "on"
	inlineImageModeOff     = "off"
	defaultInlineImageCols = 28
	defaultInlineImageRows = 10
	inlineImageMinPaneRows = 8
	inlineImageMaxBytes    = 900_000
	inlineImageCellAspect  = 0.5
	inlineImagePixelWidth  = 16
	inlineImagePixelHeight = 32
)

type inlineImageConfig struct {
	Mode     string
	Protocol inlineImageProtocol
	Cols     int
	MaxRows  int
}

type inlinePreviewImage struct {
	Path string
	Cols int
	Rows int
	Name string
}

var inlineImageEnvLookup = os.Getenv
var inlineImageTTYCheck = interactiveTTYAvailable

func currentInlineImageConfig() inlineImageConfig {
	mode := strings.ToLower(strings.TrimSpace(inlineImageEnvLookup("TERMIGRAM_INLINE_IMAGES")))
	switch mode {
	case "", inlineImageModeAuto:
		mode = inlineImageModeAuto
	case inlineImageModeOn, inlineImageModeOff:
	default:
		mode = inlineImageModeAuto
	}

	forcedProtocol := parseInlineImageProtocol(inlineImageEnvLookup("TERMIGRAM_INLINE_IMAGE_PROTOCOL"))
	protocol := detectInlineImageProtocol(inlineImageEnvLookup, mode, forcedProtocol)

	cols := parsePositiveEnvInt(inlineImageEnvLookup("TERMIGRAM_INLINE_IMAGE_COLS"), defaultInlineImageCols)
	maxRows := parsePositiveEnvInt(inlineImageEnvLookup("TERMIGRAM_INLINE_IMAGE_ROWS"), defaultInlineImageRows)

	return inlineImageConfig{
		Mode:     mode,
		Protocol: protocol,
		Cols:     cols,
		MaxRows:  maxRows,
	}
}

func parseInlineImageProtocol(raw string) inlineImageProtocol {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "kitty":
		return inlineImageProtocolKitty
	case "iterm2":
		return inlineImageProtocolITerm2
	default:
		return inlineImageProtocolNone
	}
}

func detectInlineImageProtocol(getenv func(string) string, mode string, forced inlineImageProtocol) inlineImageProtocol {
	if mode == inlineImageModeOff || !inlineImageTTYCheck() {
		return inlineImageProtocolNone
	}
	if forced != inlineImageProtocolNone {
		return forced
	}

	if mode != inlineImageModeOn {
		if strings.TrimSpace(getenv("TMUX")) != "" || strings.Contains(strings.ToLower(strings.TrimSpace(getenv("TERM"))), "screen") {
			return inlineImageProtocolNone
		}
	}

	if strings.EqualFold(strings.TrimSpace(getenv("TERM_PROGRAM")), "iTerm.app") {
		return inlineImageProtocolITerm2
	}

	termValue := strings.ToLower(strings.TrimSpace(getenv("TERM")))
	if strings.Contains(termValue, "kitty") || strings.TrimSpace(getenv("KITTY_WINDOW_ID")) != "" {
		return inlineImageProtocolKitty
	}

	return inlineImageProtocolNone
}

func parsePositiveEnvInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (cfg inlineImageConfig) enabled() bool {
	return cfg.Protocol != inlineImageProtocolNone && cfg.Cols > 0 && cfg.MaxRows > 0
}

func inlineImageColsForWidth(width int, cfg inlineImageConfig) int {
	if width <= 0 {
		return cfg.Cols
	}

	maxCols := transcriptBubbleWidth(width) - 2
	if maxCols < 8 {
		maxCols = 8
	}
	if cfg.Cols < maxCols {
		maxCols = cfg.Cols
	}
	if maxCols > width {
		maxCols = width
	}
	return maxCols
}

func (cli *TelegramCLI) renderInlineImageBlock(target string, entry legacyTranscriptEntry, width int, cfg inlineImageConfig) (string, int, bool) {
	if entry.Image == nil {
		return "", 0, false
	}

	cols := inlineImageColsForWidth(width, cfg)
	if cols <= 0 {
		return "", 0, false
	}

	path, err := cli.ensureImageDownloaded(ensureContext(cli.ctx), target, entry)
	if err != nil || !fileExists(path) {
		return "", 0, false
	}

	preview, err := ensureInlinePreviewImage(target, entry, path, cols, cfg.MaxRows)
	if err != nil {
		return "", 0, false
	}

	sequence, err := renderInlineImageSequence(preview, cfg.Protocol)
	if err != nil {
		return "", 0, false
	}

	indent := 0
	if entry.Outgoing && width > preview.Cols {
		indent = width - preview.Cols
	}
	return strings.Repeat(" ", maxInt(indent, 0)) + sequence, preview.Rows, true
}

func ensureInlinePreviewImage(target string, entry legacyTranscriptEntry, sourcePath string, cols int, maxRows int) (inlinePreviewImage, error) {
	path := filepath.Join(previewCacheDirPath(), cachedInlinePreviewFilename(target, entry, cols, maxRows))
	rows := maxRows

	if fileExists(path) {
		width, height, err := decodeImageSize(path)
		if err == nil {
			rows = calculateInlinePreviewRows(width, height, cols, maxRows)
		}
		return inlinePreviewImage{
			Path: path,
			Cols: cols,
			Rows: rows,
			Name: inlinePreviewName(entry),
		}, nil
	}

	data, rows, err := buildInlinePreviewPNG(sourcePath, cols, maxRows)
	if err != nil {
		return inlinePreviewImage{}, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return inlinePreviewImage{}, fmt.Errorf("write inline preview: %w", err)
	}

	return inlinePreviewImage{
		Path: path,
		Cols: cols,
		Rows: rows,
		Name: inlinePreviewName(entry),
	}, nil
}

func previewCacheDirPath() string {
	dir, err := mediaCacheDir()
	if err != nil {
		dir = filepath.Join(os.TempDir(), "termigram-media")
		_ = os.MkdirAll(dir, 0o755)
	}
	return dir
}

func cachedInlinePreviewFilename(target string, entry legacyTranscriptEntry, cols int, maxRows int) string {
	name := cachedImageFilename(target, entry)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s-inline-%dx%d-preview.png", base, cols, maxRows)
}

func inlinePreviewName(entry legacyTranscriptEntry) string {
	if entry.Image != nil && strings.TrimSpace(entry.Image.Name) != "" {
		return strings.TrimSpace(entry.Image.Name)
	}
	return "image"
}

func buildInlinePreviewPNG(sourcePath string, cols int, maxRows int) ([]byte, int, error) {
	src, err := decodeImageFile(sourcePath)
	if err != nil {
		return nil, 0, err
	}

	maxWidth := maxInt(cols*inlineImagePixelWidth, 64)
	maxHeight := maxInt(maxRows*inlineImagePixelHeight, 64)

	for attempt := 0; attempt < 5; attempt++ {
		resized := resizeImageToFit(src, maxWidth, maxHeight)
		rows := calculateInlinePreviewRows(resized.Bounds().Dx(), resized.Bounds().Dy(), cols, maxRows)

		var buf bytes.Buffer
		if err := png.Encode(&buf, resized); err != nil {
			return nil, 0, fmt.Errorf("encode inline preview png: %w", err)
		}
		if buf.Len() <= inlineImageMaxBytes || (maxWidth <= 96 && maxHeight <= 96) {
			return buf.Bytes(), rows, nil
		}

		maxWidth = maxInt((maxWidth*3)/4, 96)
		maxHeight = maxInt((maxHeight*3)/4, 96)
	}

	return nil, 0, fmt.Errorf("could not reduce inline preview below %d bytes", inlineImageMaxBytes)
}

func decodeImageFile(path string) (image.Image, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("open preview source: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode preview source: %w", err)
	}
	return img, nil
}

func decodeImageSize(path string) (int, int, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func resizeImageToFit(src image.Image, maxWidth int, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	if srcWidth <= 0 || srcHeight <= 0 {
		return src
	}
	if srcWidth <= maxWidth && srcHeight <= maxHeight {
		return src
	}

	scale := math.Min(float64(maxWidth)/float64(srcWidth), float64(maxHeight)/float64(srcHeight))
	dstWidth := maxInt(int(math.Round(float64(srcWidth)*scale)), 1)
	dstHeight := maxInt(int(math.Round(float64(srcHeight)*scale)), 1)

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	for y := 0; y < dstHeight; y++ {
		srcY := bounds.Min.Y + int(math.Floor((float64(y)*float64(srcHeight))/float64(dstHeight)))
		if srcY >= bounds.Max.Y {
			srcY = bounds.Max.Y - 1
		}
		for x := 0; x < dstWidth; x++ {
			srcX := bounds.Min.X + int(math.Floor((float64(x)*float64(srcWidth))/float64(dstWidth)))
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func calculateInlinePreviewRows(width int, height int, cols int, maxRows int) int {
	if width <= 0 || height <= 0 {
		return maxInt(3, maxRows)
	}

	rows := int(math.Ceil((float64(height) / float64(width)) * float64(cols) * inlineImageCellAspect))
	if rows < 3 {
		rows = 3
	}
	if maxRows > 0 && rows > maxRows {
		rows = maxRows
	}
	return rows
}

func renderInlineImageSequence(preview inlinePreviewImage, protocol inlineImageProtocol) (string, error) {
	switch protocol {
	case inlineImageProtocolKitty:
		return kittyInlineImageSequence(preview), nil
	case inlineImageProtocolITerm2:
		return iTerm2InlineImageSequence(preview)
	default:
		return "", fmt.Errorf("unsupported inline image protocol %q", protocol)
	}
}

func kittyInlineImageSequence(preview inlinePreviewImage) string {
	encodedPath := base64.StdEncoding.EncodeToString([]byte(preview.Path))
	return fmt.Sprintf("\x1b_Ga=T,f=100,t=f,c=%d,r=%d;%s\x1b\\", preview.Cols, preview.Rows, encodedPath)
}

func iTerm2InlineImageSequence(preview inlinePreviewImage) (string, error) {
	data, err := os.ReadFile(preview.Path)
	if err != nil {
		return "", fmt.Errorf("read inline preview: %w", err)
	}
	if len(data) > inlineImageMaxBytes {
		return "", fmt.Errorf("inline preview exceeds %d bytes", inlineImageMaxBytes)
	}

	encodedName := base64.StdEncoding.EncodeToString([]byte(filepath.Base(preview.Name)))
	encodedData := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf(
		"\x1b]1337;File=name=%s;size=%d;width=%d;height=%d;preserveAspectRatio=1;inline=1:%s\a",
		encodedName,
		len(data),
		preview.Cols,
		preview.Rows,
		encodedData,
	), nil
}
