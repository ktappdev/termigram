package main

import (
	"strings"
	"time"

	"github.com/gotd/td/tg"
)

func messageOutputFromTGMessage(cli *TelegramCLI, message *tg.Message) MessageOutput {
	return messageOutputFromTGMessageWithReply(cli, message, fallbackReplyReference(message))
}

func messageOutputFromTGMessageWithReply(cli *TelegramCLI, message *tg.Message, reply *ReplyReference) MessageOutput {
	fromID, fromName := messageSenderInfo(cli, message)
	attachment, _ := imageAttachmentFromMessage(message)
	text := strings.TrimSpace(message.Message)
	preview := messagePreviewText(text, attachment)

	return MessageOutput{
		ID:       int64(message.ID),
		FromID:   fromID,
		FromName: fromName,
		Message:  messageBodyText(int64(message.ID), text, attachment, reply),
		Date:     int64(message.Date),
		Outgoing: message.GetOut(),
		Text:     text,
		Preview:  preview,
		Reply:    cloneReplyReference(reply),
		Image:    attachment,
	}
}

func transcriptEntryFromMessageOutput(target string, label string, msg MessageOutput) transcriptEntry {
	timestamp := time.Unix(msg.Date, 0).Format("15:04:05")
	entry := transcriptEntry{
		MessageID: msg.ID,
		Outgoing:  msg.Outgoing,
		Sender:    msg.FromName,
		Body:      msg.Message,
		Meta:      "",
		Text:      msg.Text,
		Preview:   msg.Preview,
		Reply:     cloneReplyReference(msg.Reply),
		Image:     msg.Image,
	}
	if msg.Outgoing {
		entry.Sender = "You"
		entry.Header = outgoingTranscriptHeader(label, target, false)
		entry.Meta = outgoingTranscriptMeta(timestamp)
	} else {
		entry.Header = incomingTranscriptHeader(msg.FromName, target, false)
		entry.Meta = incomingTranscriptMeta(timestamp)
	}
	return entry
}
