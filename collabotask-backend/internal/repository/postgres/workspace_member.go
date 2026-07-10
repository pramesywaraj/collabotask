package postgres

import (
	"collabotask/internal/domain"
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

type workspaceMemberRepository struct {
	db *pgxpool.Pool
}

func NewWorkspaceMemberRepository(db *pgxpool.Pool) repository.WorkspaceMemberRepository {
	return &workspaceMemberRepository{
		db: db,
	}
}

func (wm *workspaceMemberRepository) Create(ctx context.Context, workspaceMember *entity.WorkspaceMember) error {
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
				return domain.ErrAlreadyMember
			}
		}

		return fmt.Errorf("failed to add member to workspace: %w", err)
	}

	return nil
}

func (wm *workspaceMemberRepository) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	result, err := wm.db.Exec(ctx, deleteWorkspaceMemberQuery, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}

	return nil
}

func (wm *workspaceMemberRepository) GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.WorkspaceMember, error) {
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
			return nil, domain.ErrMemberNotFound
		}

		return nil, fmt.Errorf("failed to get member in workspace: %w", err)
	}

	return workspaceMember, nil
}

func (wm *workspaceMemberRepository) GetMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error) {
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

func (wm *workspaceMemberRepository) IsUserExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
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

func (wm *workspaceMemberRepository) UpdateRole(ctx context.Context, workspaceID, userID uuid.UUID, role entity.WorkspaceRole) (*entity.WorkspaceMember, error) {
	member := &entity.WorkspaceMember{}
	err := wm.db.QueryRow(ctx, updateWorkspaceMemberRoleQuery, workspaceID, userID, role).Scan(
		&member.WorkspaceID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	return member, nil
}

func (wm *workspaceMemberRepository) CountAdmins(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := wm.db.QueryRow(ctx, countWorkspaceAdminsQuery, workspaceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workspace admins: %w", err)
	}

	return count, nil
}

func (wm *workspaceMemberRepository) RemoveWithParticipationCascade(ctx context.Context, workspaceID, userID uuid.UUID) ([]repository.AffectedCard, error) {
	tx, err := wm.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, deleteWorkspaceMemberQuery, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove workspace member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrMemberNotFound
	}

	if _, err = tx.Exec(ctx, deleteBoardMembershipsForUserQuery, workspaceID, userID); err != nil {
		return nil, fmt.Errorf("failed to remove board memberships: %w", err)
	}

	rows, err := tx.Query(ctx, unassignCardsForUserQuery, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to unassign cards: %w", err)
	}
	defer rows.Close()

	var affected []repository.AffectedCard
	for rows.Next() {
		var card repository.AffectedCard
		if err := rows.Scan(&card.CardID, &card.ColumnID); err != nil {
			return nil, fmt.Errorf("failed to scan affected card: %w", err)
		}
		affected = append(affected, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating affected cards: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit cascade transaction: %w", err)
	}

	return affected, nil
}
