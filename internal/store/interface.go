package store

import (
	"context"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

type Store interface {
	Close()
	Begin(ctx context.Context) (Store, error)
	Commit() error
	Rollback() error
	EventGet(ctx context.Context, eventID string) (*models.Event, error)
	EventUpsert(ctx context.Context, event *models.Event) error
	EventGetUpdatedAfter(ctx context.Context, after time.Time) ([]*models.Event, error)
	EventRemoveDraftsBefore(ctx context.Context, before time.Time) ([]string, error)
	EventGetMy(ctx context.Context, profile *models.Profile) ([]string, error)
	EventSubscribersNotifiedSet(ctx context.Context, eventID string) (bool, error)
	UserGet(ctx context.Context, id int64) (*models.User, error)
	UserUpsertProfile(ctx context.Context, user *models.User) error
	UserUpdateSession(ctx context.Context, user *models.User) error
	UserUpdateSettings(ctx context.Context, user *models.User) error
	HistoryCreate(ctx context.Context, item *models.HistoryItem) error
	HistoryRemoveByEventIDs(ctx context.Context, eventIDs []string) (int, error)
	SubscriptionGet(ctx context.Context, key models.SubscriptionKey) (*models.Subscription, error)
	SubscriptionCreate(ctx context.Context, subscription *models.Subscription) (bool, error)
	SubscriptionRemove(ctx context.Context, key models.SubscriptionKey) (*models.Subscription, error)
	SubscriptionGetByChatID(ctx context.Context, chatID int64) ([]*models.Subscription, error)
}

var _ Store = (*SQLiteStore)(nil)
