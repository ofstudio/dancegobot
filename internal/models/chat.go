package models

// Chat is information about a chat where the event announcement was posted.
// Available only if bot is member of the chat.
type Chat struct {
	Profile            // Chat profile
	Type      ChatType `json:"type"`       // Chat type
	Title     string   `json:"title"`      // Chat title
	MessageID int      `json:"message_id"` // ID of the announcement message in the chat
}

type ChatType string

const (
	ChatPrivate ChatType = "private"
	ChatGroup   ChatType = "group"
	ChatChannel ChatType = "channel"
	ChatSuper   ChatType = "supergroup"
)
