package models

// Chat is
type Chat struct {
	Profile
	Type  ChatType `json:"type"`
	Title string   `json:"title"`
}

type ChatType string

const (
	ChatPrivate ChatType = "private"
	ChatGroup   ChatType = "group"
	ChatChannel ChatType = "channel"
	ChatSuper   ChatType = "supergroup"
)
