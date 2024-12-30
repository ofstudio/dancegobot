package views

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

var (
	BtnUserSettingsAutoPair = tele.Btn{Unique: "user_settings_auto_pair"}
	BtnUserSettingsHelp     = tele.Btn{Unique: "user_settings_help"}
	BtnUserSettingsBack     = tele.Btn{Unique: "user_settings_back"}
)

// btnUserSettingsScene creates buttons for the user settings scene.
func btnUserSettingsScene(settings *models.UserSettings) *tele.ReplyMarkup {
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

// UserSettingsScene returns a message with the user settings.
func UserSettingsScene(settings *models.UserSettings) (string, *tele.ReplyMarkup) {
	text := locale.UserSettingsCaption +
		locale.EventSettingsAutoPair[settings.Event.AutoPairing]
	rm := btnUserSettingsScene(settings)
	return text, rm
}

// UserSettingsHelp returns a message with the user settings help.
func UserSettingsHelp() (string, *tele.ReplyMarkup) {
	return locale.UserSettingsHelp, btnUserSettingsBack()
}
