package workspace

import (
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (wu *WorkspaceUseCaseImpl) WorkspaceDetail(ctx context.Context, input WorkspaceDetailInput) (*WorkspaceDetailOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("workspace detail validation failed: %w", err)
	}

	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || requesterMember == nil || requesterMember.IsEmpty() {
		return nil, ErrUserNotInWorkspace
	}

	workspace, err := wu.workspaceRepo.GetByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	members, err := wu.workspaceMemberRepo.ListMemberByWorkspace(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspace members: %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	usersMap, err := wu.userRepo.GetByIds(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch member details: %w", err)
	}

	workspaceMembers := make([]WorkspaceMemberDTO, 0, len(members))
	for _, member := range members {
		user, ok := usersMap[member.UserID]
		if !ok || user == nil {
			return nil, ErrUserNotFound
		}
		workspaceMembers = append(workspaceMembers, workspaceMemberToDTO(member, user))
	}

	output := &WorkspaceDetailDTO{
		WorkspaceDTO: workspaceToDTO(workspace),
		UserRole:     requesterMember.Role,
		Members:      workspaceMembers,
	}

	return &WorkspaceDetailOutput{
		Workspace: *output,
	}, nil
}
