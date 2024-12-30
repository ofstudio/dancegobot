package handlers

import (
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// CbEventSettings - handles event settings callback button.
func (h *Handlers) CbEventSettings(c tele.Context) error {
	h.log.Info("[handlers] event settings callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	eventID := c.Args()[0]
	offset := c.Args()[1]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings callback: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	if !h.eventService.CanManage(event, &h.userGet(c).Profile) {
		h.log.Error("[handlers] event settings callback: user can't manage the event",
			"event_id", eventID,
			"profile", u.Profile.LogValue(),
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	_ = c.Respond()
	return c.Edit(views.EventSettingsMsg(event), views.BtnEventSettingsScene(event, offset),
		tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// CbEventSettingsBack - handles event settings scene back callback button.
func (h *Handlers) CbEventSettingsBack(c tele.Context) error {
	h.log.Info("[handlers] event_settings_back callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] event_settings_back callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	offset, _ := strconv.Atoi(c.Args()[0])

	text, rm, err := h.myScene(c, offset)
	if err != nil {
		h.log.Error("[handlers] failed to get my scene: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	_ = c.Respond()

	return c.Edit(text, rm, tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
}

// CbEventSettingsToggles - handles event settings scene callback buttons toggles.
func (h *Handlers) CbEventSettingsToggles(c tele.Context) error {
	h.log.Info("[handlers] event settings toggle callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings toggle callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	eventID := c.Args()[0]
	offset := c.Args()[1]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings toggle callback: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	var unique string
	if c.Callback() != nil {
		unique = c.Callback().Unique
	}

	switch unique {
	case views.BtnEventSettingsAutoPair.Unique:
		event.Settings.AutoPairing = !event.Settings.AutoPairing
	case views.BtnEventSettingsClose.Unique:
		event.Settings.Closed = !event.Settings.Closed
	default:
		h.log.Error("[handlers] event settings toggle callback: unknown unique",
			"unique", unique,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	if event, err = h.eventService.UpdateSettings(h.ctx(c), event.ID, &u.Profile, event.Settings); err != nil {
		h.log.Error("[handlers] event settings toggle callback: failed to update event settings: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	h.log.Info("[handlers] event settings updated",
		"event", event.LogValue(),
		"settings", event.Settings.LogValue(),
		telelog.Trace(c))

	_ = c.Respond()
	return c.Edit(views.EventSettingsMsg(event), views.BtnEventSettingsScene(event, offset),
		tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}
