package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	storepkg "github.com/ofstudio/dancegobot/internal/store"
)

func TestEventServiceMissingEvent(t *testing.T) {
	ctx := context.Background()
	service, _ := newEventServiceTest(t)

	owner := models.Profile{ID: 1, FirstName: "Owner"}
	dancer := models.Profile{ID: 2, FirstName: "Dancer"}
	partner := models.Profile{ID: 3, FirstName: "Partner"}
	chat := models.Chat{ID: -100, Type: models.ChatSuper, Title: "Test"}

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "settings update",
			run: func() error {
				_, err := service.SettingsUpdate(ctx, "missing_event", owner, models.EventSettings{Limit: 1})
				return err
			},
		},
		{
			name: "post add",
			run: func() error {
				_, err := service.PostAdd(ctx, "missing_event", "inline_message_id")
				return err
			},
		},
		{
			name: "post chat add",
			run: func() error {
				_, err := service.PostChatAdd(ctx, "missing_event", &chat, 10)
				return err
			},
		},
		{
			name: "couple add",
			run: func() error {
				_, err := service.CoupleAdd(ctx, "missing_event", &dancer, models.RoleLeader, &partner)
				return err
			},
		},
		{
			name: "single add",
			run: func() error {
				_, err := service.SingleAdd(ctx, "missing_event", dancer, models.RoleLeader)
				return err
			},
		},
		{
			name: "dancer remove",
			run: func() error {
				_, err := service.DancerRemove(ctx, "missing_event", dancer)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				err := tt.run()
				require.Error(t, err)
				require.Contains(t, err.Error(), "event not found")
			})
		})
	}
}

func TestEventServiceInputValidation(t *testing.T) {
	ctx := context.Background()
	service, _ := newEventServiceTest(t)

	t.Run("can manage nil event", func(t *testing.T) {
		require.False(t, service.CanManage(nil, models.Profile{ID: 1, FirstName: "User"}))
	})

	t.Run("couple add nil dancer profile", func(t *testing.T) {
		require.NotPanics(t, func() {
			_, err := service.CoupleAdd(ctx, "event_id", nil, models.RoleLeader, "Partner")
			require.Error(t, err)
			require.Contains(t, err.Error(), "profile is nil")
		})
	})

	t.Run("couple add nil partner profile", func(t *testing.T) {
		dancer := models.Profile{ID: 1, FirstName: "Dancer"}
		var partner *models.Profile
		require.NotPanics(t, func() {
			_, err := service.CoupleAdd(ctx, "event_id", &dancer, models.RoleLeader, partner)
			require.Error(t, err)
			require.Contains(t, err.Error(), "partner profile is nil")
		})
	})

	t.Run("create collects validation errors", func(t *testing.T) {
		_, err := service.Create(ctx, strings.Repeat("a", 4), models.Profile{}, models.EventSettings{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "profile ID must be positive")
		require.Contains(t, err.Error(), "profile first name must be provided")
	})

	t.Run("create trims and rejects empty caption", func(t *testing.T) {
		_, err := service.Create(ctx, "   ", models.Profile{ID: 1, FirstName: "Owner"}, models.EventSettings{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "event text must be provided")
	})

	t.Run("couple add rejects empty manual partner name", func(t *testing.T) {
		event, err := service.Create(ctx, "abc", models.Profile{ID: 1, FirstName: "Owner"}, models.EventSettings{})
		require.NoError(t, err)

		_, err = service.CoupleAdd(ctx, event.ID,
			&models.Profile{ID: 2, FirstName: "Dancer"},
			models.RoleLeader,
			"   ",
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "full name must be between")
	})
}

func newEventServiceTest(t *testing.T) (*EventService, *storepkg.SQLiteStore) {
	t.Helper()

	cfg := config.Default()
	cfg.Settings.RendererRepeats = []time.Duration{}
	cfg.Settings.ReRenderOnStartup = 0
	cfg.Settings.DraftCleanupEvery = 0
	cfg.Settings.EventTextMaxLen = 3

	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)

	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)

	render := NewRenderService(cfg.Settings, st, func(*models.Event, string) error {
		return nil
	})
	notifier := NewNotifierService(cfg.Settings, st, func(*models.Notification) error {
		return nil
	})
	return NewEventService(cfg.Settings, st, render, notifier), st
}
