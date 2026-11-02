package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
)

// Fixtures holds the shared pool and exposes helpers for seeding test data.
// Create one per test package via NewFixtures(testPool).
type Fixtures struct {
	pool *pgxpool.Pool
}

func NewFixtures(pool *pgxpool.Pool) *Fixtures {
	return &Fixtures{pool: pool}
}

// CreateTestUser inserts a user with a unique email and returns the created entity.
func (f *Fixtures) CreateTestUser(t *testing.T) *entity.User {
	t.Helper()
	repo := pgRepos.NewUserRepository(f.pool)
	user := &entity.User{
		Email:        fmt.Sprintf("user-%s@test.com", uuid.New().String()[:8]),
		Name:         "Test User",
		PasswordHash: "$2a$10$testhashfortestingonly",
		SystemRole:   entity.SystemRoleUser,
	}
	require.NoError(t, repo.Create(context.Background(), user))
	return user
}

// CreateTestWorkspace inserts a workspace owned by ownerID and returns the created entity.
func (f *Fixtures) CreateTestWorkspace(t *testing.T, ownerID uuid.UUID) *entity.Workspace {
	t.Helper()
	repo := pgRepos.NewWorkspaceRepository(f.pool)
	ws := &entity.Workspace{
		Name:    fmt.Sprintf("Workspace-%s", uuid.New().String()[:8]),
		OwnerID: ownerID,
	}
	require.NoError(t, repo.Create(context.Background(), ws))
	return ws
}

// CreateTestBoard inserts a WORKSPACE-visible board (with ownerID as BOARD_OWNER) and
// returns it. The seeded defaults live in CreateTestBoardWithVisibility.
func (f *Fixtures) CreateTestBoard(t *testing.T, workspaceID, ownerID uuid.UUID) *entity.Board {
	t.Helper()
	return f.CreateTestBoardWithVisibility(t, workspaceID, ownerID, entity.BoardVisibilityWorkspace)
}

// CreateTestColumn inserts a column in the given board and returns it. It lands one
// PositionStep past the last default column — the same position create_column.go would
// append it at — so it never collides with a default.
func (f *Fixtures) CreateTestColumn(t *testing.T, boardID uuid.UUID) *entity.Column {
	t.Helper()
	repo := pgRepos.NewColumnRepository(f.pool)
	col := &entity.Column{
		BoardID:  boardID,
		Title:    fmt.Sprintf("Column-%s", uuid.New().String()[:8]),
		Position: float64(len(domain.DefaultNewBoardColumnTitles)+1) * domain.PositionStep,
	}
	require.NoError(t, repo.Create(context.Background(), col))
	return col
}

// CreateTestCard inserts a card at position 1000 in the given column/board and returns
// it. The seeded defaults live in CreateTestCardAt.
func (f *Fixtures) CreateTestCard(t *testing.T, columnID, boardID, createdBy uuid.UUID) *entity.Card {
	t.Helper()
	return f.CreateTestCardAt(
		t,
		columnID,
		boardID,
		createdBy,
		fmt.Sprintf("Card-%s", uuid.New().String()[:8]),
		1000,
	)
}

// AddBoardMember adds userID as a BOARD_MEMBER of boardID.
func (f *Fixtures) AddBoardMember(t *testing.T, boardID, userID uuid.UUID) {
	t.Helper()
	repo := pgRepos.NewBoardMemberRepository(f.pool)
	_, err := repo.CreateIfAbsent(context.Background(), &entity.BoardMember{
		BoardID: boardID,
		UserID:  userID,
		Role:    entity.BoardRoleMember,
	})
	require.NoError(t, err)
}

// CreateTestBoardWithVisibility inserts a board with an explicit visibility and returns
// it, and holds the seeded board defaults for CreateTestBoard too. PRIVATE is what
// break-glass joins and the visibility filter turn on.
func (f *Fixtures) CreateTestBoardWithVisibility(t *testing.T, workspaceID, ownerID uuid.UUID, visibility entity.BoardVisibility) *entity.Board {
	t.Helper()
	repo := pgRepos.NewBoardRepository(f.pool)
	board := &entity.Board{
		WorkspaceID:     workspaceID,
		Title:           fmt.Sprintf("Board-%s", uuid.New().String()[:8]),
		CreatedBy:       ownerID,
		BackgroundColor: "#0079BF",
		Visibility:      visibility,
	}
	require.NoError(t, repo.CreateWithOwner(context.Background(), board, ownerID))
	return board
}

// AddWorkspaceMember adds userID to the workspace with the given role. Workspace Admin
// authority outranks board authority, so the role decides what the user may see and join.
func (f *Fixtures) AddWorkspaceMember(t *testing.T, workspaceID, userID uuid.UUID, role entity.WorkspaceRole) {
	t.Helper()
	repo := pgRepos.NewWorkspaceMemberRepository(f.pool)
	require.NoError(t, repo.Create(context.Background(), &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	}))
}

// CreateTestCardAt inserts a card with an explicit title and position, and holds the
// seeded card defaults for CreateTestCard too (which always uses position 1000).
func (f *Fixtures) CreateTestCardAt(t *testing.T, columnID, boardID, createdBy uuid.UUID, title string, position float64) *entity.Card {
	t.Helper()
	repo := pgRepos.NewCardRepository(f.pool)
	card := &entity.Card{
		ColumnID:  columnID,
		BoardID:   boardID,
		Title:     title,
		Position:  position,
		CreatedBy: createdBy,
	}
	require.NoError(t, repo.Create(context.Background(), card))
	return card
}
