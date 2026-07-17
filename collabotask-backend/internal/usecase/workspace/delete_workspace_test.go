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

func TestDeleteWorkspace(t *testing.T) {
	ownerID := uuid.New()
	nonOwnerID := uuid.New()
	workspaceID := uuid.New()

	ws := &entity.Workspace{
		ID:      workspaceID,
		OwnerID: ownerID,
	}

	validInput := workspace.DeleteWorkspaceInput{
		RequesterID: ownerID,
		WorkspaceID: workspaceID,
	}

	tests := []struct {
		name       string
		input      workspace.DeleteWorkspaceInput
		setupMocks func(wsRepo *mocks.MockWorkspaceRepository)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input (zero RequesterID) → validation error",
			input:      workspace.DeleteWorkspaceInput{WorkspaceID: workspaceID},
			setupMocks: func(_ *mocks.MockWorkspaceRepository) {},
			wantErrMsg: "validation",
		},
		{
			name: "requester is not the workspace owner → ErrNotWorkspaceOwner",
			input: workspace.DeleteWorkspaceInput{
				RequesterID: nonOwnerID,
				WorkspaceID: workspaceID,
			},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
			},
			wantErr: domain.ErrNotWorkspaceOwner,
		},
		{
			name:  "workspace not found → ErrWorkspaceNotFound",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(nil, domain.ErrWorkspaceNotFound)
			},
			wantErr: domain.ErrWorkspaceNotFound,
		},
		{
			name:  "Delete repo error → wrapped error",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
				wsRepo.EXPECT().Delete(mock.Anything, workspaceID).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to delete workspace",
		},
		{
			name:  "successful delete",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
				wsRepo.EXPECT().Delete(mock.Anything, workspaceID).Return(nil)
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
			activityRepo := mocks.NewMockActivityRepository(t)
			activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)

			tt.setupMocks(wsRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
			err := uc.DeleteWorkspace(context.Background(), tt.input)

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
