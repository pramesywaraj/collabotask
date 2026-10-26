package column_test

import (
	"context"
	"testing"

	"collabotask/internal/mocks"
	"collabotask/internal/usecase/column"

	"github.com/stretchr/testify/mock"
)

type columnTestDeps struct {
	checker      *mocks.MockBoardAccessChecker
	columnRepo   *mocks.MockColumnRepository
	activityRepo *mocks.MockActivityRepository
	broadcaster  *mocks.MockBroadcaster
	tx           *mocks.MockTransactor
	uc           *column.ColumnUseCase
}

func newDeps(t *testing.T) columnTestDeps {
	t.Helper()
	d := columnTestDeps{
		checker:      mocks.NewMockBoardAccessChecker(t),
		columnRepo:   mocks.NewMockColumnRepository(t),
		activityRepo: mocks.NewMockActivityRepository(t),
		broadcaster:  mocks.NewMockBroadcaster(t),
		tx:           mocks.NewMockTransactor(t),
	}
	d.uc = column.NewColumnUseCase(d.columnRepo, d.checker, d.activityRepo, d.broadcaster, d.tx)
	return d
}

// passthroughTx registers a permissive passthrough on the Transactor mock.
func passthroughTx(d columnTestDeps) {
	d.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Maybe()
}
