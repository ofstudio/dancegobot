package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ofstudio/dancegobot/helpers"
	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
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
		log:   helpers.NopLogger(),
	}
}

func (s *UserService) WithLogger(l *slog.Logger) *UserService {
	s.log = l
	return s
}

func (s *UserService) Get(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.store.UserGet(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) ProfileUpsert(ctx context.Context, user *models.User) error {
	if err := s.store.UserProfileUpsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user profile: %w", err)
	}
	return nil
}

func (s *UserService) SessionUpsert(ctx context.Context, user *models.User) error {
	if err := s.store.UserSessionUpsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}
