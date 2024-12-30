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
	"github.com/ofstudio/dancegobot/pkg/trace"
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
	if profile.ID == 0 {
		return nil, errors.New("profile ID is 0")
	}

	user, err := s.store.UserGet(ctx, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// If user exists and the profile is the same, return the user
	if user != nil && !s.profileUpdated(user, profile) {
		return user, nil
	}

	// Otherwise, upsert the user profile
	user = &models.User{
		Profile: profile,
	}
	if err = s.store.UserUpsertProfile(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to upsert user profile: %w", err)
	}
	s.log.Info("[user service] user profile upserted",
		"profile", profile,
		trace.Attr(ctx))

	return user, nil
}

// UpdateSession updates the user session.
func (s *UserService) UpdateSession(ctx context.Context, user *models.User) error {
	if err := s.store.UserUpdateSession(ctx, user); err != nil {
		return fmt.Errorf("failed to update user session: %w", err)
	}
	return nil
}

// UpdateSettings updates the user settings.
func (s *UserService) UpdateSettings(ctx context.Context, user *models.User) error {
	if err := s.store.UserUpdateSettings(ctx, user); err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	if err := s.store.HistoryCreate(ctx, &models.HistoryItem{
		Action:    models.HistoryUserSettingsUpdated,
		Initiator: &user.Profile,
		Details:   user.Settings,
		CreatedAt: nowFn(),
	}); err != nil {
		s.log.Error("[user service] failed to create history item"+err.Error(),
			"profile", user.Profile,
			trace.Attr(ctx))
	}
	return nil
}

// profileUpdated returns true if the user profile updated.
func (s *UserService) profileUpdated(user *models.User, profile models.Profile) bool {
	return user == nil ||
		user.Profile.FirstName != profile.FirstName ||
		user.Profile.LastName != profile.LastName ||
		user.Profile.Username != profile.Username
}
