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
	tempWorkspace := &entity.Workspace{}

	err := w.db.QueryRow(ctx, getWorkspaceByIdQuery, workspaceID).Scan(
		&tempWorkspace.ID,
		&tempWorkspace.Name,
		&description,
		&tempWorkspace.OwnerID,
		&tempWorkspace.CreatedAt,
		&tempWorkspace.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workspace not found")
		}

		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	tempWorkspace.Description = description

	return tempWorkspace, nil
}

func (w *WorkspaceRepositoryImpl) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error) {
	rows, err := w.db.Query(ctx, getUserWorkspacesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user workspaces: %w", err)
	}
	defer rows.Close()

	tempWorkspaces := []*entity.Workspace{}
	for rows.Next() {
		var description *string
		tempWorkspace := &entity.Workspace{}

		err := rows.Scan(
			&tempWorkspace.ID,
			&tempWorkspace.Name,
			&description,
			&tempWorkspace.OwnerID,
			&tempWorkspace.CreatedAt,
			&tempWorkspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user's workspace: %w", err)
		}

		tempWorkspace.Description = description
		tempWorkspaces = append(tempWorkspaces, tempWorkspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user's workspaces: %w", err)
	}

	return tempWorkspaces, nil
}
