package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCaseImpl) GetBoardDetail(ctx context.Context, input GetBoardDetailInput) (*GetBoardDetailOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate board detail input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil || board == nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to fetch board: %w", err)
	}
	if board.IsArchived {
		return nil, domain.ErrBoardNotFound
	}

	workspaceMembership, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, board.WorkspaceID, input.RequesterID)
	if err != nil || workspaceMembership == nil || workspaceMembership.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMembership, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if !errors.Is(err, domain.ErrBoardMemberNotFound) {
			return nil, fmt.Errorf("failed to fetch board membership: %w", err)
		}

		boardMembership = nil
	}

	hasAccess := workspaceMembership.IsAdmin() || board.CreatedBy == input.RequesterID || (boardMembership != nil && !boardMembership.IsEmpty())
	if !hasAccess {
		return nil, domain.ErrBoardAccessDenied
	}

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
	var accessStatus entity.BoardAccessStatus = entity.BoardJoined
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
