package workspace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/workspace"
)

func TestGetWorkspaces(t *testing.T) {
	userID := uuid.New()
	wsID := uuid.New()

	tests := []struct {
		name       string
		setupMocks func(wsRepo *mocks.MockWorkspaceRepository)
		wantErrMsg string
		checkOut   func(t *testing.T, out *workspace.GetWorkspacesOutput)
	}{
		{
			name: "GetUserWorkspaces fails → wrapped error",
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetUserWorkspaces(mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch user workspaces",
		},
		{
			name: "success — empty list",
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetUserWorkspaces(mock.Anything, userID).Return([]*entity.WorkspaceListItem{}, nil)
			},
			checkOut: func(t *testing.T, out *workspace.GetWorkspacesOutput) {
				t.Helper()
				assert.Empty(t, out.Workspaces)
			},
		},
		{
			name: "success — with workspaces",
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().GetUserWorkspaces(mock.Anything, userID).Return([]*entity.WorkspaceListItem{
					{
						Workspace:   entity.Workspace{ID: wsID, Name: "Dev Team", OwnerID: userID},
						Role:        entity.WorkspaceRoleAdmin,
						MemberCount: 3,
						BoardCount:  2,
					},
				}, nil)
			},
			checkOut: func(t *testing.T, out *workspace.GetWorkspacesOutput) {
				t.Helper()
				require.Len(t, out.Workspaces, 1)
				assert.Equal(t, wsID, out.Workspaces[0].Workspace.ID)
				assert.Equal(t, "Dev Team", out.Workspaces[0].Workspace.Name)
				assert.Equal(t, entity.WorkspaceRoleAdmin, out.Workspaces[0].Role)
				assert.Equal(t, uint(3), out.Workspaces[0].MemberCount)
				assert.Equal(t, uint(2), out.Workspaces[0].BoardCount)
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

			tt.setupMocks(wsRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo)
			out, err := uc.GetWorkspaces(context.Background(), workspace.GetWorkspacesInput{UserID: userID})

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Nil(t, out)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			if tt.checkOut != nil {
				tt.checkOut(t, out)
			}
		})
	}
}
