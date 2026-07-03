package column_test

import (
	"testing"

	"collabotask/internal/mocks"
	"collabotask/internal/usecase/column"
)

type columnTestDeps struct {
	checker    *mocks.MockBoardAccessChecker
	columnRepo *mocks.MockColumnRepository
	uc         *column.ColumnUseCase
}

func newDeps(t *testing.T) columnTestDeps {
	t.Helper()
	d := columnTestDeps{
		checker:    mocks.NewMockBoardAccessChecker(t),
		columnRepo: mocks.NewMockColumnRepository(t),
	}
	d.uc = column.NewColumnUseCase(d.columnRepo, d.checker)
	return d
}
