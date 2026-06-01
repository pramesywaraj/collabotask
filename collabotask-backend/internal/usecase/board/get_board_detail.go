package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCaseImpl) GetBoardDetail(ctx context.Context, input GetBoardDetailInput) (*GetBoardDetailOutput, error) {
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

	boardMembers := make([]dto.BoardMemberDTO, 0, len(members))
	for _, member := range members {
		user, ok := users[member.UserID]
		if !ok || user == nil {
			return nil, domain.ErrUserNotFound
		}
		boardMembers = append(boardMembers, dto.BoardMemberToDTO(member, user))
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
		Board: dto.BoardDetailDTO{
			BoardDTO:     dto.BoardToDTO(board),
			UserRole:     userRole,
			AccessStatus: accessStatus,
			Members:      boardMembers,
		},
	}, nil
}
