package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

type subscriptionRow struct {
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}

// SubscriptionGet returns a subscription by its key.
// If the subscription does not exist, returns nil.
func (s *SQLiteStore) SubscriptionGet(
	ctx context.Context,
	key models.SubscriptionKey,
) (*models.Subscription, error) {
	const query = `SELECT data, created_at
FROM subscriptions
WHERE subscriber_id = ?1 AND chat_id = ?2`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	row := &subscriptionRow{}
	if err = stmt.QueryRowxContext(ctx, key.SubscriberID, key.ChatID).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	return s.subscriptionUnmarshal(row)
}

// SubscriptionCreate creates a subscription.
// Returns false if a subscription with the same key already exists.
func (s *SQLiteStore) SubscriptionCreate(
	ctx context.Context,
	subscription *models.Subscription,
) (bool, error) {
	if subscription == nil {
		return false, ErrNil
	}
	const query =
	// language=SQLite
	`INSERT INTO subscriptions (subscriber_id, chat_id, data)
VALUES (?1, ?2, ?3)
ON CONFLICT (subscriber_id, chat_id) DO NOTHING
RETURNING data, created_at;`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := s.marshal("data", subscription)
	if err != nil {
		return false, err
	}
	row := &subscriptionRow{}
	if err = stmt.QueryRowxContext(ctx, subscription.Subscriber.ID, subscription.Chat.ID, data).
		StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	created, err := s.subscriptionUnmarshal(row)
	if err != nil {
		return false, err
	}
	*subscription = *created
	return true, nil
}

// SubscriptionRemove removes a subscription by its key and returns it.
// If the subscription does not exist, returns nil.
func (s *SQLiteStore) SubscriptionRemove(
	ctx context.Context,
	key models.SubscriptionKey,
) (*models.Subscription, error) {
	const query = `DELETE FROM subscriptions
WHERE subscriber_id = ?1 AND chat_id = ?2
RETURNING data, created_at;`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	row := &subscriptionRow{}
	if err = stmt.QueryRowxContext(ctx, key.SubscriberID, key.ChatID).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	return s.subscriptionUnmarshal(row)
}

// SubscriptionGetByChatID returns all subscriptions for the chat.
func (s *SQLiteStore) SubscriptionGetByChatID(
	ctx context.Context,
	chatID int64,
) ([]*models.Subscription, error) {
	const query = `SELECT data, created_at
FROM subscriptions
WHERE chat_id = ?1
ORDER BY created_at, subscriber_id;`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	rows, err := stmt.QueryxContext(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	defer rows.Close()

	var subscriptions []*models.Subscription
	for rows.Next() {
		row := &subscriptionRow{}
		if err = rows.StructScan(row); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrScan, err)
		}
		subscription, unmarshalErr := s.subscriptionUnmarshal(row)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		subscriptions = append(subscriptions, subscription)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrScan, err)
	}
	return subscriptions, nil
}

func (s *SQLiteStore) subscriptionUnmarshal(row *subscriptionRow) (*models.Subscription, error) {
	if row == nil {
		return nil, ErrNil
	}
	subscription := &models.Subscription{}
	if err := s.unmarshal("data", row.Data, subscription); err != nil {
		return nil, err
	}
	subscription.CreatedAt = row.CreatedAt
	return subscription, nil
}
