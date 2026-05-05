package views

import (
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

// UserSettingsScene renders the user settings scene.
func UserSettingsScene(c tele.Context, settings models.UserSettings) error {
	text := locale.UserSettingsCaption +
		locale.UserSettingsDescription + "\n\n" +
		locale.EventSettingsAutoPair[settings.Event.AutoPairing] + "\n" +
		eventSettingsLimitText(settings.Event.Limit)
	rm := btnUserSettingsScene(settings)
	return c.EditOrSend(text, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// UserSettingsLimitScene renders the default event limit settings scene.
func UserSettingsLimitScene(c tele.Context, page int) error {
	return c.Edit(btnUserSettingsLimitScene(page), tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// UserSettingsHelp renders the user settings help message.
func UserSettingsHelp(c tele.Context) error {
	return c.EditOrSend(locale.UserSettingsHelp, btnUserSettingsBack(), tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

var (
	BtnUserSettingsAutoPair = tele.Btn{Unique: "usr_set_auto_pair"}
	BtnUserSettingsLimit    = tele.Btn{Unique: "usr_set_lim"}
	BtnUserSettingsLimitNum = tele.Btn{Unique: "usr_set_lim_num"}
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
			rm.Data(locale.BtnEventSettingsLimit,
				BtnUserSettingsLimit.Unique,
				"0",
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

// btnUserSettingsLimitScene creates default event limit settings buttons.
func btnUserSettingsLimitScene(page int) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		RemoveKeyboard: true,
	}

	rows := []tele.Row{
		rm.Row(rm.Data(
			locale.BtnEventSettingsLimitNone,
			BtnUserSettingsLimitNum.Unique,
			"0",
			randtoken.New(4),
		)),
	}

	for i := 0; i < 2; i++ {
		var row tele.Row
		for j := 0; j < 5; j++ {
			limit := strconv.Itoa(page*10 + i*5 + j + 1)
			row = append(row, rm.Data(
				limit,
				BtnUserSettingsLimitNum.Unique,
				limit,
				randtoken.New(4),
			))
		}
		rows = append(rows, row)
	}

	pageStr := "0"
	caption := locale.BtnEventSettingsLimitLess
	if page == 0 {
		pageStr = "1"
		caption = locale.BtnEventSettingsLimitMore
	}
	rows = append(rows, rm.Row(rm.Data(
		caption,
		BtnUserSettingsLimit.Unique,
		pageStr,
		randtoken.New(4),
	)))

	rm.Inline(rows...)
	return rm
}
