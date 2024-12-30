package views

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

var (
	BtnEventSettings         = tele.Btn{Unique: "evt_set"}
	BtnEventSettingsAutoPair = tele.Btn{Unique: "evt_set_auto_pair"}
	BtnEventSettingsLimit    = tele.Btn{Unique: "evt_set_lim"}
	BtnEventSettingsLimitNum = tele.Btn{Unique: "evt_set_lim_num"}
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
			"0", // page number
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

// BtnEventSettingsLimitScene creates a buttons for the event settings limit scene
func BtnEventSettingsLimitScene(eventID string, page int, offset string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}

	// Add no limit button
	rows := []tele.Row{
		rm.Row(rm.Data(
			locale.BtnEventSettingsLimitNone,
			BtnEventSettingsLimitNum.Unique,
			eventID,
			"0",
			offset,
			randtoken.New(2),
		)),
	}

	// Add 2 rows with 5 buttons each
	for i := 0; i < 2; i++ {
		var row tele.Row
		for j := 0; j < 5; j++ {
			limit := strconv.Itoa(page*10 + i*5 + j + 1)
			row = append(row, rm.Data(
				limit,
				BtnEventSettingsLimitNum.Unique,
				eventID,
				limit,
				offset,
			))
		}
		rows = append(rows, row)
	}

	// Add 'more' or 'less' button
	pageStr := "0"
	caption := locale.BtnEventSettingsLimitLess
	if page == 0 {
		pageStr = "1"
		caption = locale.BtnEventSettingsLimitMore
	}
	rows = append(rows, rm.Row(rm.Data(
		caption,
		BtnEventSettingsLimit.Unique,
		eventID,
		pageStr,
		offset,
		randtoken.New(2),
	)))

	rm.Inline(rows...)
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
