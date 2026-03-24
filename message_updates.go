package main

import "github.com/gotd/td/tg"

func sentMessageIDFromUpdates(updates tg.UpdatesClass) int64 {
	switch value := updates.(type) {
	case *tg.UpdateShortSentMessage:
		return int64(value.ID)
	case *tg.Updates:
		return sentMessageIDFromUpdateList(value.GetUpdates())
	case *tg.UpdatesCombined:
		return sentMessageIDFromUpdateList(value.GetUpdates())
	case *tg.UpdateShort:
		return sentMessageIDFromSingleUpdate(value.GetUpdate())
	default:
		return 0
	}
}

func sentMessageIDFromUpdateList(updates []tg.UpdateClass) int64 {
	for _, update := range updates {
		if id := sentMessageIDFromSingleUpdate(update); id > 0 {
			return id
		}
	}
	return 0
}

func sentMessageIDFromSingleUpdate(update tg.UpdateClass) int64 {
	switch value := update.(type) {
	case *tg.UpdateNewMessage:
		return int64(value.GetMessage().GetID())
	case *tg.UpdateNewChannelMessage:
		return int64(value.GetMessage().GetID())
	case *tg.UpdateMessageID:
		return int64(value.ID)
	default:
		return 0
	}
}

func messageFromShortUserUpdate(update *tg.UpdateShortMessage) *tg.Message {
	if update == nil {
		return nil
	}
	message := &tg.Message{
		ID:      update.GetID(),
		Date:    update.GetDate(),
		Message: update.GetMessage(),
		FromID: &tg.PeerUser{
			UserID: update.GetUserID(),
		},
	}
	if replyTo, ok := update.GetReplyTo(); ok {
		message.ReplyTo = replyTo
	} else if update.ReplyTo != nil {
		message.ReplyTo = update.ReplyTo
	}
	return message
}

func messageFromShortChatUpdate(update *tg.UpdateShortChatMessage) *tg.Message {
	if update == nil {
		return nil
	}
	message := &tg.Message{
		ID:      update.GetID(),
		Date:    update.GetDate(),
		Message: update.GetMessage(),
		FromID: &tg.PeerUser{
			UserID: update.GetFromID(),
		},
	}
	if replyTo, ok := update.GetReplyTo(); ok {
		message.ReplyTo = replyTo
	} else if update.ReplyTo != nil {
		message.ReplyTo = update.ReplyTo
	}
	return message
}
