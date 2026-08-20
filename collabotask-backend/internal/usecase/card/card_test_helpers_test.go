package card_test

import (
	"testing"

	"collabotask/internal/mocks"
	"collabotask/internal/usecase/card"
)

type cardTestDeps struct {
	checker         *mocks.MockBoardAccessChecker
	cardRepo        *mocks.MockCardRepository
	columnRepo      *mocks.MockColumnRepository
	userRepo        *mocks.MockUserRepository
	boardMemberRepo *mocks.MockBoardMemberRepository
	activityRepo    *mocks.MockActivityRepository
	broadcaster     *mocks.MockBroadcaster
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
	}
	d.uc = card.NewCardUseCase(d.cardRepo, d.columnRepo, d.userRepo, d.checker, d.boardMemberRepo, d.activityRepo, d.broadcaster)
	return d
}

// ptr returns a pointer to v — handy for the optional/patch fields on card inputs.
func ptr[T any](v T) *T { return &v }
