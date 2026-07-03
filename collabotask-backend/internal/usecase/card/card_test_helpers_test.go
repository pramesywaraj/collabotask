package card_test

import (
	"testing"

	"collabotask/internal/mocks"
	"collabotask/internal/usecase/card"
)

type cardTestDeps struct {
	checker    *mocks.MockBoardAccessChecker
	cardRepo   *mocks.MockCardRepository
	columnRepo *mocks.MockColumnRepository
	userRepo   *mocks.MockUserRepository
	uc         *card.CardUseCase
}

func newDeps(t *testing.T) cardTestDeps {
	t.Helper()
	d := cardTestDeps{
		checker:    mocks.NewMockBoardAccessChecker(t),
		cardRepo:   mocks.NewMockCardRepository(t),
		columnRepo: mocks.NewMockColumnRepository(t),
		userRepo:   mocks.NewMockUserRepository(t),
	}
	d.uc = card.NewCardUseCase(d.cardRepo, d.columnRepo, d.userRepo, d.checker)
	return d
}

// ptr returns a pointer to v — handy for the optional/patch fields on card inputs.
func ptr[T any](v T) *T { return &v }
