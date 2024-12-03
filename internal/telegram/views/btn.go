package views

import (
	"fmt"
	"math/rand"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/deeplink"
)

// btnTry creates a button for the "Try" option on the start message.
func btnTry() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Row(
		rm.Query(locale.BtnTry, " "),
	))
	return rm
}

// btnAnnouncement creates a buttons for the event announcement message.
func btnAnnouncement(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.URL(locale.RoleIcon[models.RoleLeader],
				deeplink.New(models.SessionSignup, eventID, models.RoleLeader.String())),
			rm.URL(locale.RoleIcon[models.RoleFollower],
				deeplink.New(models.SessionSignup, eventID, models.RoleFollower.String())),
		),
	)
	return rm
}

// btnSignup creates a buttons for the signup.
func btnSignup(dancer *models.Dancer, singles []models.SessionSingle) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		Placeholder:    locale.SignupPlaceholder,
	}
	var rows []tele.Row

	// if dancer can signup
	if dancer.Status.SignupAvailable() {
		// add user sharing button
		rows = append(rows, rm.Row(
			rm.User(locale.BtnSignupContact, &tele.ReplyRecipient{
				ID:              rand.Int31(),
				Quantity:        1,
				Bot:             tele.Flag(false),
				RequestName:     tele.Flag(true),
				RequestUsername: tele.Flag(true),
			})))
		// add event singles with the opposite role if any
		for _, s := range singles {
			rows = append(rows, rm.Row(rm.Text(s.Caption)))
		}
	}

	// add "signup as single" button if there are no singles in opposite role
	if len(singles) == 0 && dancer.Status == models.StatusNotRegistered {
		rows = append(rows, rm.Row(rm.Text(locale.BtnSingle[dancer.Role])))
	}

	// Add "remove" button if the dancer is signed up
	if dancer.Status.SignedUp() {
		rows = append(rows, rm.Row(rm.Text(locale.BtnRemove)))
	}

	// Add "close" button
	rows = append(rows, rm.Row(rm.Text(locale.BtnClose)))

	rm.Reply(rows...)
	return rm
}

// BtnChatLink creates an inline button with a link to the chat.
//
// Known Telegram limitations:
//   - Only supergroups and channels can be linked
//   - Supergroup or channel can be either public or private
//   - Bot should be a member of supergroup or an admin in the channel
//
// Link format:
//
//	https://t.me/c/{chat_link_id}/{message_id}
//
// Where {chat_link_id} = - {chat_id} - 1000000000000
//
// For example:
//
//	message_id:     1234
//	chat_id:       -1001234567890 (supergroup or channel)
//	chat_link_id:  -(-1001234567890) - 1000000000000 = 1234567890
//
// Which gives us the link: https://t.me/c/1234567890/1234
func BtnChatLink(event *models.Event) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	// skip if the chat is not set
	if event.Post.Chat == nil {
		return rm
	}
	// skip if the chat is not a supergroup or channelq
	if event.Post.Chat.Type != models.ChatSuper && event.Post.Chat.Type != models.ChatChannel {
		return rm
	}
	chatLinkId := -event.Post.Chat.ID - 1000000000000
	url := fmt.Sprintf("https://t.me/c/%d/%d", chatLinkId, event.Post.MessageID)
	rm.Inline(rm.Row(
		rm.URL(locale.BtnChatLink, url),
	))
	return rm

}
