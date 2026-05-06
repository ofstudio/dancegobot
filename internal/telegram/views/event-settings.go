package views

import (
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

// EventSettingsScene renders the event settings scene
func EventSettingsScene(c tele.Context, event *models.Event, offset int) error {
	date := event.CreatedAt.Format("02.01.2006")

	msg := fmt.Sprintf(locale.MyEventHeader, date) +
		fmtCaption(event.Caption) + "\n\n" +
		locale.EventSettingsCaption

	// Auto pairing setting
	msg += locale.EventSettingsAutoPair[event.Settings.AutoPairing] + "\n"

	// Limit setting
	msg += eventSettingsLimitText(event.Settings.Limit) + "\n"

	// Closed setting
	msg += locale.EventSettingsClosed[event.Settings.Closed]

	return c.EditOrSend(msg, btnEventSettingsScene(event, strconv.Itoa(offset)),
		tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

func eventSettingsLimitText(limit int) string {
	if limit <= 0 {
		return locale.EventSettingsLimitNone
	}
	return fmt.Sprintf(
		locale.EventSettingsLimit,
		locale.NumLimitCome.N(limit),
		limit,
		locale.NumLimitCouples.N(limit),
	)
}

// EventSettingsLimitScene renders the event settings limit scene.
// The page parameter is used to display the next or previous 10 limit buttons.
func EventSettingsLimitScene(c tele.Context, eventID string, page int, offset string) error {
	return c.Edit(btnEventSettingsLimitScene(eventID, page, offset),
		tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// SendLimitChanged sends notice about the event settings limit change.
// The oldLimit parameter is the previous limit value for the event.
// The increased parameter indicates whether the limit was increased or decreased.
// The idx parameter is the index of the first affected couple in the event couples list.
func SendLimitChanged(c tele.Context, event *models.Event, affected models.AffectedCouples) error {
	sb := &strings.Builder{}

	// Add notice about the limit change
	switch {
	case event.Settings.Limit == 0:
		sb.WriteString(locale.LimitChangedNoLimit)
	case affected.Increased:
		sb.WriteString(locale.LimitChangedIncreased)
	default:
		sb.WriteString(locale.LimitChangedDecreased)
	}
	if affected.Increased {
		sb.WriteString(fmt.Sprintf(
			locale.NumLimitIncreased.N(len(affected.Couples)),
			len(affected.Couples),
		))
	} else {
		sb.WriteString(fmt.Sprintf(
			locale.NumLimitDecreased.N(len(affected.Couples)),
			len(affected.Couples),
		))
	}

	// Add affected couples
	postCouplesBuild(sb, affected.Couples, 0, affected.Position)

	rm := btnLimitChanged(event.ID)
	return c.Send(sb.String(), rm, tele.RemoveKeyboard, tele.NoPreview, tele.ModeHTML)
}

// SendLimitChangedNotified sends a message that dancers were notified about the limit change.
func SendLimitChangedNotified(c tele.Context) error {
	return c.Send(locale.LimitChangedNotified, tele.RemoveKeyboard, tele.NoPreview, tele.ModeHTML)
}

var (
	BtnEventSettings         = tele.Btn{Unique: "evt_set"}
	BtnEventSettingsAutoPair = tele.Btn{Unique: "evt_set_auto_pair"}
	BtnEventSettingsLimit    = tele.Btn{Unique: "evt_set_lim"}
	BtnEventSettingsLimitNum = tele.Btn{Unique: "evt_set_lim_num"}
	BtnEventSettingsClose    = tele.Btn{Unique: "evt_set_close"}
	BtnEventSettingsBack     = tele.Btn{Unique: "evt_set_back"}
	BtnLimitChangedNotify    = tele.Btn{Unique: "lim_chg_ntf"}
	BtnLimitChangedSkip      = tele.Btn{Unique: "lim_chg_skp"}
)

// btnEventSettingsScene creates a buttons for the event settings scene
func btnEventSettingsScene(event *models.Event, offset string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		// Auto pairing toggle button
		rm.Row(rm.Data(
			locale.BtnEventSettingsAutoPair[event.Settings.AutoPairing],
			BtnEventSettingsAutoPair.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		// Event limit scene button
		rm.Row(rm.Data(
			locale.BtnEventSettingsLimit,
			BtnEventSettingsLimit.Unique,
			event.ID,
			"0", // page number
			offset,
			randtoken.New(2),
		)),
		// Event close toggle button
		rm.Row(rm.Data(
			locale.BtnEventSettingsClosed[event.Settings.Closed],
			BtnEventSettingsClose.Unique,
			event.ID,
			offset,
			randtoken.New(2),
		)),
		// Back button
		rm.Row(rm.Data(
			locale.BtnBack,
			BtnEventSettingsBack.Unique,
			offset,
			randtoken.New(2),
		)),
	)
	return rm
}

// btnEventSettingsLimitScene creates a buttons for the event settings limit scene.
// The page parameter is used to display the next or previous 10 limit buttons.
func btnEventSettingsLimitScene(eventID string, page int, offset string) *tele.ReplyMarkup {
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

// btnLimitChanged creates a buttons for the limit changed notice
func btnLimitChanged(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(rm.Data(
			locale.BtnLimitChangedNotify,
			BtnLimitChangedNotify.Unique,
			eventID,
			randtoken.New(2),
		)),
		rm.Row(rm.Data(
			locale.BtnLimitChangedSkip,
			BtnLimitChangedSkip.Unique,
		)),
	)
	return rm
}
