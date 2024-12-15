package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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
	// language=SQLite
	const query = `
INSERT INTO events (id, owner_id, data)
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
