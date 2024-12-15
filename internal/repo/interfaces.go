package repo

import (
	"context"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

type Store interface {
	Close()
	Commit() error
	Rollback() error
	BeginTx(ctx context.Context) (Store, error)
	EventGet(ctx context.Context, eventID string) (*models.Event, error)
	EventUpsert(ctx context.Context, event *models.Event) error
	EventGetUpdatedAfter(ctx context.Context, after time.Time) ([]*models.Event, error)
	EventRemoveDraftsBefore(ctx context.Context, before time.Time) ([]string, error)
	UserGet(ctx context.Context, id int64) (*models.User, error)
	UserUpsert(ctx context.Context, user *models.User) error
	HistoryInsert(ctx context.Context, item *models.HistoryItem) error
	HistoryRemoveByEventIDs(ctx context.Context, eventIDs []string) (int, error)
}
