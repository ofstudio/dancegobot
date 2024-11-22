package views

import (
	"math/rand"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/deeplink"
)

// btnAnnouncement creates a buttons for the event announcement message.
func btnAnnouncement(botName string, eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.URL(locale.RoleIcon[models.RoleLeader],
				deeplink.New(botName, models.SessionSignup, eventID, string(models.RoleLeader))),
			rm.URL(locale.RoleIcon[models.RoleFollower],
				deeplink.New(botName, models.SessionSignup, eventID, string(models.RoleFollower))),
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
