package models

// Chat is information about a chat where the event announcement was posted.
type Chat struct {
	Profile          // Telegram profile
	Type    ChatType `json:"type"`  // Chat type
	Title   string   `json:"title"` // Chat title
}

type ChatType string

const (
	ChatPrivate ChatType = "private"
	ChatGroup   ChatType = "group"
	ChatChannel ChatType = "channel"
	ChatSuper   ChatType = "supergroup"
)
