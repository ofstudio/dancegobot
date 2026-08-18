package views

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

var (
	BtnMyTurnPage     = tele.Btn{Unique: "my_turn_page"}
	BtnMyRefresh      = tele.Btn{Unique: "my_refresh"}
	BtnMySubscription = tele.Btn{Unique: "my_subscription"}
)

const (
	SubscriptionActionSubscribe   = "subscribe"
	SubscriptionActionUnsubscribe = "unsubscribe"
)

// MyScene renders the /my scene.
func MyScene(
	c tele.Context,
	reg models.Registration,
	canManage bool,
	subscription services.SubscriptionStatus,
	offset, next int,
) error {
	date := reg.Event.CreatedAt.Format("02.01.2006")
	msg := fmt.Sprintf(locale.MyEventHeader, date) + postTextBuilder(reg.Event).String()
	rm := btnMyScene(reg, canManage, subscription, offset, next)
	err := c.EditOrSend(msg, rm, tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
	if editErrorIsSuccess(err) {
		return nil
	}
	return err
}

// MySceneNoEvents renders a message that there are no events for /my scene.
func MySceneNoEvents(c tele.Context) error {
	return c.EditOrSend(locale.MyNoEvents, btnTry(), tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
}

// btnMyScene creates buttons for the /my scene.
func btnMyScene(
	reg models.Registration,
	canManage bool,
	subscription services.SubscriptionStatus,
	offset, next int,
) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	var rows []tele.Row

	// add pagination buttons
	pagination := rm.Row()
	if offset > 0 {
		pagination = append(pagination, rm.Data(
			locale.BtnMyPrev,
			BtnMyTurnPage.Unique,
			strconv.Itoa(offset-1),
			randtoken.New(2),
		))
	}
	if next > 0 {
		pagination = append(pagination, rm.Data(
			locale.BtnMyNext,
			BtnMyTurnPage.Unique,
			strconv.Itoa(next),
			randtoken.New(2),
		))
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

	if subscription.Available {
		action := SubscriptionActionSubscribe
		caption := locale.BtnSubscribe
		if subscription.Subscribed {
			action = SubscriptionActionUnsubscribe
			caption = locale.BtnUnsubscribe
		}
		rows = append(rows, rm.Row(rm.Data(
			caption,
			BtnMySubscription.Unique,
			action,
			reg.Event.ID,
			strconv.Itoa(offset),
		)))
	}

	// add settings button
	if canManage {
		rows = append(rows, rm.Row(rm.Data(
			locale.BtnEventSettings,
			BtnEventSettings.Unique,
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
func btnMySceneModifySignup(rm *tele.ReplyMarkup, reg models.Registration, offset int) tele.Row {
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
