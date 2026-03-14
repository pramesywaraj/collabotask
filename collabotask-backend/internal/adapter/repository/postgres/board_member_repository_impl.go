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

type BoardMemberRepositoryImpl struct {
	db *pgxpool.Pool
}

const boardMembersListCap = 16

func NewBoardMemberRepository(db *pgxpool.Pool) repository.BoardMemberRepository {
	return &BoardMemberRepositoryImpl{
		db: db,
	}
}

func (bmr *BoardMemberRepositoryImpl) Create(ctx context.Context, boardMember *entity.BoardMember) error {
	_, err := bmr.db.Exec(
		ctx,
		createBoardMemberQuery,
		boardMember.BoardID,
		boardMember.UserID,
		boardMember.Role,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrBoardAlreadyMember
			}
		}
		return fmt.Errorf("failed to add member to board: %w", err)
	}

	return nil
}

func (bmr *BoardMemberRepositoryImpl) CreateMany(ctx context.Context, boardMembers []*entity.BoardMember) error {
	if len(boardMembers) == 0 {
		return nil
	}

	tx, err := bmr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, member := range boardMembers {
		_, err := tx.Exec(ctx, createBoardMemberQuery, member.BoardID, member.UserID, member.Role)
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

func (bmr *BoardMemberRepositoryImpl) Delete(ctx context.Context, boardID, userID uuid.UUID) error {
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

func (bmr *BoardMemberRepositoryImpl) ListMemberByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.BoardMember, error) {
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

func (bmr *BoardMemberRepositoryImpl) GetMemberByBoardAndUser(ctx context.Context, boardID, userID uuid.UUID) (*entity.BoardMember, error) {
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

func (bmr *BoardMemberRepositoryImpl) IsUserExists(ctx context.Context, boardID, userID uuid.UUID) (bool, error) {
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
