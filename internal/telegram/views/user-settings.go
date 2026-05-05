package views

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

// UserSettingsScene renders the user settings scene.
func UserSettingsScene(c tele.Context, settings models.UserSettings) error {
	text := locale.UserSettingsCaption +
		locale.EventSettingsAutoPair[settings.Event.AutoPairing]
	rm := btnUserSettingsScene(settings)
	return c.EditOrSend(text, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// UserSettingsHelp renders the user settings help message.
func UserSettingsHelp(c tele.Context) error {
	return c.EditOrSend(locale.UserSettingsHelp, btnUserSettingsBack(), tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

var (
	BtnUserSettingsAutoPair = tele.Btn{Unique: "usr_set_auto_pair"}
	BtnUserSettingsHelp     = tele.Btn{Unique: "usr_set_help"}
	BtnUserSettingsBack     = tele.Btn{Unique: "usr_set_back"}
)

// btnUserSettingsScene creates buttons for the user settings scene.
func btnUserSettingsScene(settings models.UserSettings) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		RemoveKeyboard: true,
	}
	rm.Inline(
		rm.Row(
			rm.Data(locale.BtnEventSettingsAutoPair[settings.Event.AutoPairing],
				BtnUserSettingsAutoPair.Unique,
				randtoken.New(4)),
		),
		rm.Row(
			rm.Data(locale.BtnUserSettingsHelp, BtnUserSettingsHelp.Unique, randtoken.New(4)),
		),
	)
	return rm
}

// btnUserSettingsBack creates a button to return to the settings scene.
func btnUserSettingsBack() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		RemoveKeyboard: true,
	}
	rm.Inline(rm.Row(
		rm.Data(locale.BtnBack, BtnUserSettingsBack.Unique, randtoken.New(4)),
	))
	return rm
}
