package services

import (
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/store"
)

// Services is a service container.
type Services struct {
	Event        *EventService
	User         *UserService
	Subscription *SubscriptionService
	Notifier     *NotifierService
	Render       *RenderService
}

func NewServices(
	cfg config.Settings,
	store store.Store,
	renderFunc RenderFunc,
	notifyFunc NotifyFunc,
	membershipFunc MembershipFunc,
) *Services {
	render := NewRenderService(cfg, store, renderFunc)
	notifier := NewNotifierService(cfg, store, notifyFunc)
	subscription := NewSubscriptionService(store, notifier, membershipFunc)
	return &Services{
		Event:        NewEventService(cfg, store, render, notifier).WithEventPublishedHandler(subscription),
		User:         NewUserService(cfg, store),
		Subscription: subscription,
		Notifier:     notifier,
		Render:       render,
	}
}

func (s *Services) WithLogger(l *slog.Logger) *Services {
	s.Event.WithLogger(l)
	s.User.WithLogger(l)
	s.Subscription.WithLogger(l)
	s.Notifier.WithLogger(l)
	s.Render.WithLogger(l)
	return s
}
