package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (s *Store) HistoryInsert(ctx context.Context, item *models.HistoryItem) error {
	// language=SQLite
	const query = `INSERT INTO history (action, initiator_id, event_id, data)
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
