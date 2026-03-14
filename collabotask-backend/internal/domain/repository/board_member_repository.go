package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type BoardMemberRepository interface {
	Create(ctx context.Context, boardMember *entity.BoardMember) error
	CreateMany(ctx context.Context, boardMembers []*entity.BoardMember) error
	Delete(ctx context.Context, boardID, userID uuid.UUID) error
	GetMemberByBoardAndUser(ctx context.Context, boardID, userID uuid.UUID) (*entity.BoardMember, error)
	ListMemberByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.BoardMember, error)
	IsUserExists(ctx context.Context, boardID, userID uuid.UUID) (bool, error)
}
