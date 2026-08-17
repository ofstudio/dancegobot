package models

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionKey(t *testing.T) {
	subscription := Subscription{
		Subscriber: Profile{ID: 42, FirstName: "John"},
		Chat:       Chat{ID: -100123, Type: ChatSuper, Title: "Dance"},
	}

	require.Equal(t, SubscriptionKey{
		SubscriberID: 42,
		ChatID:       -100123,
	}, subscription.Key())
}

func TestSubscriptionLogValue(t *testing.T) {
	subscription := Subscription{
		Subscriber: Profile{ID: 42, FirstName: "John"},
		Chat:       Chat{ID: -100123, Type: ChatSuper, Title: "Dance"},
	}

	value := subscription.LogValue()
	require.Equal(t, slog.KindGroup, value.Kind())
	require.Equal(t, []slog.Attr{
		slog.Any("subscriber", slog.GroupValue(slog.Int64("id", 42))),
		slog.Any("chat", subscription.Chat.LogValue()),
	}, value.Group())

	keyValue := subscription.Key().LogValue()
	require.Equal(t, slog.KindGroup, keyValue.Kind())
	require.Equal(t, []slog.Attr{
		slog.Int64("subscriber_id", 42),
		slog.Int64("chat_id", -100123),
	}, keyValue.Group())
}
