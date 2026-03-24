package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	qmessages "github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
)

const (
	imageKind              = "image"
	imagePlaceholderPrefix = "[image"
)

type ImageAttachment struct {
	Kind       string                    `json:"kind,omitempty"`
	Name       string                    `json:"name,omitempty"`
	MIMEType   string                    `json:"mime_type,omitempty"`
	Location   tg.InputFileLocationClass `json:"-"`
	CachedPath string                    `json:"-"`
}

func isSupportedImageMIME(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func isSupportedImageDocument(doc *tg.Document) bool {
	if doc == nil {
		return false
	}
	if isSupportedImageMIME(doc.MimeType) {
		return true
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeImageSize); ok {
			return true
		}
	}
	return false
}

func imageAttachmentFromMessage(msg *tg.Message) (*ImageAttachment, bool) {
	if msg == nil {
		return nil, false
	}

	switch media := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		if _, ok := media.Photo.AsNotEmpty(); !ok {
			return nil, false
		}
	case *tg.MessageMediaDocument:
		doc, ok := media.Document.AsNotEmpty()
		if !ok || !isSupportedImageDocument(doc) {
			return nil, false
		}
	default:
		return nil, false
	}

	file, ok := qmessages.Elem{Msg: msg}.File()
	if !ok || file.Location == nil || !isSupportedImageMIME(file.MIMEType) {
		return nil, false
	}

	return &ImageAttachment{
		Kind:     imageKind,
		Name:     filepath.Base(strings.TrimSpace(file.Name)),
		MIMEType: strings.TrimSpace(file.MIMEType),
		Location: file.Location,
	}, true
}

func imagePlaceholderBody(messageID int64, name string, caption string) string {
	label := imagePlaceholderPrefix + "]"
	if messageID > 0 {
		label = fmt.Sprintf("%s #%d]", imagePlaceholderPrefix, messageID)
	}
	name = strings.TrimSpace(name)
	if name != "" {
		label += " " + name
	}

	caption = strings.TrimSpace(caption)
	if caption == "" {
		return label
	}
	return label + "\n" + caption
}

func imagePreviewText(attachment *ImageAttachment, caption string) string {
	caption = strings.TrimSpace(caption)
	if caption != "" {
		return "[image] " + caption
	}
	if attachment != nil && strings.TrimSpace(attachment.Name) != "" {
		return "[image] " + strings.TrimSpace(attachment.Name)
	}
	return "[image]"
}

func latestImageEntry(entries []legacyTranscriptEntry) (*legacyTranscriptEntry, bool) {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Image != nil {
			entry := entries[i]
			return &entry, true
		}
	}
	return nil, false
}

func findImageEntryByID(entries []legacyTranscriptEntry, messageID int64) (*legacyTranscriptEntry, bool) {
	for _, entry := range entries {
		if entry.MessageID == messageID && entry.Image != nil {
			copy := entry
			return &copy, true
		}
	}
	return nil, false
}

func (cli *TelegramCLI) openImageFromCurrentChat(ctx context.Context, selector string) (string, error) {
	target, label := cli.currentChat()
	if target == "" {
		return "", fmt.Errorf("no active chat; switch chats with \\to or \\msg first")
	}

	if err := cli.ensureLegacyTranscript(ctx, target, label); err != nil {
		return "", err
	}

	entries, _ := cli.legacyTranscriptSnapshot(target)
	if len(entries) == 0 {
		return "", fmt.Errorf("no messages available for the active chat")
	}

	var (
		entry *legacyTranscriptEntry
		ok    bool
	)

	selector = strings.TrimSpace(selector)
	switch {
	case selector == "":
		result := cli.pickImageEntry("Open image", "", entries)
		switch {
		case !result.Interactive:
			entry, ok = latestImageEntry(entries)
			if !ok {
				return "", fmt.Errorf("no image messages found in the active chat")
			}
		case result.Cancelled:
			return "", errImagePickerCancelled
		case result.Chosen == nil:
			return "", fmt.Errorf("no image messages found in the active chat")
		default:
			entry = result.Chosen
		}
	case strings.EqualFold(selector, "last"):
		entry, ok = latestImageEntry(entries)
		if !ok {
			return "", fmt.Errorf("no image messages found in the active chat")
		}
	default:
		messageID, err := parseMessageID(selector)
		if err == nil {
			entry, ok = findImageEntryByID(entries, messageID)
			if !ok {
				return "", fmt.Errorf("image message %d not found in the active chat transcript", messageID)
			}
			break
		}

		result := cli.pickImageEntry("Open image", selector, entries)
		switch {
		case !result.Interactive:
			return "", fmt.Errorf("image picker requires an interactive terminal; use \\openimage last")
		case result.Cancelled:
			return "", errImagePickerCancelled
		case result.Chosen == nil:
			return "", fmt.Errorf("no image messages match %q in the active chat", selector)
		default:
			entry = result.Chosen
		}
	}

	if entry.Image == nil {
		return "", fmt.Errorf("selected message does not contain an image")
	}

	path, err := cli.ensureImageDownloaded(ctx, target, *entry)
	if err != nil {
		return "", err
	}
	if err := openLocalPath(path); err != nil {
		return path, err
	}
	return path, nil
}

func (cli *TelegramCLI) recordOutgoingImage(target string, label string, messageID int64, attachment *ImageAttachment, caption string, reply *ReplyReference) {
	if attachment == nil {
		return
	}

	now := time.Now()
	body := messageBodyText(messageID, caption, attachment, reply)
	preview := imagePreviewText(attachment, caption)

	cli.setCurrentChat(target, label)
	cli.markChatActivity(target, preview, now)
	entry := legacyTranscriptEntry{
		MessageID: messageID,
		Outgoing:  true,
		Sender:    "You",
		Header:    outgoingTranscriptHeader(label, target, false),
		Body:      body,
		Meta:      outgoingTranscriptMeta(now.Format("15:04:05")),
		Text:      strings.TrimSpace(caption),
		Preview:   preview,
		Reply:     cloneReplyReference(reply),
		Image:     attachment,
	}
	cli.appendLegacyTranscriptEntry(target, entry)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func ensureContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
