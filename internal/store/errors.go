package store

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrStmtPrepare = errors.New("failed to prepare statement")
	ErrStmtExec    = errors.New("failed to execute statement")
	ErrScan        = errors.New("failed to scan data")
	ErrMarshal     = errors.New("failed to marshal data")
	ErrUnmarshal   = errors.New("failed to unmarshal data")
	ErrNilProfile  = errors.New("profile is nil")
)
