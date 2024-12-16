package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

// EventGet returns an event by its id.
// If the event does not exist, returns ErrNotFound.
func (s *SQLiteStore) EventGet(ctx context.Context, eventID string) (*models.Event, error) {
	const query = `SELECT data FROM events WHERE id = ?1`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}
	var data []byte
	if err = stmt.QueryRowxContext(ctx, eventID).Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	event := &models.Event{}
	if err = json.Unmarshal(data, event); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
	}

	return event, nil
}

// EventUpsert inserts or updates an event.
func (s *SQLiteStore) EventUpsert(ctx context.Context, event *models.Event) error {
	const query =
	// language=SQLite
	`INSERT INTO events (id, owner_id, data)
VALUES (?1, ?2, ?3)
ON CONFLICT (id) DO UPDATE SET owner_id   = excluded.owner_id,
                               data       = excluded.data,
                               updated_at = CURRENT_TIMESTAMP;`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshal, err)
	}

	if _, err = stmt.ExecContext(ctx, event.ID, event.Owner.ID, data); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return nil
}

// EventGetUpdatedAfter returns all non-draft events updated after the specified time.
func (s *SQLiteStore) EventGetUpdatedAfter(ctx context.Context, after time.Time) ([]*models.Event, error) {
	// language=SQLite
	const query = `SELECT data
FROM events
WHERE updated_at > ?1
  AND json_extract(data, '$.post.inline_message_id') IS NOT NULL`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	rows, err := stmt.QueryxContext(ctx, after)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	//goland:noinspection ALL
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
		}

		event := &models.Event{}
		if err = json.Unmarshal(data, event); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
		}

		events = append(events, event)
	}

	return events, nil
}

// EventRemoveDraftsBefore removes all draft events updated before the specified time.
// Returns the ids of the removed events.
func (s *SQLiteStore) EventRemoveDraftsBefore(ctx context.Context, before time.Time) ([]string, error) {
	// language=SQLite
	const query = `DELETE
FROM events
WHERE updated_at < ?1
  AND json_extract(data, '$.post.inline_message_id') IS NULL
  AND ifnull(json_array_length(data, '$.couples'), 0) = 0
  AND ifnull(json_array_length(data, '$.singles'), 0) = 0
RETURNING id;`

	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	rows, err := stmt.QueryxContext(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	//goland:noinspection ALL
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}
