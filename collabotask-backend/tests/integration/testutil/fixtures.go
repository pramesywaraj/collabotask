package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

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

// CreateTestBoard inserts a board (with ownerID as BOARD_OWNER) and returns it.
func (f *Fixtures) CreateTestBoard(t *testing.T, workspaceID, ownerID uuid.UUID) *entity.Board {
	t.Helper()
	repo := pgRepos.NewBoardRepository(f.pool)
	board := &entity.Board{
		WorkspaceID:     workspaceID,
		Title:           fmt.Sprintf("Board-%s", uuid.New().String()[:8]),
		CreatedBy:       ownerID,
		BackgroundColor: "#0079BF",
		Visibility:      entity.BoardVisibilityWorkspace,
	}
	require.NoError(t, repo.CreateWithOwner(context.Background(), board, ownerID))
	return board
}

// CreateTestColumn inserts a column in the given board and returns it.
func (f *Fixtures) CreateTestColumn(t *testing.T, boardID uuid.UUID) *entity.Column {
	t.Helper()
	repo := pgRepos.NewColumnRepository(f.pool)
	col := &entity.Column{
		BoardID:  boardID,
		Title:    fmt.Sprintf("Column-%s", uuid.New().String()[:8]),
		Position: 1000,
	}
	require.NoError(t, repo.Create(context.Background(), col))
	return col
}

// CreateTestCard inserts a card in the given column/board and returns it.
func (f *Fixtures) CreateTestCard(t *testing.T, columnID, boardID, createdBy uuid.UUID) *entity.Card {
	t.Helper()
	repo := pgRepos.NewCardRepository(f.pool)
	card := &entity.Card{
		ColumnID:  columnID,
		BoardID:   boardID,
		Title:     fmt.Sprintf("Card-%s", uuid.New().String()[:8]),
		Position:  1000,
		CreatedBy: createdBy,
	}
	require.NoError(t, repo.Create(context.Background(), card))
	return card
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
