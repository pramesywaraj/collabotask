package board_test

import (
	"testing"

	"collabotask/internal/mocks"
	"collabotask/internal/usecase/board"
)

type boardTestDeps struct {
	checker      *mocks.MockBoardAccessChecker
	boardRepo    *mocks.MockBoardRepository
	boardMbrRepo *mocks.MockBoardMemberRepository
	wsMbrRepo    *mocks.MockWorkspaceMemberRepository
	userRepo     *mocks.MockUserRepository
	columnRepo   *mocks.MockColumnRepository
	cardRepo     *mocks.MockCardRepository
	activityRepo *mocks.MockActivityRepository
	broadcaster  *mocks.MockBroadcaster
	uc           *board.BoardUseCase
}

func newDeps(t *testing.T) boardTestDeps {
	t.Helper()
	d := boardTestDeps{
		checker:      mocks.NewMockBoardAccessChecker(t),
		boardRepo:    mocks.NewMockBoardRepository(t),
		boardMbrRepo: mocks.NewMockBoardMemberRepository(t),
		wsMbrRepo:    mocks.NewMockWorkspaceMemberRepository(t),
		userRepo:     mocks.NewMockUserRepository(t),
		columnRepo:   mocks.NewMockColumnRepository(t),
		cardRepo:     mocks.NewMockCardRepository(t),
		activityRepo: mocks.NewMockActivityRepository(t),
		broadcaster:  mocks.NewMockBroadcaster(t),
	}
	d.uc = board.NewBoardUseCase(
		d.checker,
		d.boardRepo,
		d.boardMbrRepo,
		d.wsMbrRepo,
		d.userRepo,
		d.columnRepo,
		d.cardRepo,
		d.activityRepo,
		d.broadcaster,
	)
	return d
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
