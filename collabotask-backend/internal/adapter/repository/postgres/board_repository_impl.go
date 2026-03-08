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

type BoardRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewBoardRepository(db *pgxpool.Pool) repository.BoardRepository {
	return &BoardRepositoryImpl{db: db}
}

const boardsCap = 16

func (br *BoardRepositoryImpl) Create(ctx context.Context, board *entity.Board) error {
	var description *string

	if board.Description != nil && *board.Description != "" {
		description = board.Description
	}

	err := br.db.QueryRow(
		ctx,
		createBoardQuery,
		board.WorkspaceID,
		board.Title,
		description,
		board.CreatedBy,
		board.BackgroundColor,
	).Scan(
		&board.ID,
		&board.WorkspaceID,
		&board.Title,
		&board.Description,
		&board.CreatedBy,
		&board.IsArchived,
		&board.BackgroundColor,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrConstraintViolation
			}
		}
		return fmt.Errorf("failed to create board: %w", err)
	}

	return nil
}

func (br *BoardRepositoryImpl) Update(ctx context.Context, board *entity.Board) error {
	var title *string
	if board.Title != "" {
		title = &board.Title
	}

	var description *string
	if board.Description != nil && *board.Description != "" {
		description = board.Description
	}

	var backgroundColor *string
	if board.BackgroundColor != "" {
		backgroundColor = &board.BackgroundColor
	}

	updatedAt := time.Now()

	err := br.db.QueryRow(
		ctx,
		updateBoardQuery,
		title,
		description,
		backgroundColor,
		updatedAt,
		board.ID,
	).Scan(
		&board.ID,
		&board.WorkspaceID,
		&board.Title,
		&board.Description,
		&board.CreatedBy,
		&board.IsArchived,
		&board.BackgroundColor,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrBoardNotFound
		}
		return fmt.Errorf("failed to update board: %w", err)
	}

	return nil
}

func (br *BoardRepositoryImpl) Delete(ctx context.Context, boardID uuid.UUID) error {
	result, err := br.db.Exec(
		ctx,
		deleteBoardQuery,
		boardID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete board: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrBoardNotFound
	}

	return nil
}

func (br *BoardRepositoryImpl) GetByID(ctx context.Context, boardID uuid.UUID) (*entity.Board, error) {
	var description *string
	board := &entity.Board{}

	err := br.db.QueryRow(
		ctx,
		getBoardByIDQuery,
		boardID,
	).Scan(
		&board.ID,
		&board.WorkspaceID,
		&board.Title,
		&description,
		&board.CreatedBy,
		&board.IsArchived,
		&board.BackgroundColor,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	board.Description = description

	return board, nil
}

func (br *BoardRepositoryImpl) GetUserBoardsInWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) ([]*entity.BoardListItem, error) {
	rows, err := br.db.Query(
		ctx,
		getUserBoardsInWorkspace,
		workspaceID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query user boards in workspace: %w", err)
	}
	defer rows.Close()

	boards := make([]*entity.BoardListItem, 0, boardsCap)
	for rows.Next() {
		board := &entity.BoardListItem{}
		var description *string
		var accessStatus *string
		var role *string
		var memberCount int64

		err := rows.Scan(
			&board.ID,
			&board.WorkspaceID,
			&board.Title,
			&description,
			&board.CreatedBy,
			&board.IsArchived,
			&board.BackgroundColor,
			&board.CreatedAt,
			&board.UpdatedAt,
			&role,
			&accessStatus,
			&memberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan board: %w", err)
		}

		board.Description = description
		if accessStatus != nil {
			board.AccessStatus = entity.BoardAccessStatus(*accessStatus)
		}
		if role != nil {
			board.UserRole = entity.BoardRole(*role)
		}
		board.MemberCount = uint(memberCount)

		boards = append(boards, board)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user's board in workspace: %w", err)
	}

	return boards, nil
}

func (br *BoardRepositoryImpl) SetArchived(ctx context.Context, boardID uuid.UUID, archived bool) error {
	result, err := br.db.Exec(
		ctx,
		setBoardArchivedQuery,
		boardID,
		archived,
	)
	if err != nil {
		return fmt.Errorf("failed to set is_archived flag for board: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrBoardNotFound
	}

	return nil
}
