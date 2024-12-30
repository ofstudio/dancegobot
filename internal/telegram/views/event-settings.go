package views

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

var (
	BtnCbEventSettings         = tele.Btn{Unique: "event_settings"}
	BtnCbEventSettingsAutoPair = tele.Btn{Unique: "event_settings_auto_pair"}
	BtnCbEventSettingsLimit    = tele.Btn{Unique: "event_settings_limit"}
	BtnCbEventSettingsClosed   = tele.Btn{Unique: "event_settings_closed_for"}
	BtnCbEventSettingsBack     = tele.Btn{Unique: "event_settings_back"}
)

// BtnEventSettings creates a buttons for the event settings
func BtnEventSettings(event *models.Event, offset string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(rm.Data(
			locale.BtnEventSettingsAutoPair[event.Settings.AutoPairing],
			BtnCbEventSettingsAutoPair.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnEventSettingsLimit,
			BtnCbEventSettingsLimit.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnEventSettingsClosed[event.Settings.Closed],
			BtnCbEventSettingsClosed.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnBack,
			BtnCbEventSettingsBack.Unique,
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
