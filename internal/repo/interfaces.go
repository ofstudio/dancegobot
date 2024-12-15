package repo

import (
	"context"

	"github.com/ofstudio/dancegobot/internal/models"
)

type Store interface {
	Close()
	Commit() error
	Rollback() error
	BeginTx(ctx context.Context) (Store, error)
	EventGet(ctx context.Context, eventID string) (*models.Event, error)
	EventUpsert(ctx context.Context, event *models.Event) error
	UserGet(ctx context.Context, id int64) (*models.User, error)
	UserUpsert(ctx context.Context, user *models.User) error
	HistoryInsert(ctx context.Context, item *models.HistoryItem) error
}
