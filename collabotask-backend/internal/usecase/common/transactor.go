package common

import "context"

// Transactor is the usecase-layer port for atomic mutation+Log writes (ADR-016).
// WithinTransaction begins a new DB transaction, stashes it in ctx, runs fn, and
// commits on success or rolls back on error. Repos resolve the ambient tx via
// exec(ctx) — see internal/repository/postgres/base.go.
//
// A Log failure inside fn now fails the mutation (rolls back); this is the
// intentional behavior reversal from ADR-007 Option A → ADR-016 Option C. Do not
// add swallow-on-error logic here: the same-DB reasoning in ADR-016 Q1 defends it.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
