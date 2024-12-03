package models

// Post represent an event announcement post in a Telegram chat.
type Post struct {
	InlineMessageID string `json:"inline_message_id"`    // Announcement inline_message_id
	Chat            *Chat  `json:"chat,omitempty"`       // Chat where the announcement was posted (only if bot is a member)
	MessageID       int    `json:"message_id,omitempty"` // ID of the announcement message in the chat (only if bot is a member)
}
