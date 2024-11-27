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

// ProfileUpsert upserts the user profile.
func (s *UserService) ProfileUpsert(ctx context.Context, user *models.User) error {
	if err := s.store.UserProfileUpsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user profile: %w", err)
	}
	return nil
}

// SessionUpsert upserts the user session as well as the user profile.
func (s *UserService) SessionUpsert(ctx context.Context, user *models.User) error {
	if err := s.store.UserSessionUpsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}
