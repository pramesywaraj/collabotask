package postgres

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewWorkspaceRepository(db *pgxpool.Pool) repository.WorkspaceRepository {
	return &WorkspaceRepositoryImpl{db: db}
}

func (w *WorkspaceRepositoryImpl) Create(ctx context.Context, workspace *entity.Workspace) error {
	var description *string
	if workspace.Description != nil && *workspace.Description != "" {
		description = workspace.Description
	}

	err := w.db.QueryRow(
		ctx,
		createWorkspaceQuery,
		workspace.Name,
		description,
		workspace.OwnerID,
	).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.OwnerID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("workspace constraint violation")
			}
		}

		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

func (w *WorkspaceRepositoryImpl) CreateWithOwner(ctx context.Context, workspace *entity.Workspace, ownerID uuid.UUID) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin create workspace with owner transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var description *string
	if workspace.Description != nil && *workspace.Description != "" {
		description = workspace.Description
	}

	err = tx.QueryRow(
		ctx,
		createWorkspaceQuery,
		workspace.Name,
		description,
		workspace.OwnerID,
	).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.OwnerID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("workspace constraint violation")
			}
		}
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	_, err = tx.Exec(ctx, createWorkspaceMemberQuery, workspace.ID, ownerID, entity.WorkspaceRoleAdmin)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("user already in workspace")
		}
		return fmt.Errorf("failed to add owner to workspace: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (w *WorkspaceRepositoryImpl) Update(ctx context.Context, workspace *entity.Workspace) error {
	var name *string
	if workspace.Name != "" {
		name = &workspace.Name
	}

	var description *string
	if workspace.Description != nil && *workspace.Description != "" {
		description = workspace.Description
	}

	updatedAt := time.Now()

	err := w.db.QueryRow(
		ctx,
		updateWorkspaceQuery,
		name,
		description,
		updatedAt,
		workspace.ID,
	).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.OwnerID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("workspace not found")
		}

		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

func (w *WorkspaceRepositoryImpl) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	result, err := w.db.Exec(ctx, deleteWorkspaceQuery, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("workspace not found")
	}

	return nil
}

func (w *WorkspaceRepositoryImpl) GetByID(ctx context.Context, workspaceID uuid.UUID) (*entity.Workspace, error) {
	var description *string
	workspace := &entity.Workspace{}

	err := w.db.QueryRow(ctx, getWorkspaceByIdQuery, workspaceID).Scan(
		&workspace.ID,
		&workspace.Name,
		&description,
		&workspace.OwnerID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workspace not found")
		}

		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	workspace.Description = description

	return workspace, nil
}

func (w *WorkspaceRepositoryImpl) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.WorkspaceListItem, error) {
	rows, err := w.db.Query(ctx, getUserWorkspacesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user workspaces: %w", err)
	}
	defer rows.Close()

	workspaces := []*entity.WorkspaceListItem{}
	for rows.Next() {
		var description *string
		var role string
		var memberCount, boardCount int64
		workspace := &entity.WorkspaceListItem{}

		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&description,
			&workspace.OwnerID,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&role,
			&memberCount,
			&boardCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user's workspace: %w", err)
		}

		workspace.Description = description
		workspace.Role = entity.WorkspaceRole(role)
		workspace.MemberCount = uint(memberCount)
		workspace.BoardCount = uint(boardCount)

		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user's workspaces: %w", err)
	}

	return workspaces, nil
}
