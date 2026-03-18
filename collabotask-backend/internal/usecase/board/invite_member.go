package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (bu *BoardUseCaseImpl) InviteMember(ctx context.Context, input InviteMemberInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate invite member input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil || board == nil {
		return domain.ErrBoardNotFound
	}
	if board.WorkspaceID != input.WorkspaceID {
		return domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || workspaceMember == nil || workspaceMember.IsEmpty() {
		return domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return fmt.Errorf("failed to check requester permission: %w", err)
	}

	canManage := canManageBoardMembers(board.CreatedBy, input.RequesterID, boardMember, workspaceMember)
	if !canManage {
		return domain.ErrBoardPermissionDenied
	}

	users, err := bu.userRepo.GetByIds(ctx, input.UserIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch users data: %w", err)
	}

	var membersToAdd []*entity.BoardMember
	for _, userID := range input.UserIDs {
		if userID == input.RequesterID {
			continue
		}

		if _, ok := users[userID]; !ok {
			return domain.ErrUserNotFound
		}

		existsInWorkspace, err := bu.workspaceMemberRepo.IsUserExists(ctx, input.WorkspaceID, userID)
		if err != nil {
			return fmt.Errorf("failed to check workspace membership for this user: %s, with error: %w", userID, err)
		}
		if !existsInWorkspace {
			return domain.ErrUserNotInWorkspace
		}

		existsInBoard, err := bu.boardMemberRepo.IsUserExists(ctx, input.BoardID, userID)
		if err != nil {
			return fmt.Errorf("failed to check board membership for this user: %s, with error: %w", userID, err)
		}
		if existsInBoard {
			return domain.ErrBoardAlreadyMember
		}

		membersToAdd = append(membersToAdd, &entity.BoardMember{
			BoardID: input.BoardID,
			UserID:  userID,
			Role:    entity.BoardRoleMember,
		})
	}

	if len(membersToAdd) == 0 {
		return domain.ErrBoardNoMembersToInvite
	}

	if err := bu.boardMemberRepo.CreateMany(ctx, membersToAdd); err != nil {
		return err
	}

	return nil
}
