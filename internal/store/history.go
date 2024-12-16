package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (s *SQLiteStore) HistoryInsert(ctx context.Context, item *models.HistoryItem) error {
	const query =
	// language=SQLite
	`INSERT INTO history (action, initiator_id, event_id, data)
VALUES (?1, ?2, ?3, $4);`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshal, err)
	}

	var initiatorID *int64
	if item.Initiator != nil {
		initiatorID = &item.Initiator.ID
	}

	if _, err = stmt.ExecContext(ctx, item.Action, initiatorID, item.EventID, data); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return nil
}

// HistoryRemoveByEventIDs removes history items by event IDs.
// Returns the number of removed items.
func (s *SQLiteStore) HistoryRemoveByEventIDs(ctx context.Context, eventIDs []string) (int, error) {
	if len(eventIDs) == 0 {
		return 0, nil
	}
	// language=SQLite
	query, args, err := sqlx.In(
		`DELETE FROM history WHERE event_id IN (?)`,
		eventIDs,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to bind query: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(affected), nil
}
