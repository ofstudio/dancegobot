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
	store store.Store
	log   *slog.Logger
}

func NewUserService(cfg config.Settings, store store.Store) *UserService {
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
// If the user does not exist, it creates a new user with the given profile.
// If the user exists but the profile is different, it updates the profile.
func (s *UserService) Get(ctx context.Context, profile models.Profile) (*models.User, error) {
	user, err := s.store.UserGet(ctx, profile.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// If the user does not exist or the profile should be updated, upsert the user
	if s.shouldUpdate(user, profile) {
		user = &models.User{
			Profile: profile,
		}
		if err = s.store.UserUpsertProfile(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to upsert user profile: %w", err)
		}
		s.log.Info("[user service] user upserted", "profile", profile)
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

// shouldUpdate returns true if the user should be updated.
func (s *UserService) shouldUpdate(user *models.User, profile models.Profile) bool {
	return user == nil ||
		user.Profile.FirstName != profile.FirstName ||
		user.Profile.LastName != profile.LastName ||
		user.Profile.Username != profile.Username
}
