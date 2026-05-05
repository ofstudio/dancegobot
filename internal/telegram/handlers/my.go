package handlers

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// My - handles /my  command.
func (h *Handlers) My(c tele.Context) error {
	h.log.Info("[handlers] /my received", telelog.Attr(c))

	// Force update my events
	if _, err := h.userGetMyEvents(c, true); err != nil {
		h.log.Error("[handlers] /my: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	// Send the /my scene
	if err := h.myScene(c, 0); err != nil {
		h.log.Error("[handlers] /my: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	return nil
}

// CbMyTurnPage - handles /my scene pagination.
func (h *Handlers) CbMyTurnPage(c tele.Context) error {
	h.log.Info("[handlers] my scene turn page callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] my scene turn page callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	offset, _ := strconv.Atoi(c.Args()[0])
	if err := h.myScene(c, offset); err != nil {
		h.log.Error("[handlers] my scene turn page callback: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	_ = c.Respond()
	return nil
}

// CbMyRefresh - handles signup refresh callback button in /my scene.
func (h *Handlers) CbMyRefresh(c tele.Context) error {
	h.log.Info("[handlers] my scene refresh callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] my scene refresh callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	offset, _ := strconv.Atoi(c.Args()[0])
	if _, err := h.userGetMyEvents(c, true); err != nil {
		h.log.Error("[handlers] my scene refresh callback: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	if err := h.myScene(c, offset); err != nil {
		h.log.Error("[handlers] my scene refresh callback: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	_ = c.Respond()
	return nil
}

// myScene returns text and reply markup for the /my scene.
func (h *Handlers) myScene(c tele.Context, offset int) error {
	u, err := h.userGetMyEvents(c)
	if err != nil {
		return fmt.Errorf("my scene failed: %w", err)
	}

	// If no events, return no events message
	if len(u.Session.MyEvents) == 0 {
		return views.MySceneNoEvents(c)
	}

	// If offset is negative set it to the first event
	if offset < 0 {
		offset = 0
	}

	// If offset is greater than the number of events set it to the last event
	if offset >= len(u.Session.MyEvents) {
		offset = len(u.Session.MyEvents) - 1
	}
	eventID := u.Session.MyEvents[offset]

	next := offset + 1
	// If no more events, set next to zero
	if next >= len(u.Session.MyEvents) {
		next = 0
	}

	// Get the event and drop stale session items.
	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		return fmt.Errorf("my scene failed: %w", err)
	}
	if event == nil || event.Removed {
		u.Session.MyEvents = append(u.Session.MyEvents[:offset], u.Session.MyEvents[offset+1:]...)
		h.userSessionUpdate(c, u)
		return h.myScene(c, offset)
	}

	reg := services.NewEventHandler(event).RegistrationGet(models.Dancer{
		Profile:  &u.Profile,
		FullName: u.Profile.FullName(),
		Role:     models.RoleUnknown,
	})
	canManage := h.eventService.CanManage(event, u.Profile)
	return views.MyScene(c, reg, canManage, offset, next)
}

// userGetMyEvents returns user with MyEvents in Session.
// If user has no events in the session, gets them from the database and saves to the session.
func (h *Handlers) userGetMyEvents(c tele.Context, force ...bool) (*models.User, error) {
	u := h.userGet(c)
	if len(u.Session.MyEvents) == 0 || (len(force) > 0 && force[0]) {
		ids, err := h.eventService.GetMy(h.ctx(c), &u.Profile)
		if err != nil {
			return nil, err
		}
		u.Session.MyEvents = ids
		h.userSessionUpdate(c, u)
	}
	return u, nil
}
