package handlers

import (
	tele "gopkg.in/telebot.v4"

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
