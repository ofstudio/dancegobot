package store

import "errors"

var (
	ErrStmtPrepare = errors.New("failed to prepare statement")
	ErrStmtExec    = errors.New("failed to execute statement")
	ErrScan        = errors.New("failed to scan")
	ErrNil         = errors.New("nil value")
	ErrNotFound    = errors.New("not found")
	ErrMarshal     = errors.New("failed to marshal")
	ErrUnmarshal   = errors.New("failed to unmarshal")
)
