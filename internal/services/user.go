package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
)

// UserService is a service that manages users.
type UserService struct {
	cfg   config.Settings
	store Store
	log   *slog.Logger
}

func NewUserService(cfg config.Settings, store Store) *UserService {
	return &UserService{
		cfg:   cfg,
		store: store,
		log:   noplog.Logger(),
	}
}

func (s *UserService) WithLogger(l *slog.Logger) *UserService {
	s.log = l
	return s
}

// Get returns a user by profile.
// If the user does not exist, it returns a [*models.User] with the given profile.
func (s *UserService) Get(ctx context.Context, profile models.Profile) (*models.User, error) {
	user, err := s.store.UserGet(ctx, profile.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &models.User{Profile: profile, CreatedAt: nowFn()}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// Upsert inserts or updates a user.
func (s *UserService) Upsert(ctx context.Context, user *models.User) error {
	if err := s.store.UserUpsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}
