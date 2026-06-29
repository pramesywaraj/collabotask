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

func TestCreateWorkspace(t *testing.T) {
	ownerID := uuid.New()
	validName := "My Workspace"
	desc := "A description"

	tests := []struct {
		name        string
		input       workspace.CreateWorkspaceInput
		setupMocks  func(wsRepo *mocks.MockWorkspaceRepository)
		wantErrMsg  string
		wantNilDesc bool
		wantDesc    *string
	}{
		{
			name:       "invalid input (empty Name) → validation error",
			input:      workspace.CreateWorkspaceInput{OwnerID: ownerID, Name: ""},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {},
			wantErrMsg: "validation",
		},
		{
			name:       "invalid input (zero OwnerID) → validation error",
			input:      workspace.CreateWorkspaceInput{OwnerID: uuid.Nil, Name: validName},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {},
			wantErrMsg: "validation",
		},
		{
			name:  "CreateWithOwner fails → wrapped error",
			input: workspace.CreateWorkspaceInput{OwnerID: ownerID, Name: validName},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().CreateWithOwner(mock.Anything, mock.Anything, ownerID).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to create workspace",
		},
		{
			name:  "success — nil description stored as nil",
			input: workspace.CreateWorkspaceInput{OwnerID: ownerID, Name: validName, Description: nil},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().CreateWithOwner(mock.Anything, mock.MatchedBy(func(w *entity.Workspace) bool {
					return w.Description == nil
				}), ownerID).Return(nil)
			},
			wantNilDesc: true,
		},
		{
			name:  "success — empty string description stored as nil",
			input: workspace.CreateWorkspaceInput{OwnerID: ownerID, Name: validName, Description: strPtr("")},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().CreateWithOwner(mock.Anything, mock.MatchedBy(func(w *entity.Workspace) bool {
					return w.Description == nil
				}), ownerID).Return(nil)
			},
			wantNilDesc: true,
		},
		{
			name:  "success — with description",
			input: workspace.CreateWorkspaceInput{OwnerID: ownerID, Name: validName, Description: &desc},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository) {
				wsRepo.EXPECT().CreateWithOwner(mock.Anything, mock.MatchedBy(func(w *entity.Workspace) bool {
					return w.Description != nil && *w.Description == desc
				}), ownerID).Return(nil)
			},
			wantDesc: &desc,
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
			out, err := uc.CreateWorkspace(context.Background(), tt.input)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Nil(t, out)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			assert.Equal(t, tt.input.Name, out.Workspace.Name)
			assert.Equal(t, ownerID, out.Workspace.OwnerID)
			if tt.wantNilDesc {
				assert.Nil(t, out.Workspace.Description)
			} else {
				require.NotNil(t, out.Workspace.Description)
				assert.Equal(t, *tt.wantDesc, *out.Workspace.Description)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
