package models

import (
	"log/slog"
	"time"
)

// Subscription represents a user's subscription to new event posts in a Telegram chat.
// A subscription is uniquely identified by the subscriber and chat IDs.
type Subscription struct {
	Subscriber Profile   `json:"subscriber"` // Telegram profile of the subscriber
	Chat       Chat      `json:"chat"`       // Chat whose new event posts the user is subscribed to
	CreatedAt  time.Time `json:"created_at"` // Creation time
}

// Key returns the identity of the subscription.
func (s Subscription) Key() SubscriptionKey {
	return SubscriptionKey{
		SubscriberID: s.Subscriber.ID,
		ChatID:       s.Chat.ID,
	}
}

// LogValue implements slog.Valuer interface for Subscription model.
func (s Subscription) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("subscriber", slog.GroupValue(
			slog.Int64("id", s.Subscriber.ID),
		)),
		slog.Any("chat", s.Chat.LogValue()),
	)
}

// SubscriptionKey uniquely identifies a subscription.
// The storage must enforce uniqueness for the subscriber and chat IDs.
type SubscriptionKey struct {
	SubscriberID int64 `json:"subscriber_id"` // Telegram user ID
	ChatID       int64 `json:"chat_id"`       // Telegram chat ID
}

// LogValue implements slog.Valuer interface for SubscriptionKey model.
func (k SubscriptionKey) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int64("subscriber_id", k.SubscriberID),
		slog.Int64("chat_id", k.ChatID),
	)
}
