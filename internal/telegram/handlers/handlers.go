package handlers

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

type Handlers struct {
	cfg          config.Settings
	eventService *services.EventService
	userService  *services.UserService
	log          *slog.Logger
}

func NewHandlers(cfg config.Settings, eventService *services.EventService, userService *services.UserService) *Handlers {
	return &Handlers{
		cfg:          cfg,
		eventService: eventService,
		userService:  userService,
		log:          noplog.Logger(),
	}
}

func (h *Handlers) WithLogger(l *slog.Logger) *Handlers {
	h.log = l
	return h
}

// ctx returns the context from the telebot context.
// If the context is not set, it returns a new context.Background().
func (h *Handlers) ctx(c tele.Context) context.Context {
	ctx, ok := c.Get("ctx").(context.Context)
	if !ok {
		ctx = context.Background()
	}
	return ctx
}

// userGet returns user from the telebot context.
func (h *Handlers) userGet(c tele.Context) *models.User {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		h.log.Error("[handlers] user not found in context", telelog.Trace(c))
	}
	return user
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
		h.userUpdateSession(c, u)
	}
	return u, nil
}

// userUpdateSession updates user session.
func (h *Handlers) userUpdateSession(c tele.Context, user *models.User) {
	if err := h.userService.UpdateSession(h.ctx(c), user); err != nil {
		h.log.Error("[handlers] failed to update user session: "+err.Error(), telelog.Trace(c))
	}
	c.Set("user", user)
}

// userUpdateSettings updates user settings.
func (h *Handlers) userUpdateSettings(c tele.Context, user *models.User) {
	if err := h.userService.UpdateSettings(h.ctx(c), user); err != nil {
		h.log.Error("[handlers] failed to update user settings: "+err.Error(), telelog.Trace(c))
	}
	c.Set("user", user)
}

// Text - handles text messages.
func (h *Handlers) Text(c tele.Context) error {
	h.log.Info("[handlers] text message received", "text", c.Text(), telelog.Attr(c))
	u := h.userGet(c)
	switch {
	case c.Text() == locale.BtnClose: // Reset user session on close button
		u.Session = models.Session{}
		h.userUpdateSession(c, u)
		return views.SendCloseOK(c)
	case u.Session.Action == models.SessionSignup: // Handle signup scene
		return h.signupText(c)
	default:
		h.log.Info("[handlers] unexpected text", telelog.Trace(c))
		return nil
	}
}

// sendErr sends an error message.
// It resets user session and removes the reply keyboard.
func (h *Handlers) sendErr(c tele.Context, msg string) error {
	// clear user session
	u := h.userGet(c)
	u.Session = models.Session{}
	h.userUpdateSession(c, u)
	return c.Send(msg, tele.RemoveKeyboard)
}

// respondErr responds callback with an error message.
// It resets user session and removes the reply keyboard.
func (h *Handlers) respondErr(c tele.Context, msg string) error {
	// clear user session
	u := h.userGet(c)
	u.Session = models.Session{}
	h.userUpdateSession(c, u)
	return c.RespondAlert(msg)
}
