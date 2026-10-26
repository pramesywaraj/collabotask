package card_test

import (
	"context"
	"testing"

	"collabotask/internal/mocks"
	"collabotask/internal/usecase/card"

	"github.com/stretchr/testify/mock"
)

type cardTestDeps struct {
	checker         *mocks.MockBoardAccessChecker
	cardRepo        *mocks.MockCardRepository
	columnRepo      *mocks.MockColumnRepository
	userRepo        *mocks.MockUserRepository
	boardMemberRepo *mocks.MockBoardMemberRepository
	activityRepo    *mocks.MockActivityRepository
	broadcaster     *mocks.MockBroadcaster
	tx              *mocks.MockTransactor
	uc              *card.CardUseCase
}

func newDeps(t *testing.T) cardTestDeps {
	t.Helper()
	d := cardTestDeps{
		checker:         mocks.NewMockBoardAccessChecker(t),
		cardRepo:        mocks.NewMockCardRepository(t),
		columnRepo:      mocks.NewMockColumnRepository(t),
		userRepo:        mocks.NewMockUserRepository(t),
		boardMemberRepo: mocks.NewMockBoardMemberRepository(t),
		activityRepo:    mocks.NewMockActivityRepository(t),
		broadcaster:     mocks.NewMockBroadcaster(t),
		tx:              mocks.NewMockTransactor(t),
	}
	d.uc = card.NewCardUseCase(d.cardRepo, d.columnRepo, d.userRepo, d.checker, d.boardMemberRepo, d.activityRepo, d.broadcaster, d.tx)
	return d
}

// passthroughTx registers a permissive passthrough on the Transactor mock:
// fn(ctx) is called and its result is returned. Use Maybe() so tests that
// fail before the tx is opened don't require it.
func passthroughTx(d cardTestDeps) {
	d.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Maybe()
}

// ptr returns a pointer to v — handy for the optional/patch fields on card inputs.
func ptr[T any](v T) *T { return &v }
