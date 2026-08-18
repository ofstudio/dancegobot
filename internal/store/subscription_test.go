package store

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (suite *TestStoreSuite) TestSubscriptionCRUD() {
	suite.Run("create get list and remove", func() {
		ctx := context.Background()
		first := &models.Subscription{
			Subscriber: models.Profile{ID: 10, FirstName: "First"},
			Chat:       models.Chat{ID: -1001, Type: models.ChatSuper, Title: "Dance"},
		}

		created, err := suite.store.SubscriptionCreate(ctx, first)
		suite.Require().NoError(err)
		suite.True(created)
		suite.False(first.CreatedAt.IsZero())

		got, err := suite.store.SubscriptionGet(ctx, first.Key())
		suite.Require().NoError(err)
		suite.Equal(first, got)

		duplicate := &models.Subscription{
			Subscriber: models.Profile{ID: 10, FirstName: "Changed"},
			Chat:       models.Chat{ID: -1001, Type: models.ChatSuper, Title: "Changed"},
		}
		created, err = suite.store.SubscriptionCreate(ctx, duplicate)
		suite.Require().NoError(err)
		suite.False(created)
		got, err = suite.store.SubscriptionGet(ctx, first.Key())
		suite.Require().NoError(err)
		suite.Equal(first, got)

		second := &models.Subscription{
			Subscriber: models.Profile{ID: 11, FirstName: "Second"},
			Chat:       first.Chat,
		}
		created, err = suite.store.SubscriptionCreate(ctx, second)
		suite.Require().NoError(err)
		suite.True(created)
		otherChat := &models.Subscription{
			Subscriber: models.Profile{ID: 12, FirstName: "Other"},
			Chat:       models.Chat{ID: -1002, Type: models.ChatChannel},
		}
		created, err = suite.store.SubscriptionCreate(ctx, otherChat)
		suite.Require().NoError(err)
		suite.True(created)

		subscriptions, err := suite.store.SubscriptionGetByChatID(ctx, first.Chat.ID)
		suite.Require().NoError(err)
		suite.Require().Len(subscriptions, 2)
		suite.Equal(int64(10), subscriptions[0].Subscriber.ID)
		suite.Equal(int64(11), subscriptions[1].Subscriber.ID)

		removed, err := suite.store.SubscriptionRemove(ctx, first.Key())
		suite.Require().NoError(err)
		suite.Equal(first, removed)
		removed, err = suite.store.SubscriptionRemove(ctx, first.Key())
		suite.Require().NoError(err)
		suite.Nil(removed)
		got, err = suite.store.SubscriptionGet(ctx, first.Key())
		suite.Require().NoError(err)
		suite.Nil(got)
	})
}

func (suite *TestStoreSuite) TestSubscriptionTransactionRollback() {
	suite.Run("rollback", func() {
		ctx := context.Background()
		tx, err := suite.store.Begin(ctx)
		suite.Require().NoError(err)
		subscription := &models.Subscription{
			Subscriber: models.Profile{ID: 10, FirstName: "First"},
			Chat:       models.Chat{ID: -1001, Type: models.ChatSuper},
		}
		created, err := tx.SubscriptionCreate(ctx, subscription)
		suite.Require().NoError(err)
		suite.True(created)
		suite.Require().NoError(tx.Rollback())

		got, err := suite.store.SubscriptionGet(ctx, subscription.Key())
		suite.Require().NoError(err)
		suite.Nil(got)
	})
}

func (suite *TestStoreSuite) TestSubscriptionGetByChatIDUsesIndex() {
	suite.Run("query plan", func() {
		rows, err := suite.store.DB().Query(`EXPLAIN QUERY PLAN
		SELECT data, created_at
		FROM subscriptions
		WHERE chat_id = ?1
		ORDER BY created_at, subscriber_id`, -1001)
		suite.Require().NoError(err)
		defer rows.Close()

		var plan []string
		for rows.Next() {
			var id, parent, unused int
			var detail string
			suite.Require().NoError(rows.Scan(&id, &parent, &unused, &detail))
			plan = append(plan, detail)
		}
		suite.Require().NoError(rows.Err())
		suite.Contains(strings.Join(plan, "\n"), "USING INDEX subscriptions_chat_id_created_at")
	})
}

func TestSubscriptionsMigrationBackfillsPublishedEvents(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/migration.db"
	db, err := NewSQLite(dbPath, 2)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events (id, owner_id, data) VALUES
('published', 1, '{"id":"published","post":{"chat_message_id":10}}'),
('draft', 1, '{"id":"draft"}')`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	db, err = NewSQLite(dbPath, 3)
	require.NoError(t, err)
	store := NewSQLiteStore(db)
	t.Cleanup(store.Close)

	published, err := store.EventGet(ctx, "published")
	require.NoError(t, err)
	require.True(t, published.SubscribersNotified)
	draft, err := store.EventGet(ctx, "draft")
	require.NoError(t, err)
	require.False(t, draft.SubscribersNotified)

	created, err := store.SubscriptionCreate(ctx, &models.Subscription{
		Subscriber: models.Profile{ID: 10, FirstName: "First"},
		Chat:       models.Chat{ID: -1001, Type: models.ChatSuper},
	})
	require.NoError(t, err)
	require.True(t, created)
}
