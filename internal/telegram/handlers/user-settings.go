package handlers

import (
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// UserSettingsScene - handles /settings command.
func (h *Handlers) UserSettingsScene(c tele.Context) error {
	h.log.Info("[handlers] /settings received", telelog.Attr(c))
	u := h.userGet(c)
	return views.UserSettingsScene(c, u.Settings)
}

// CbUserSettingsAutoPair - toggles auto pair user setting.
func (h *Handlers) CbUserSettingsAutoPair(c tele.Context) error {
	h.log.Info("[handlers] user settings auto pair callback received", telelog.Attr(c))
	u := h.userGet(c)
	u.Settings.Event.AutoPairing = !u.Settings.Event.AutoPairing
	h.log.Info("[handlers] user settings updated",
		"settings", u.Settings.LogValue(),
		telelog.Trace(c))
	h.userSettingsUpdate(c, u)
	_ = c.Respond()
	return views.UserSettingsScene(c, u.Settings)
}

// CbUserSettingsLimitScene - opens default event limit settings scene.
func (h *Handlers) CbUserSettingsLimitScene(c tele.Context) error {
	h.log.Info("[handlers] user settings limit callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] user settings limit callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	page, _ := strconv.Atoi(c.Args()[0])
	if page < 0 || page > 1 {
		h.log.Error("[handlers] user settings limit callback: invalid page",
			"page", page,
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	_ = c.Respond()
	return views.UserSettingsLimitScene(c, page)
}

// CbUserSettingsLimitNum - updates default event limit user setting.
func (h *Handlers) CbUserSettingsLimitNum(c tele.Context) error {
	h.log.Info("[handlers] user settings limit number callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] user settings limit number callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	limit, _ := strconv.Atoi(c.Args()[0])
	if limit < 0 || limit > 20 {
		h.log.Error("[handlers] user settings limit number callback: invalid limit",
			"limit", limit,
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	u.Settings.Event.Limit = limit
	h.log.Info("[handlers] user settings updated",
		"settings", u.Settings.LogValue(),
		telelog.Trace(c))
	h.userSettingsUpdate(c, u)
	_ = c.Respond()
	return views.UserSettingsScene(c, u.Settings)
}

// CbUserSettingsHelp - sends settings help message.
func (h *Handlers) CbUserSettingsHelp(c tele.Context) error {
	h.log.Info("[handlers] user settings help callback received", telelog.Attr(c))
	_ = c.Respond()
	return views.UserSettingsHelp(c)
}

// CbUserSettingsBack - sends settings scene message.
func (h *Handlers) CbUserSettingsBack(c tele.Context) error {
	h.log.Info("[handlers] user settings back callback received", telelog.Attr(c))
	u := h.userGet(c)
	_ = c.Respond()
	return views.UserSettingsScene(c, u.Settings)
}
