package middleware

import (
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/pkg/noplog"
)

// Middleware is a collection of middlewares.
type Middleware struct {
	cfg          config.Settings
	eventService *services.EventService
	userService  *services.UserService
	log          *slog.Logger
}

// NewMiddleware creates a new middleware collection.
func NewMiddleware(cfg config.Settings, eventService *services.EventService, userService *services.UserService) *Middleware {
	return &Middleware{
		cfg:          cfg,
		eventService: eventService,
		userService:  userService,
		log:          noplog.Logger(),
	}
}

func (m *Middleware) WithLogger(l *slog.Logger) *Middleware {
	m.log = l
	return m
}
