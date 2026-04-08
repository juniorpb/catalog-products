package database

import (
	"context"
	"database/sql"
)

// Executor is satisfied by both *sql.DB and *sql.Tx, allowing Insert functions
// to participate in a transaction without changing their call sites.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
