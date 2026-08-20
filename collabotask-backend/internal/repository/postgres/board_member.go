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

type boardMemberRepository struct {
	db *pgxpool.Pool
}

const boardMembersListCap = 16

func NewBoardMemberRepository(db *pgxpool.Pool) repository.BoardMemberRepository {
	return &boardMemberRepository{
		db: db,
	}
}

func (bmr *boardMemberRepository) CreateIfAbsent(ctx context.Context, boardMember *entity.BoardMember) (bool, error) {
	var boardID uuid.UUID
	err := bmr.db.QueryRow(
		ctx,
		createBoardMemberIfAbsentQuery,
		boardMember.BoardID,
		boardMember.UserID,
		boardMember.Role,
	).Scan(&boardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT DO NOTHING returned no row: already a member.
			return false, nil
		}
		return false, fmt.Errorf("failed to add member to board: %w", err)
	}

	return true, nil
}

func (bmr *boardMemberRepository) CreateMany(ctx context.Context, boardMembers []*entity.BoardMember) error {
	if len(boardMembers) == 0 {
		return nil
	}

	tx, err := bmr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, member := range boardMembers {
		// RETURNING joined_at so the DB-assigned timestamp flows back onto the
		// struct — callers (e.g. invite broadcast) need the real join time.
		err := tx.QueryRow(ctx, createBoardMemberQuery, member.BoardID, member.UserID, member.Role).
			Scan(&member.JoinedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrBoardAlreadyMember
			}
			return fmt.Errorf("failed to add member to board: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (bmr *boardMemberRepository) Delete(ctx context.Context, boardID, userID uuid.UUID) error {
	result, err := bmr.db.Exec(
		ctx,
		deleteBoardMemberQuery,
		boardID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete member in board: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrBoardMemberNotFound
	}

	return nil
}

func (bmr *boardMemberRepository) GetMembersByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.BoardMember, error) {
	rows, err := bmr.db.Query(
		ctx,
		listMemberByBoardQuery,
		boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query members in board: %w", err)
	}
	defer rows.Close()

	boardMembers := make([]*entity.BoardMember, 0, boardMembersListCap)
	for rows.Next() {
		boardMember := &entity.BoardMember{}
		errScan := rows.Scan(
			&boardMember.BoardID,
			&boardMember.UserID,
			&boardMember.Role,
			&boardMember.JoinedAt,
		)
		if errScan != nil {
			return nil, fmt.Errorf("failed to scan member in board: %w", errScan)
		}

		boardMembers = append(boardMembers, boardMember)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members in board: %w", err)
	}

	return boardMembers, nil
}

func (bmr *boardMemberRepository) GetMemberByBoardAndUser(ctx context.Context, boardID, userID uuid.UUID) (*entity.BoardMember, error) {
	boardMember := &entity.BoardMember{}
	err := bmr.db.QueryRow(
		ctx,
		getMemberByBoardAndUserQuery,
		boardID,
		userID,
	).Scan(
		&boardMember.BoardID,
		&boardMember.UserID,
		&boardMember.Role,
		&boardMember.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBoardMemberNotFound
		}
		return nil, fmt.Errorf("failed to get member by board and user: %w", err)
	}

	return boardMember, nil
}

func (bmr *boardMemberRepository) IsUserExists(ctx context.Context, boardID, userID uuid.UUID) (bool, error) {
	var isExists bool
	err := bmr.db.QueryRow(
		ctx,
		isUserExistsOnBoardQuery,
		boardID,
		userID,
	).Scan(&isExists)

	if err != nil {
		return false, fmt.Errorf("failed to check if member exists in board: %w", err)
	}

	return isExists, nil
}

func (bmr *boardMemberRepository) RemoveWithParticipationCascade(ctx context.Context, boardID, userID uuid.UUID) ([]repository.AffectedCard, error) {
	tx, err := bmr.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx, deleteBoardMemberForCascadeQuery, boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove board member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrBoardMemberNotFound
	}

	rows, err := tx.Query(ctx, unassignBoardCardsForUserQuery, boardID, userID)
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

func (bmr *boardMemberRepository) TransferOwnership(ctx context.Context, boardID, newOwnerID uuid.UUID) (*uuid.UUID, error) {
	tx, err := bmr.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var fromUserID uuid.UUID
	err = tx.QueryRow(ctx, demoteCurrentOwnerQuery, boardID).Scan(&fromUserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to demote current owner: %w", err)
	}
	var fromPtr *uuid.UUID
	if err == nil {
		fromPtr = &fromUserID
	}

	result, err := tx.Exec(ctx, promoteNewOwnerQuery, boardID, newOwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to promote new owner: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrBoardMemberNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit ownership transfer: %w", err)
	}

	return fromPtr, nil
}
