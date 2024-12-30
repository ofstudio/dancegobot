package handlers

import (
	"errors"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// CbMyTurnPage - handles /my scene pagination.
func (h *Handlers) CbMyTurnPage(c tele.Context) error {
	h.log.Info("[handlers] my scene turn page callback received", telelog.Attr(c))
	offset := 0
	if len(c.Args()) > 0 {
		offset, _ = strconv.Atoi(c.Args()[0])
	}
	text, rm, err := h.myScene(c, offset)
	if err != nil {
		h.log.Error("[handlers] my scene turn page callback: failed to get my scene: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	_ = c.Respond()
	return c.Edit(text, rm, tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
}

// CbMyRefresh - handles signup refresh callback button in /my scene.
func (h *Handlers) CbMyRefresh(c tele.Context) error {
	h.log.Info("[handlers] my scene refresh callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] my scene refresh callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	offset, _ := strconv.Atoi(c.Args()[0])

	msg, rm, err := h.myScene(c, offset)
	if err != nil {
		h.log.Error("[handlers] my scene refresh callback: failed to get my scene: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	_ = c.Respond()

	return c.Edit(msg, rm, tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
}

// My - handles /my  command.
func (h *Handlers) My(c tele.Context) error {
	h.log.Info("[handlers] /my received", telelog.Attr(c))

	// Force update my events
	_, err := h.userGetMyEvents(c, true)
	if err != nil {
		h.log.Error("[handlers] /my: failed to get my events: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	// Send the /my scene
	text, rm, err := h.myScene(c, 0)
	if err != nil {
		h.log.Error("[handlers] /my: failed to get my scene: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	return c.Send(text, rm, tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
}

// myScene returns text and reply markup for the /my scene.
func (h *Handlers) myScene(c tele.Context, offset int) (string, *tele.ReplyMarkup, error) {
	u, err := h.userGetMyEvents(c, false)
	if err != nil {
		return "", nil, err
	}

	// If no events, return no events message
	if len(u.Session.MyEvents) == 0 {
		return locale.MyNoEvents, views.BtnTry(), nil
	}

	// If offset is greater than the number of events set it to the last event
	if offset >= len(u.Session.MyEvents) {
		offset = len(u.Session.MyEvents) - 1
	}

	next := offset + 1
	// If no more events, set next to zero
	if next >= len(u.Session.MyEvents) {
		next = 0
	}

	// Get the event
	event, err := h.eventService.Get(h.ctx(c), u.Session.MyEvents[offset])
	if err != nil {
		return "", nil, err
	}
	if event == nil {
		return "", nil, errors.New("event not found")
	}

	// If event marked as removed, delete it from the session MyEvents list
	if event.Removed {
		u.Session.MyEvents = append(u.Session.MyEvents[:offset], u.Session.MyEvents[offset+1:]...)
		h.userUpdateSession(c, u)
		return h.myScene(c, offset)
	}

	canManage := h.eventService.CanManage(event, &u.Profile)
	reg := h.eventService.RegistrationGet(event, &u.Profile, models.RoleLeader) // role doesn't matter here

	return views.MySceneMsg(event), views.BtnMyScene(reg, canManage, offset, next), nil
}
