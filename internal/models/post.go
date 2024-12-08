package models

// Post represent an event post in a Telegram chat.
type Post struct {
	InlineMessageID string `json:"inline_message_id"`         // Post inline_message_id
	Chat            *Chat  `json:"chat,omitempty"`            // Chat where the post was published (only if bot is a member)
	ChatMessageID   int    `json:"chat_message_id,omitempty"` // ID of the post message in the chat (only if bot is a member)
}
