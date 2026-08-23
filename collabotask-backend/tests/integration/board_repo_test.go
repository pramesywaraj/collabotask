package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
	"collabotask/tests/integration/testutil"
)

var testPool *pgxpool.Pool
var fx *testutil.Fixtures

func TestMain(m *testing.M) {
	pool, terminate := testutil.NewTestDB()
	testPool = pool
	fx = testutil.NewFixtures(pool)
	code := m.Run()
	terminate()
	os.Exit(code)
}

func TestBoardRepository_CreateWithOwner(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)

	repo := pgRepos.NewBoardRepository(testPool)
	board := &entity.Board{
		WorkspaceID:     ws.ID,
		Title:           "Test Board",
		CreatedBy:       user.ID,
		BackgroundColor: "#0079BF",
		Visibility:      entity.BoardVisibilityWorkspace,
	}

	err := repo.CreateWithOwner(context.Background(), board, user.ID)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, board.ID)
	assert.Equal(t, "Test Board", board.Title)
	assert.Equal(t, ws.ID, board.WorkspaceID)

	// Verify the creator was added as BOARD_OWNER
	memberRepo := pgRepos.NewBoardMemberRepository(testPool)
	member, err := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, entity.BoardRoleOwner, member.Role)
}
