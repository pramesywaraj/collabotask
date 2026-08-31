package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// base is embedded by every postgres repo. It exposes exec(ctx) and tx(ctx, fn)
// as the ONLY paths to the database — there is no public db field, so
// r.db.QueryRow(...) will not compile and the "bypasses ambient tx" footgun is
// closed at compile time (ADR-016, Decision #4).
//
// The residual base.pool surface is guarded by the forbidigo CI rule
// (\.pool\.(Query|QueryRow|Exec|Begin) forbidden outside base.go).
type base struct {
	pool *pgxpool.Pool
}

// exec returns the ambient pgx.Tx from ctx if one exists (placed there by
// Transactor.WithinTransaction or by tx below), otherwise returns the pool.
// All repo SELECT/INSERT/UPDATE/DELETE statements call r.exec(ctx).QueryRow/Query/Exec.
func (b base) exec(ctx context.Context) DBTX {
	if tx, ok := txFromCtx(ctx); ok {
		return tx
	}
	return b.pool
}

// tx implements "join-or-begin": if an ambient tx is already in ctx it runs fn
// directly on it (joining the outer transaction); otherwise it opens a new
// transaction on the pool, stores it in ctx, runs fn, and commits or rolls back.
//
// This is the internal counterpart to the Transactor port. Repos use it for the
// 9 methods that already owned their own transaction (card.Move, column.UpdatePosition,
// boardMember.TransferOwnership, etc.) so those methods remain atomic both when
// called standalone and when composed inside a Transactor.WithinTransaction closure.
func (b base) tx(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := txFromCtx(ctx); ok {
		return fn(ctx)
	}

	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctxWithTx(ctx, tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
