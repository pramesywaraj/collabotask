package postgres

import (
	"collabotask/internal/domain"
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

type workspaceRepository struct {
	base
}

func NewWorkspaceRepository(pool *pgxpool.Pool) repository.WorkspaceRepository {
	return &workspaceRepository{base: base{pool: pool}}
}

const workspacesCap = 16

func (w *workspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	var description *string
	if workspace.Description != nil && *workspace.Description != "" {
		description = workspace.Description
	}

	err := w.exec(ctx).QueryRow(
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

func (w *workspaceRepository) CreateWithOwner(ctx context.Context, workspace *entity.Workspace, ownerID uuid.UUID) error {
	return w.tx(ctx, func(ctx context.Context) error {
		var description *string
		if workspace.Description != nil && *workspace.Description != "" {
			description = workspace.Description
		}

		err := w.exec(ctx).QueryRow(
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
					return domain.ErrConstraintViolation
				}
			}
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		_, err = w.exec(ctx).Exec(ctx, createWorkspaceMemberQuery, workspace.ID, ownerID, entity.WorkspaceRoleAdmin)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrAlreadyMember
			}
			return fmt.Errorf("failed to add owner to workspace: %w", err)
		}

		return nil
	})
}

func (w *workspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	var name *string
	if workspace.Name != "" {
		name = &workspace.Name
	}

	var description *string
	if workspace.Description != nil && *workspace.Description != "" {
		description = workspace.Description
	}

	updatedAt := time.Now()

	err := w.exec(ctx).QueryRow(
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
			return domain.ErrWorkspaceNotFound
		}

		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

func (w *workspaceRepository) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	result, err := w.exec(ctx).Exec(ctx, deleteWorkspaceQuery, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrWorkspaceNotFound
	}

	return nil
}

func (w *workspaceRepository) GetByID(ctx context.Context, workspaceID uuid.UUID) (*entity.Workspace, error) {
	var description *string
	workspace := &entity.Workspace{}

	err := w.exec(ctx).QueryRow(ctx, getWorkspaceByIdQuery, workspaceID).Scan(
		&workspace.ID,
		&workspace.Name,
		&description,
		&workspace.OwnerID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkspaceNotFound
		}

		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	workspace.Description = description

	return workspace, nil
}

func (w *workspaceRepository) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.WorkspaceListItem, error) {
	rows, err := w.exec(ctx).Query(ctx, getUserWorkspacesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user workspaces: %w", err)
	}
	defer rows.Close()

	workspaces := make([]*entity.WorkspaceListItem, 0, workspacesCap)
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
