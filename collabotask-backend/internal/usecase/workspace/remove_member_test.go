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

func TestRemoveMember(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()
	targetID := uuid.New()

	adminMember := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleAdmin,
	}
	regularMember := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleMember,
	}

	validInput := workspace.RemoveMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		UserID:      targetID,
	}

	tests := []struct {
		name       string
		input      workspace.RemoveMemberInput
		setupMocks func(wsMemberRepo *mocks.MockWorkspaceMemberRepository)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input (zero RequesterID) → validation error",
			input:      workspace.RemoveMemberInput{RequesterID: uuid.Nil, WorkspaceID: workspaceID, UserID: targetID},
			setupMocks: func(wsMemberRepo *mocks.MockWorkspaceMemberRepository) {},
			wantErrMsg: "validation",
		},
		{
			name:  "requester is not workspace admin → ErrNotWorkspaceAdmin",
			input: validInput,
			setupMocks: func(wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
			},
			wantErr: domain.ErrNotWorkspaceAdmin,
		},
		{
			name: "RequesterID == UserID (requester IS admin) → ErrCannotRemoveYourself",
			input: workspace.RemoveMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				UserID:      requesterID,
			},
			setupMocks: func(wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
			},
			wantErr: domain.ErrCannotRemoveYourself,
		},
		{
			name:  "target member not found → ErrMemberNotFound",
			input: validInput,
			setupMocks: func(wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				wsMemberRepo.EXPECT().Delete(mock.Anything, workspaceID, targetID).Return(domain.ErrMemberNotFound)
			},
			wantErr: domain.ErrMemberNotFound,
		},
		{
			name:  "Delete fails with unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				wsMemberRepo.EXPECT().Delete(mock.Anything, workspaceID, targetID).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to remove member from the workspace",
		},
		{
			name:  "success",
			input: validInput,
			setupMocks: func(wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				wsMemberRepo.EXPECT().Delete(mock.Anything, workspaceID, targetID).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wsRepo := mocks.NewMockWorkspaceRepository(t)
			wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
			userRepo := mocks.NewMockUserRepository(t)

			tt.setupMocks(wsMemberRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo)
			err := uc.RemoveMember(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			assert.NoError(t, err)
		})
	}
}
