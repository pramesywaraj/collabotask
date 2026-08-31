package postgres

import (
	"collabotask/internal/usecase/common"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type transactor struct {
	pool *pgxpool.Pool
}

// NewTransactor returns the postgres implementation of common.Transactor.
// It shares txKey with base so that r.exec(ctx) in every repo picks up the
// ambient transaction that WithinTransaction places in ctx.
func NewTransactor(pool *pgxpool.Pool) common.Transactor {
	return &transactor{pool: pool}
}

// WithinTransaction begins a new transaction on the pool, stores it in ctx under
// txKey, runs fn, and commits on success or rolls back on error or panic.
// Repos pick up the tx via base.exec(ctx) — no explicit tx parameter needed.
func (t *transactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctxWithTx(ctx, tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
