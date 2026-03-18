package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (bu *BoardUseCaseImpl) ListWorkspaceInviteesForBoard(ctx context.Context, input ListWorkspaceInviteesForBoardInput) (*ListWorkspaceInviteesForBoardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate list workspace invitees for board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to fetch board: %w", err)
	}
	if board == nil || board.WorkspaceID != input.WorkspaceID || board.IsArchived {
		return nil, domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence in workspace: %w", err)
	}
	if workspaceMember == nil || workspaceMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester permission: %w", err)
	}

	canManage := canManageBoardMembers(board.CreatedBy, input.RequesterID, boardMember, workspaceMember)
	if !canManage {
		return nil, domain.ErrBoardPermissionDenied
	}

	workspaceMembers, err := bu.workspaceMemberRepo.ListMemberByWorkspace(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspace members: %w", err)
	}

	boardMembers, err := bu.boardMemberRepo.ListMemberByBoard(ctx, input.BoardID)
	if err != nil {
		return nil, fmt.Errorf("failed to list board members: %w", err)
	}

	boardMemberIDs := make(map[uuid.UUID]bool, len(boardMembers))
	for _, bm := range boardMembers {
		boardMemberIDs[bm.UserID] = true
	}

	userIDs := make([]uuid.UUID, 0, len(workspaceMembers))
	for _, wm := range workspaceMembers {
		userIDs = append(userIDs, wm.UserID)
	}

	users, err := bu.userRepo.GetByIds(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user details: %w", err)
	}

	members := make([]dto.BoardInviteeDTO, 0, len(workspaceMembers))
	for _, wm := range workspaceMembers {
		user, ok := users[wm.UserID]
		if !ok || user == nil {
			continue
		}
		members = append(members, dto.BoardInviteeDTO{
			UserID:        wm.UserID,
			Email:         user.Email,
			Name:          user.Name,
			AvatarURL:     user.AvatarURL,
			WorkspaceRole: wm.Role,
			IsBoardMember: boardMemberIDs[wm.UserID],
		})
	}

	return &ListWorkspaceInviteesForBoardOutput{
		Members: members,
	}, nil
}
