package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
)

// SQLiteStore is bot data store.
type SQLiteStore struct {
	db            *sqlx.DB                   // Database connection
	execer        execer                     // Database query interface
	stmtsIdx      map[string]*sqlx.Stmt      // Prepared statements cache
	namedStmtsIdx map[string]*sqlx.NamedStmt // Prepared named statements cache
	mu            sync.Mutex                 // Mutex for prepared statements caches
}

// NewSQLiteStore creates a new SQLite data store.
func NewSQLiteStore(db *sqlx.DB) *SQLiteStore {
	return &SQLiteStore{
		db:            db,
		execer:        db,
		stmtsIdx:      make(map[string]*sqlx.Stmt),
		namedStmtsIdx: make(map[string]*sqlx.NamedStmt),
	}
}

// DB returns a connection to the database.
func (s *SQLiteStore) DB() *sqlx.DB {
	return s.db
}

// Close closes the storage.
func (s *SQLiteStore) Close() {
	s.closeAllStmts()
	_ = s.db.Close()
}

// Commit commits the transaction.
// If the store is not within a transaction, an error is returned.
func (s *SQLiteStore) Commit() error {
	tx, ok := s.execer.(txer)
	if !ok {
		return fmt.Errorf("unable to commit non-existent transaction")
	}
	s.closeAllStmts()
	return tx.Commit()
}

// Rollback aborts the transaction.
// If the store is not within a transaction, an error is returned.
func (s *SQLiteStore) Rollback() error {
	tx, ok := s.execer.(txer)
	if !ok {
		return fmt.Errorf("unable to rollback non-existent transaction")
	}
	s.closeAllStmts()
	return tx.Rollback()
}

// Begin returns a new [SQLiteStore] within a transaction.
func (s *SQLiteStore) Begin(ctx context.Context) (Store, error) {
	if s.execer != s.db {
		return nil, fmt.Errorf("unable to start a transaction within another transaction")
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{
		db:            s.db,
		execer:        &txExt{Tx: tx},
		stmtsIdx:      make(map[string]*sqlx.Stmt),
		namedStmtsIdx: make(map[string]*sqlx.NamedStmt),
	}, nil
}

// stmt returns a prepared statement.
// If the statement was prepared previously, it will be returned from the cache.
func (s *SQLiteStore) stmt(ctx context.Context, query string) (*sqlx.Stmt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stmt, ok := s.stmtsIdx[query]
	if !ok {
		var err error
		stmt, err = s.execer.PreparexContext(ctx, query)
		if err != nil {
			return nil, err
		}
		s.stmtsIdx[query] = stmt
	}
	return stmt, nil
}

// namedStmt returns a prepared named statement.
// If the statement was prepared previously, it will be returned from the cache.
func (s *SQLiteStore) namedStmt(ctx context.Context, query string) (*sqlx.NamedStmt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stmt, ok := s.namedStmtsIdx[query]
	if !ok {
		var err error
		stmt, err = s.execer.PrepareNamedContext(ctx, query)
		if err != nil {
			return nil, err
		}
		s.namedStmtsIdx[query] = stmt
	}
	return stmt, nil
}

// closeAllStmts closes all prepared statements.
func (s *SQLiteStore) closeAllStmts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for query, stmt := range s.stmtsIdx {
		_ = stmt.Close()
		delete(s.stmtsIdx, query)
	}
	for query, stmt := range s.namedStmtsIdx {
		_ = stmt.Close()
		delete(s.namedStmtsIdx, query)
	}
}

func (s *SQLiteStore) marshal(filedName string, field any) ([]byte, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrMarshal, filedName, err)
	}
	return data, nil
}

func (s *SQLiteStore) unmarshal(filedName string, data []byte, field any) error {
	if err := json.Unmarshal(data, field); err != nil {
		return fmt.Errorf("%w %s: %w", ErrUnmarshal, filedName, err)
	}
	return nil
}

// execer - interface for executing queries to the database.
type execer interface {
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

// txer - interface for executing queries to the database within a transaction.
type txer interface {
	execer
	Commit() error
	Rollback() error
}

// txExt - wrapper for [sqlx.Tx], extended with the NamedQueryContext method.
// See https://github.com/jmoiron/sqlx/issues/447
type txExt struct {
	*sqlx.Tx
}

// NamedQueryContext is [sqlx.NamedQueryContext] method for sqlx.Tx.
func (tx *txExt) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return sqlx.NamedQueryContext(ctx, tx, query, arg)
}
