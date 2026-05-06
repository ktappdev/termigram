package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gotd/td/tg"
)

func messageReplyHeader(msg *tg.Message) *tg.MessageReplyHeader {
	if msg == nil {
		return nil
	}
	replyTo, ok := msg.GetReplyTo()
	if !ok || replyTo == nil {
		replyTo = msg.ReplyTo
		if replyTo == nil {
			return nil
		}
	}
	header, _ := replyTo.(*tg.MessageReplyHeader)
	return header
}

func replyMessageID(msg *tg.Message) (int64, bool) {
	if msg == nil {
		return 0, false
	}
	header := messageReplyHeader(msg)
	if header == nil {
		return 0, false
	}
	replyToID, ok := header.GetReplyToMsgID()
	if !ok {
		replyToID = header.ReplyToMsgID
	}
	if replyToID <= 0 {
		return 0, false
	}
	return int64(replyToID), true
}

func fallbackReplyReference(msg *tg.Message) *ReplyReference {
	replyToID, ok := replyMessageID(msg)
	if !ok {
		return nil
	}

	ref := &ReplyReference{MessageID: replyToID}
	header := messageReplyHeader(msg)
	if header == nil {
		return ref
	}

	if quote, ok := header.GetQuoteText(); ok {
		ref.Preview = normalizeReplyPreview(quote)
	} else if header.QuoteText != "" {
		ref.Preview = normalizeReplyPreview(header.QuoteText)
	}
	if ref.Preview == "" {
		if media, ok := header.GetReplyMedia(); ok {
			ref.Preview = normalizeReplyPreview(replyPreviewFromMedia(media))
		} else if header.ReplyMedia != nil {
			ref.Preview = normalizeReplyPreview(replyPreviewFromMedia(header.ReplyMedia))
		}
	}
	return ref
}

func replyPreviewFromMedia(media tg.MessageMediaClass) string {
	switch value := media.(type) {
	case *tg.MessageMediaPhoto:
		return "[image]"
	case *tg.MessageMediaDocument:
		doc, ok := value.Document.AsNotEmpty()
		if !ok {
			return "[media]"
		}
		if isSupportedImageDocument(doc) {
			name := replyDocumentName(doc)
			if name != "" {
				return "[image] " + name
			}
			return "[image]"
		}
	}
	return "[media]"
}

func replyDocumentName(doc *tg.Document) string {
	if doc == nil {
		return ""
	}
	for _, attr := range doc.Attributes {
		if filename, ok := attr.(*tg.DocumentAttributeFilename); ok {
			return filepath.Base(strings.TrimSpace(filename.FileName))
		}
	}
	return ""
}

func messageSenderInfo(cli *TelegramCLI, message *tg.Message) (int64, string) {
	if message == nil {
		return 0, "Unknown"
	}
	if message.GetOut() || message.Out {
		return 0, "You"
	}

	fromID := int64(0)
	fromName := "Unknown"
	from, _ := message.GetFromID()
	if from == nil {
		from = message.FromID
	}
	peerUser, ok := from.(*tg.PeerUser)
	if !ok {
		return fromID, fromName
	}

	fromID = peerUser.UserID
	if cli != nil {
		if user, found := cli.getUserByID(peerUser.UserID); found {
			fromName = strings.TrimSpace(user.FirstName + " " + user.LastName)
			if fromName == "" && user.Username != "" {
				fromName = "@" + user.Username
			}
		}
	}
	if fromName == "" || fromName == "Unknown" {
		fromName = fmt.Sprintf("User %d", peerUser.UserID)
	}
	return fromID, fromName
}

func replyReferenceFromMessage(cli *TelegramCLI, message *tg.Message) *ReplyReference {
	if message == nil {
		return nil
	}
	_, sender := messageSenderInfo(cli, message)
	attachment, _ := imageAttachmentFromMessage(message)
	return &ReplyReference{
		MessageID: int64(message.ID),
		Sender:    strings.TrimSpace(sender),
		Preview:   normalizeReplyPreview(messagePreviewText(message.Message, attachment)),
	}
}

func replyReferenceFromEntry(entry transcriptEntry) *ReplyReference {
	if entry.MessageID <= 0 {
		return nil
	}
	sender := strings.TrimSpace(entry.Sender)
	if sender == "" && entry.Outgoing {
		sender = "You"
	}
	return &ReplyReference{
		MessageID: entry.MessageID,
		Sender:    sender,
		Preview:   entryPreviewText(entry),
	}
}

func resolveReplyReferencesForMessages(ctx context.Context, cli *TelegramCLI, messages []*tg.Message) map[int64]*ReplyReference {
	refs := make(map[int64]*ReplyReference, len(messages))
	if len(messages) == 0 {
		return refs
	}

	byID := make(map[int64]*tg.Message, len(messages))
	missingSet := map[int64]struct{}{}
	for _, message := range messages {
		if message == nil {
			continue
		}
		byID[int64(message.ID)] = message
	}

	for _, message := range messages {
		if message == nil {
			continue
		}
		replyToID, ok := replyMessageID(message)
		if !ok {
			continue
		}
		if original, ok := byID[replyToID]; ok {
			refs[int64(message.ID)] = replyReferenceFromMessage(cli, original)
			continue
		}
		missingSet[replyToID] = struct{}{}
		refs[int64(message.ID)] = fallbackReplyReference(message)
	}

	if len(missingSet) == 0 {
		return refs
	}

	missingIDs := make([]int64, 0, len(missingSet))
	for id := range missingSet {
		missingIDs = append(missingIDs, id)
	}
	sort.Slice(missingIDs, func(i, j int) bool { return missingIDs[i] < missingIDs[j] })

	fetched, err := cli.fetchReplyMessagesFunc(ctx, cli, missingIDs)
	if err != nil {
		return refs
	}

	for _, message := range messages {
		if message == nil {
			continue
		}
		replyToID, ok := replyMessageID(message)
		if !ok {
			continue
		}
		if original, ok := fetched[replyToID]; ok {
			refs[int64(message.ID)] = replyReferenceFromMessage(cli, original)
			continue
		}
		if refs[int64(message.ID)] == nil {
			refs[int64(message.ID)] = &ReplyReference{MessageID: replyToID}
		}
	}

	return refs
}

func (cli *TelegramCLI) resolveReplyReferenceForTarget(ctx context.Context, target string, message *tg.Message) *ReplyReference {
	replyToID, ok := replyMessageID(message)
	if !ok {
		return nil
	}

	entries, _ := cli.transcriptStore.transcriptSnapshot(target)
	if entry, ok := findTranscriptEntryByID(entries, replyToID); ok {
		return replyReferenceFromEntry(*entry)
	}

	fetched, err := cli.fetchReplyMessagesFunc(ctx, cli, []int64{replyToID})
	if err == nil {
		if original, ok := fetched[replyToID]; ok {
			return replyReferenceFromMessage(cli, original)
		}
	}

	if fallback := fallbackReplyReference(message); fallback != nil {
		return fallback
	}
	return &ReplyReference{MessageID: replyToID}
}

func fetchReplyMessages(ctx context.Context, cli *TelegramCLI, ids []int64) (map[int64]*tg.Message, error) {
	result := make(map[int64]*tg.Message, len(ids))
	if cli == nil || cli.api == nil || len(ids) == 0 {
		return result, nil
	}

	inputIDs := make([]tg.InputMessageClass, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		inputIDs = append(inputIDs, &tg.InputMessageID{ID: int(id)})
	}
	if len(inputIDs) == 0 {
		return result, nil
	}

	response, err := cli.api.MessagesGetMessages(ctx, inputIDs)
	if err != nil {
		return result, err
	}

	users, messages, err := unpackMessagesResponse(response)
	if err != nil {
		return result, err
	}
	cli.cacheUsersFromClasses(users)
	for _, class := range messages {
		message, ok := class.(*tg.Message)
		if !ok {
			continue
		}
		result[int64(message.ID)] = message
	}
	return result, nil
}

func unpackMessagesResponse(response tg.MessagesMessagesClass) ([]tg.UserClass, []tg.MessageClass, error) {
	switch value := response.(type) {
	case *tg.MessagesMessages:
		return value.GetUsers(), value.GetMessages(), nil
	case *tg.MessagesMessagesSlice:
		return value.GetUsers(), value.GetMessages(), nil
	case *tg.MessagesChannelMessages:
		return value.GetUsers(), value.GetMessages(), nil
	default:
		return nil, nil, fmt.Errorf("unsupported messages response type: %T", response)
	}
}

func messageClassesToMessages(classes []tg.MessageClass) []*tg.Message {
	messages := make([]*tg.Message, 0, len(classes))
	for _, class := range classes {
		message, ok := class.(*tg.Message)
		if !ok {
			continue
		}
		messages = append(messages, message)
	}
	return messages
}

func findTranscriptEntryByID(entries []transcriptEntry, messageID int64) (*transcriptEntry, bool) {
	for _, entry := range entries {
		if entry.MessageID == messageID {
			copy := entry
			return &copy, true
		}
	}
	return nil, false
}
