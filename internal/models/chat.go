package models

import tele "gopkg.in/telebot.v4"

// Chat is information about a chat where the event announcement was posted.
type Chat struct {
	ID       int64    `json:"id"`                 // Chat ID
	Username string   `json:"username,omitempty"` // Chat username
	Type     ChatType `json:"type"`               // Chat type
	Title    string   `json:"title"`              // Chat title
}

func NewChat(c *tele.Chat) Chat {
	var t ChatType
	switch c.Type {
	case tele.ChatGroup:
		t = ChatGroup
	case tele.ChatSuperGroup:
		t = ChatSuper
	case tele.ChatChannel:
		t = ChatChannel
	case tele.ChatChannelPrivate:
		t = ChatChannel
	default:
		t = ChatPrivate
	}
	return Chat{
		ID:       c.ID,
		Username: c.Username,
		Type:     t,
		Title:    c.Title,
	}
}

type ChatType string

const (
	ChatPrivate ChatType = "private"
	ChatGroup   ChatType = "group"
	ChatChannel ChatType = "channel"
	ChatSuper   ChatType = "supergroup"
)
