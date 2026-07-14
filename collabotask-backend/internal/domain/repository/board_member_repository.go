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
	// TransferOwnership atomically moves the sole BOARD_OWNER title to newOwnerID:
	// demotes the current owner by role and promotes newOwnerID by id, in one tx.
	// Returns the demoted owner's id (nil when the board has no owner) for the
	// future OWNERSHIP_TRANSFERRED broadcast — returned now so step ④ is a pure wire-up.
	TransferOwnership(ctx context.Context, boardID, newOwnerID uuid.UUID) (fromUserID *uuid.UUID, err error)
}
