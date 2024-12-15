package models

import "log/slog"

// Post represent an event post in a Telegram chat.
type Post struct {
	InlineMessageID string `json:"inline_message_id"`         // Post inline_message_id
	Chat            *Chat  `json:"chat,omitempty"`            // Chat where the post was published (only if bot is a member)
	ChatMessageID   int    `json:"chat_message_id,omitempty"` // ID of the post message in the chat (only if bot is a member)
}

// LogValue implements slog.Valuer interface for Post model.
func (p Post) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("inline_message_id", p.InlineMessageID),
	}
	if p.Chat != nil {
		attrs = append(attrs, slog.Any("chat", p.Chat), slog.Int("chat_message_id", p.ChatMessageID))
	}
	return slog.GroupValue(attrs...)
}
