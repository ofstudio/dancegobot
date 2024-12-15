package telegram

import (
	"context"

	"github.com/ofstudio/dancegobot/internal/models"
)

type UserService interface {
	Get(ctx context.Context, profile models.Profile) (*models.User, error)
	Upsert(ctx context.Context, user *models.User) error
}

type EventService interface {
	Create(ctx context.Context, caption string, owner models.Profile, settings models.EventSettings) (*models.Event, error)
	Get(ctx context.Context, id string) (*models.Event, error)
	PostAdd(ctx context.Context, eventID string, inlineMessageID string) (*models.Event, *models.Post, error)
	PostChatAdd(ctx context.Context, eventID string, chat *models.Chat, chatMessageID int) (*models.Event, *models.Post, error)
	RegistrationGet(event *models.Event, profile *models.Profile, role models.Role) *models.Registration
	CoupleAdd(ctx context.Context, eventID string, profile *models.Profile, role models.Role, other any) (*models.Registration, error)
	SingleAdd(ctx context.Context, eventID string, profile *models.Profile, role models.Role) (*models.Registration, error)
	DancerRemove(ctx context.Context, eventID string, profile *models.Profile) (*models.Registration, error)
}
