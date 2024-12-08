package app

import tele "gopkg.in/telebot.v4"

var botUser = &tele.User{
	ID:        1234567890,
	IsBot:     true,
	FirstName: "test_bot",
	Username:  "test_bot",
}

var (
	userJohn = &tele.User{
		ID:        100,
		FirstName: "John",
	}
	userJane = &tele.User{
		ID:        101,
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "jane_doe",
	}
)

var (
	chatSuperGroup = &tele.Chat{
		ID:    -1001234567890,
		Type:  tele.ChatSuperGroup,
		Title: "Test Super Group",
	}
)

var (
	queryA = tele.Query{
		Sender:   userJohn,
		Text:     "Query A",
		ChatType: "supergroup",
	}

	queryB = tele.Query{
		Sender:   userJane,
		Text:     "Query B",
		ChatType: "channel",
	}
)

var (
	rxUrlSignupLeader   = `^https://t.me/` + botUser.Username + `\?start=.+-signup-.+leader$`
	rxUrlSignupFollower = `^https://t.me/` + botUser.Username + `\?start=.+-signup-.+follower$`
)
