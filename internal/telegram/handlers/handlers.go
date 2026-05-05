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

// Text - handles text messages.
func (h *Handlers) Text(c tele.Context) error {
	h.log.Info("[handlers] text message received", "text", c.Text(), telelog.Attr(c))
	u := h.userGet(c)
	switch {
	case c.Text() == locale.BtnClose: // Reset user session on close button
		u.Session = models.Session{}
		h.userSessionUpdate(c, u)
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
	h.userSessionUpdate(c, u)
	return c.Send(msg, tele.RemoveKeyboard)
}

// respondErr responds callback with an error message.
// It resets user session and removes the reply keyboard.
func (h *Handlers) respondErr(c tele.Context, msg string) error {
	// clear user session
	u := h.userGet(c)
	u.Session = models.Session{}
	h.userSessionUpdate(c, u)
	return c.RespondAlert(msg)
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

// userSessionUpdate updates user session.
func (h *Handlers) userSessionUpdate(c tele.Context, user *models.User) {
	if err := h.userService.UpdateSession(h.ctx(c), user); err != nil {
		h.log.Error("[handlers] "+err.Error(), telelog.Trace(c))
	}
	c.Set("user", user)
}

// userSettingsUpdate updates user settings.
func (h *Handlers) userSettingsUpdate(c tele.Context, user *models.User) {
	if err := h.userService.UpdateSettings(h.ctx(c), user); err != nil {
		h.log.Error("[handlers] "+err.Error(), telelog.Trace(c))
	}
	c.Set("user", user)
}
