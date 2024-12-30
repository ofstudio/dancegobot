package views

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

var (
	BtnEventSettings         = tele.Btn{Unique: "evt_set"}
	BtnEventSettingsAutoPair = tele.Btn{Unique: "evt_set_auto_pair"}
	BtnEventSettingsLimit    = tele.Btn{Unique: "evt_set_lim"}
	BtnEventSettingsClose    = tele.Btn{Unique: "evt_set_closed"}
	BtnEventSettingsBack     = tele.Btn{Unique: "evt_set_back"}
)

// BtnEventSettingsScene creates a buttons for the event settings scene
func BtnEventSettingsScene(event *models.Event, offset string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(rm.Data(
			locale.BtnEventSettingsAutoPair[event.Settings.AutoPairing],
			BtnEventSettingsAutoPair.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnEventSettingsLimit,
			BtnEventSettingsLimit.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnEventSettingsClosed[event.Settings.Closed],
			BtnEventSettingsClose.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnBack,
			BtnEventSettingsBack.Unique,
			offset,
			randtoken.New(2),
		)),
	)
	return rm
}

// EventSettingsMsg returns a message with the event settings
func EventSettingsMsg(event *models.Event) string {
	date := event.CreatedAt.Format("02.01.2006")
	msg := fmt.Sprintf(locale.MyEventHeader, date) +
		event.Caption + "\n\n" +
		locale.EventSettingsCaption +
		locale.EventSettingsAutoPair[event.Settings.AutoPairing] + "\n"

	if event.Settings.Limit > 0 {
		msg += fmt.Sprintf(
			locale.EventSettingsLimit,
			locale.NumLimitCome.N(event.Settings.Limit),
			event.Settings.Limit,
			locale.NameLimitCouples.N(event.Settings.Limit),
		) + "\n"
	} else {
		msg += locale.EventSettingsLimitNone + "\n"
	}

	return msg + locale.EventSettingsClosed[event.Settings.Closed]
}
