package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type BoardRepository interface {
	Create(ctx context.Context, board *entity.Board) error
	Update(ctx context.Context, board *entity.Board) error
	Delete(ctx context.Context, boardID uuid.UUID) error
	GetByID(ctx context.Context, boardID uuid.UUID) (*entity.Board, error)
	GetUserBoardsInWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) ([]*entity.BoardListItem, error)
	SetArchived(ctx context.Context, boardID uuid.UUID, archived bool) error
}
