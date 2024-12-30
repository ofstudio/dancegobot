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
	BtnMyTurnPage = tele.Btn{Unique: "my_turn_page"}
	BtnMyRefresh  = tele.Btn{Unique: "my_refresh"}
)

// BtnMyScene creates buttons for the /my scene.
func BtnMyScene(reg *models.Registration, canManage bool, offset, next int) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	var rows []tele.Row

	// add pagination buttons
	pagination := rm.Row()
	if offset > 0 {
		pagination = append(pagination, rm.Data(locale.BtnMyPrev, BtnMyTurnPage.Unique, strconv.Itoa(offset-1)))
	}
	if next > 0 {
		pagination = append(pagination, rm.Data(locale.BtnMyNext, BtnMyTurnPage.Unique, strconv.Itoa(next)))
	}
	if len(pagination) > 0 {
		rows = append(rows, pagination)
	}

	// add modify registration button
	modifySignup := btnMySceneModifySignup(rm, reg, offset)
	if len(modifySignup) > 0 {
		rows = append(rows, modifySignup)
	}

	// add chat link button
	link, ok := chatLink(reg.Event)
	if ok {
		rows = append(rows, rm.Row(rm.URL(locale.BtnChatLink, link)))
	}

	// add settings button
	if canManage {
		rows = append(rows, rm.Row(rm.Data(
			locale.BtnEventSettings,
			BtnCbEventSettings.Unique,
			reg.Event.ID,
			strconv.Itoa(offset),
			strconv.Itoa(next),
			randtoken.New(2),
		)))
	}

	if len(rows) != 0 {
		rm.Inline(rows...)
	}
	return rm
}

// btnMySceneModifySignup returns a button row to modify the registration.
func btnMySceneModifySignup(rm *tele.ReplyMarkup, reg *models.Registration, offset int) tele.Row {
	// If user can't register return nil
	if reg.Event.Settings.Closed {
		return nil
	}

	refresh := rm.Data(
		locale.BtnSignupRefresh,
		BtnMyRefresh.Unique,
		strconv.Itoa(offset),
		randtoken.New(2),
	)

	// If user is not registered return signup buttons
	if !reg.Status.IsRegistered() && reg.Status.CanRegister() {
		return rm.Row(
			rm.URL(locale.RoleIcon[models.RoleLeader], EventSignupURL(reg.Event.ID, models.RoleLeader)),
			rm.URL(locale.RoleIcon[models.RoleFollower], EventSignupURL(reg.Event.ID, models.RoleFollower)),
			refresh,
		)
	}

	// If user is registered return modify button
	if reg.Status.IsRegistered() {
		return rm.Row(
			rm.URL(locale.RoleIcon[reg.Role]+locale.BtnSignupModify, EventSignupURL(reg.Event.ID, reg.Role)),
			refresh,
		)
	}

	return nil
}

// MySceneMsg returns a message with the user event
func MySceneMsg(event *models.Event) string {
	sb := postTextBuilder(event)
	date := event.CreatedAt.Format("02.01.2006")
	return fmt.Sprintf(locale.MyEventHeader, date) + sb.String()
}
