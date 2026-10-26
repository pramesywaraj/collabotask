package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the minimal query surface shared by *pgxpool.Pool and pgx.Tx.
// Every repo statement goes through exec(ctx) which returns one of these two;
// the ambient tx (if any) is stored in ctx under txKey.
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// txKey is the context key that carries the ambient pgx.Tx. It is unexported and
// package-local so that base.exec and the Transactor impl share it without leaking
// the key to callers.
type txKey struct{}

func txFromCtx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func ctxWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
