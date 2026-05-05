package services

import (
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/store"
)

// Services is a service container.
type Services struct {
	Event    *EventService
	User     *UserService
	Notifier *NotifierService
	Render   *RenderService
}

func NewServices(cfg config.Settings, store store.Store, renderFunc RenderFunc, notifyFunc NotifyFunc) *Services {
	render := NewRenderService(cfg, store, renderFunc)
	notifier := NewNotifierService(cfg, store, notifyFunc)
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
