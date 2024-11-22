package telegram

import (
	"context"

	"github.com/ofstudio/dancegobot/internal/models"
)

type UserService interface {
	Get(ctx context.Context, userID int64) (*models.User, error)
	ProfileUpsert(ctx context.Context, user *models.User) error
	SessionUpsert(ctx context.Context, user *models.User) error
}

type EventService interface {
	NewID() string
	Create(ctx context.Context, event *models.Event) error
	Get(ctx context.Context, id string) (*models.Event, error)
	DancerGet(event *models.Event, profile *models.Profile, role models.Role) *models.Dancer
	CoupleAdd(ctx context.Context, eventID string, profile *models.Profile, role models.Role, other any) (*models.EventUpdate, error)
	SingleAdd(ctx context.Context, eventID string, profile *models.Profile, role models.Role) (*models.EventUpdate, error)
	DancerRemove(ctx context.Context, eventID string, profile *models.Profile) (*models.EventUpdate, error)
}
