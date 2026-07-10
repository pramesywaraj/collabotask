package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type BoardMemberRepository interface {
	// CreateIfAbsent idempotently adds a board member. It returns true when a
	// new row was inserted and false when the user was already a member.
	CreateIfAbsent(ctx context.Context, boardMember *entity.BoardMember) (bool, error)
	CreateMany(ctx context.Context, boardMembers []*entity.BoardMember) error
	Delete(ctx context.Context, boardID, userID uuid.UUID) error
	GetMemberByBoardAndUser(ctx context.Context, boardID, userID uuid.UUID) (*entity.BoardMember, error)
	GetMembersByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.BoardMember, error)
	IsUserExists(ctx context.Context, boardID, userID uuid.UUID) (bool, error)
}
