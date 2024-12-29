package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

type userRow struct {
	ID        int64     `db:"id"`
	Profile   []byte    `db:"profile"`
	Session   []byte    `db:"session"`
	Settings  []byte    `db:"settings"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserGet returns user by its id.
// If the user does not exist, returns nil.
func (s *SQLiteStore) UserGet(ctx context.Context, id int64) (*models.User, error) {
	const query = `SELECT profile, session, settings, created_at, updated_at
FROM users
WHERE id = ?1
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	row := &userRow{}
	if err = stmt.QueryRowxContext(ctx, id).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	user := &models.User{}
	if err = s.userUnmarshal(row, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UserUpsertProfile creates new user or updates profile of existing one.
func (s *SQLiteStore) UserUpsertProfile(ctx context.Context, user *models.User) error {
	if user == nil {
		return ErrNil
	}
	const query =
	// language=SQLite
	`INSERT INTO users (id, profile, session, settings)
VALUES (:id, :profile, :session, :settings)
ON CONFLICT (id) DO UPDATE SET profile    = excluded.profile,
                               updated_at = CURRENT_TIMESTAMP
RETURNING id, profile, session, settings, created_at, updated_at
`
	stmt, err := s.namedStmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	row, err := s.userMarshal(user)
	if err != nil {
		return err
	}

	if err = stmt.QueryRowxContext(ctx, row).StructScan(row); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return s.userUnmarshal(row, user)
}

// UserUpdateSession updates user session.
// If the user does not exist, returns ErrNotFound.
func (s *SQLiteStore) UserUpdateSession(ctx context.Context, user *models.User) error {
	if user == nil {
		return ErrNil
	}
	const query =
	// language=SQLite
	`UPDATE users
SET session    = ?2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?1
RETURNING id, profile, session, settings, created_at, updated_at
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := s.marshal("session", user.Session)
	if err != nil {
		return err
	}

	row := &userRow{}
	if err = stmt.QueryRowxContext(ctx, user.Profile.ID, data).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return s.userUnmarshal(row, user)
}

// UserUpdateSettings updates user settings.
// If the user does not exist, returns ErrNotFound.
func (s *SQLiteStore) UserUpdateSettings(ctx context.Context, user *models.User) error {
	if user == nil {
		return ErrNil
	}
	const query =
	// language=SQLite
	`UPDATE users
SET settings   = ?2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?1
RETURNING id, profile, session, settings, created_at, updated_at
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := s.marshal("settings", user.Settings)
	if err != nil {
		return err
	}

	row := &userRow{}
	if err = stmt.QueryRowxContext(ctx, user.Profile.ID, data).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return s.userUnmarshal(row, user)
}

func (s *SQLiteStore) userMarshal(user *models.User) (*userRow, error) {
	if user == nil {
		return nil, ErrNil
	}
	var row userRow
	var err error
	row.ID = user.Profile.ID
	if row.Profile, err = s.marshal("profile", user.Profile); err != nil {
		return nil, err
	}
	if row.Session, err = s.marshal("session", user.Session); err != nil {
		return nil, err
	}
	if row.Settings, err = s.marshal("settings", user.Settings); err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *SQLiteStore) userUnmarshal(row *userRow, user *models.User) error {
	if user == nil || row == nil {
		return ErrNil
	}
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	if err := s.unmarshal("profile", row.Profile, &user.Profile); err != nil {
		return err
	}
	if err := s.unmarshal("session", row.Session, &user.Session); err != nil {
		return err
	}
	if err := s.unmarshal("settings", row.Settings, &user.Settings); err != nil {
		return err
	}
	return nil
}
