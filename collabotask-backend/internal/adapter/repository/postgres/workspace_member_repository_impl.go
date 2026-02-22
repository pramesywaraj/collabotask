package postgres

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceMemberRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewWorkspaceMemberRepository(db *pgxpool.Pool) repository.WorkspaceMemberRepository {
	return &WorkspaceMemberRepositoryImpl{
		db: db,
	}
}

func (wm *WorkspaceMemberRepositoryImpl) Create(ctx context.Context, workspaceMember *entity.WorkspaceMember) error {
	err := wm.db.QueryRow(
		ctx,
		createWorkspaceMemberQuery,
		workspaceMember.WorkspaceID,
		workspaceMember.UserID,
		workspaceMember.Role,
	).Scan(
		&workspaceMember.WorkspaceID,
		&workspaceMember.UserID,
		&workspaceMember.Role,
		&workspaceMember.JoinedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("user already in workspace")
			}
		}

		return fmt.Errorf("failed to add member to workspace: %w", err)
	}

	return nil
}

func (wm *WorkspaceMemberRepositoryImpl) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	result, err := wm.db.Exec(ctx, deleteWorkspaceMemberQuery, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (wm *WorkspaceMemberRepositoryImpl) GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.WorkspaceMember, error) {
	workspaceMember := &entity.WorkspaceMember{}
	err := wm.db.QueryRow(
		ctx,
		getByWorkspaceAndUserQuery,
		workspaceID,
		userID,
	).Scan(
		&workspaceMember.WorkspaceID,
		&workspaceMember.UserID,
		&workspaceMember.Role,
		&workspaceMember.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("member not found")
		}

		return nil, fmt.Errorf("failed to get member in workspace: %w", err)
	}

	return workspaceMember, nil
}

func (wm *WorkspaceMemberRepositoryImpl) ListMemberByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error) {
	rows, err := wm.db.Query(
		ctx,
		listMemberByWorkspaceQuery,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query members in workspace: %w", err)
	}

	defer rows.Close()

	workspaceMembers := []*entity.WorkspaceMember{}
	for rows.Next() {
		workspaceMember := &entity.WorkspaceMember{}
		errScan := rows.Scan(
			&workspaceMember.WorkspaceID,
			&workspaceMember.UserID,
			&workspaceMember.Role,
			&workspaceMember.JoinedAt,
		)
		if errScan != nil {
			return nil, fmt.Errorf("failed to scan member in workspace: %w", errScan)
		}

		workspaceMembers = append(workspaceMembers, workspaceMember)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members in workspace: %w", err)
	}

	return workspaceMembers, nil
}

func (wm *WorkspaceMemberRepositoryImpl) IsUserExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := wm.db.QueryRow(
		ctx,
		isUserExistsOnWorkspaceQuery,
		workspaceID,
		userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if member exists in workspace: %w", err)
	}

	return exists, nil
}
