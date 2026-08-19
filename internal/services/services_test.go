package services

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	storepkg "github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
)

func TestNewServices(t *testing.T) {
	cfg := config.Default()
	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)

	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)

	container := NewServices(
		cfg.Settings,
		st,
		func(*models.Event, string) error { return nil },
		func(*models.Notification) error { return nil },
		func(int64, int64) (models.Membership, error) { return models.Membership{}, nil },
	)

	require.NotNil(t, container.Event)
	require.NotNil(t, container.User)
	require.NotNil(t, container.Subscription)
	require.NotNil(t, container.Notifier)
	require.NotNil(t, container.Render)
	require.Same(t, container.Render, container.Event.renderer)
	require.Same(t, container.Notifier, container.Event.notifier)
	require.Same(t, container.Subscription, container.Event.eventPublishedHandler)
	require.Same(t, container.Notifier, container.Subscription.notifier)
	require.Same(t, st, container.Event.store)
	require.Same(t, st, container.User.store)
	require.Same(t, st, container.Subscription.store)
	require.Same(t, st, container.Notifier.store)
	require.Same(t, st, container.Render.store)
}

func TestServicesWithLogger(t *testing.T) {
	cfg := config.Default()
	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)

	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)

	container := NewServices(
		cfg.Settings,
		st,
		func(*models.Event, string) error { return nil },
		func(*models.Notification) error { return nil },
		func(int64, int64) (models.Membership, error) { return models.Membership{}, nil },
	)
	log := noplog.Logger()

	require.Same(t, container, container.WithLogger(log))
	require.Same(t, log, container.Event.log)
	require.Same(t, log, container.User.log)
	require.Same(t, log, container.Subscription.log)
	require.Same(t, log, container.Notifier.log)
	require.Same(t, log, container.Render.log)
}
