package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCase) GetBoardDetail(ctx context.Context, input GetBoardDetailInput) (*GetBoardDetailOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate board detail input: %w", err)
	}

	access, err := bu.boardAccessChecker.Resolve(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	board := access.Board
	boardMembership := access.BoardMember

	members, err := bu.boardMemberRepo.GetMembersByBoard(ctx, input.BoardID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch board members: %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	users, err := bu.userRepo.GetByIds(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members details: %w", err)
	}

	boardMembers := make([]BoardMember, 0, len(members))
	for _, member := range members {
		user, ok := users[member.UserID]
		if !ok || user == nil {
			return nil, domain.ErrUserNotFound
		}
		boardMembers = append(boardMembers, BoardMember{
			UserID:    member.UserID,
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
			Role:      member.Role,
			JoinedAt:  member.JoinedAt,
		})
	}

	var userRole *entity.BoardRole
	accessStatus := entity.BoardJoined
	if boardMembership != nil && !boardMembership.IsEmpty() {
		role := boardMembership.Role
		userRole = &role
	} else if board.CreatedBy == input.RequesterID {
		r := entity.BoardRoleOwner
		userRole = &r
	} else {
		accessStatus = entity.BoardCanJoin
	}

	return &GetBoardDetailOutput{
		Board: BoardDetail{
			Board:        board,
			UserRole:     userRole,
			AccessStatus: accessStatus,
			Members:      boardMembers,
		},
	}, nil
}
