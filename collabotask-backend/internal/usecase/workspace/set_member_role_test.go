package workspace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/workspace"
)

func TestSetMemberRole(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()
	targetID := uuid.New()
	ownerID := uuid.New()

	adminRequester := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleAdmin,
	}
	memberRequester := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleMember,
	}
	adminTarget := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      targetID,
		Role:        entity.WorkspaceRoleAdmin,
	}
	memberTarget := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      targetID,
		Role:        entity.WorkspaceRoleMember,
	}
	ownerTarget := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      ownerID,
		Role:        entity.WorkspaceRoleAdmin,
	}
	ws := &entity.Workspace{
		ID:      workspaceID,
		OwnerID: ownerID,
	}

	validDemoteInput := workspace.SetMemberRoleInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		UserID:      targetID,
		Role:        entity.WorkspaceRoleMember,
	}
	validPromoteInput := workspace.SetMemberRoleInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		UserID:      targetID,
		Role:        entity.WorkspaceRoleAdmin,
	}

	tests := []struct {
		name       string
		input      workspace.SetMemberRoleInput
		setupMocks func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository)
		wantErr    error
		wantErrMsg string
		wantRole   entity.WorkspaceRole
	}{
		{
			name:       "invalid input (zero RequesterID) → validation error",
			input:      workspace.SetMemberRoleInput{WorkspaceID: workspaceID, UserID: targetID, Role: entity.WorkspaceRoleMember},
			setupMocks: func(_ *mocks.MockWorkspaceRepository, _ *mocks.MockWorkspaceMemberRepository) {},
			wantErrMsg: "validation",
		},
		{
			name:  "requester is not workspace admin → ErrNotWorkspaceAdmin",
			input: validDemoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(memberRequester, nil)
			},
			wantErr: domain.ErrNotWorkspaceAdmin,
		},
		{
			name:  "requester lookup fails with DB error → wrapped error (not 403)",
			input: validDemoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db connection lost"))
			},
			wantErrMsg: "failed to get requester",
		},
		{
			name:  "target not a member → ErrMemberNotFound",
			input: validDemoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, targetID).Return(nil, domain.ErrMemberNotFound)
			},
			wantErr: domain.ErrMemberNotFound,
		},
		{
			name: "demote workspace owner → ErrCannotDemoteOwner",
			input: workspace.SetMemberRoleInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				UserID:      ownerID,
				Role:        entity.WorkspaceRoleMember,
			},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, ownerID).Return(ownerTarget, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
			},
			wantErr: domain.ErrCannotDemoteOwner,
		},
		{
			name:  "demote last admin → ErrCannotDemoteLastAdmin",
			input: validDemoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, targetID).Return(adminTarget, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
				wsMemberRepo.EXPECT().CountAdmins(mock.Anything, workspaceID).Return(1, nil)
			},
			wantErr: domain.ErrCannotDemoteLastAdmin,
		},
		{
			name: "self-demotion allowed (not owner, not last admin)",
			input: workspace.SetMemberRoleInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				UserID:      requesterID,
				Role:        entity.WorkspaceRoleMember,
			},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil).Once()
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil).Once()
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
				wsMemberRepo.EXPECT().CountAdmins(mock.Anything, workspaceID).Return(2, nil)
				updatedMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}
				wsMemberRepo.EXPECT().UpdateRole(mock.Anything, workspaceID, requesterID, entity.WorkspaceRoleMember).Return(updatedMember, nil)
			},
			wantRole: entity.WorkspaceRoleMember,
		},
		{
			name:  "idempotent — already has the role → 200 no-op, return member",
			input: validDemoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, targetID).Return(memberTarget, nil)
			},
			wantRole: entity.WorkspaceRoleMember,
		},
		{
			name:  "successful promote",
			input: validPromoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, targetID).Return(memberTarget, nil)
				updatedMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: targetID, Role: entity.WorkspaceRoleAdmin}
				wsMemberRepo.EXPECT().UpdateRole(mock.Anything, workspaceID, targetID, entity.WorkspaceRoleAdmin).Return(updatedMember, nil)
			},
			wantRole: entity.WorkspaceRoleAdmin,
		},
		{
			name:  "successful demote",
			input: validDemoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, targetID).Return(adminTarget, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
				wsMemberRepo.EXPECT().CountAdmins(mock.Anything, workspaceID).Return(2, nil)
				updatedMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: targetID, Role: entity.WorkspaceRoleMember}
				wsMemberRepo.EXPECT().UpdateRole(mock.Anything, workspaceID, targetID, entity.WorkspaceRoleMember).Return(updatedMember, nil)
			},
			wantRole: entity.WorkspaceRoleMember,
		},
		{
			name:  "UpdateRole repo error → wrapped error",
			input: validPromoteInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminRequester, nil)
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, targetID).Return(memberTarget, nil)
				wsMemberRepo.EXPECT().UpdateRole(mock.Anything, workspaceID, targetID, entity.WorkspaceRoleAdmin).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to update member role",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wsRepo := mocks.NewMockWorkspaceRepository(t)
			wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
			userRepo := mocks.NewMockUserRepository(t)
			activityRepo := mocks.NewMockActivityRepository(t)
			activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)

			tt.setupMocks(wsRepo, wsMemberRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
			out, err := uc.SetMemberRole(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRole, out.Member.Role)
		})
	}
}
