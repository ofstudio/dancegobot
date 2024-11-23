package services

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
)

// Services is a service container.
type Services struct {
	Event    *EventService
	User     *UserService
	Notifier *NotifierService
	Render   *RenderService
}

func NewServices(cfg config.Settings, store Store, bot tele.API) *Services {
	render := NewRenderService(cfg, store, bot)
	notifier := NewNotifierService(cfg, store, bot)
	return &Services{
		Event:    NewEventService(cfg, store, render, notifier),
		User:     NewUserService(cfg, store),
		Notifier: notifier,
		Render:   render,
	}
}

func (s *Services) WithLogger(l *slog.Logger) *Services {
	s.Event.WithLogger(l)
	s.User.WithLogger(l)
	s.Notifier.WithLogger(l)
	s.Render.WithLogger(l)
	return s
}