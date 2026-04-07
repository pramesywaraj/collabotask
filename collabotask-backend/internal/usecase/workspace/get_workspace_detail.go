package workspace

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (wu *WorkspaceUseCaseImpl) GetWorkspaceDetail(ctx context.Context, input GetWorkspaceDetailInput) (*GetWorkspaceDetailOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("workspace detail validation failed: %w", err)
	}

	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || requesterMember == nil || requesterMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	workspace, err := wu.workspaceRepo.GetByID(ctx, input.WorkspaceID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) || workspace == nil {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to fetch workspace: %w", err)
	}

	members, err := wu.workspaceMemberRepo.GetMembersByWorkspace(ctx, input.WorkspaceID)
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

	workspaceMembers := make([]dto.WorkspaceMemberDTO, 0, len(members))
	for _, member := range members {
		user, ok := usersMap[member.UserID]

		//TODO: Need to check whether it can be optimized
		// Context: if one fail how's the others?
		if !ok || user == nil {
			return nil, domain.ErrUserNotFound
		}
		workspaceMembers = append(workspaceMembers, dto.WorkspaceMemberToDTO(member, user))
	}

	output := &dto.WorkspaceDetailDTO{
		WorkspaceDTO: dto.WorkspaceToDTO(workspace),
		UserRole:     requesterMember.Role,
		Members:      workspaceMembers,
	}

	return &GetWorkspaceDetailOutput{
		Workspace: *output,
	}, nil
}
