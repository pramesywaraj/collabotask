package workspace_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/workspace"
)

func TestGetWorkspaceDetail(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()
	memberUserID := uuid.New()

	aWorkspace := &entity.Workspace{
		ID:      workspaceID,
		Name:    "Dev Team",
		OwnerID: requesterID,
	}

	requesterWsMember := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleAdmin,
	}
	otherWsMember := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      memberUserID,
		Role:        entity.WorkspaceRoleMember,
		JoinedAt:    time.Now(),
	}

	validInput := workspace.GetWorkspaceDetailInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
	}

	tests := []struct {
		name       string
		input      workspace.GetWorkspaceDetailInput
		setupMocks func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *workspace.GetWorkspaceDetailOutput)
	}{
		{
			name:  "invalid input (zero RequesterID) → validation error",
			input: workspace.GetWorkspaceDetailInput{RequesterID: uuid.Nil, WorkspaceID: workspaceID},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
			},
			wantErrMsg: "validation",
		},
		{
			name:  "requester not in workspace → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "GetByID returns ErrWorkspaceNotFound → ErrWorkspaceNotFound",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(nil, domain.ErrWorkspaceNotFound)
			},
			wantErr: domain.ErrWorkspaceNotFound,
		},
		{
			name:  "GetByID returns (nil, nil) → ErrWorkspaceNotFound",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(nil, nil)
			},
			wantErr: domain.ErrWorkspaceNotFound,
		},
		{
			name:  "GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(nil, errors.New("db timeout"))
			},
			wantErrMsg: "failed to fetch workspace",
		},
		{
			name:  "GetMembersByWorkspace fails → wrapped error",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(aWorkspace, nil)
				wsMemberRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch workspace members",
		},
		{
			name:  "userRepo.GetByIds fails → wrapped error",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(aWorkspace, nil)
				wsMemberRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{otherWsMember}, nil)
				userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch member details",
		},
		{
			name:  "user missing from map for a member → ErrUserNotFound",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(aWorkspace, nil)
				wsMemberRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{otherWsMember}, nil)
				// return an empty map — memberUserID is not in it
				userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{}, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:  "success — returns workspace + member list with roles",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(requesterWsMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(aWorkspace, nil)
				wsMemberRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{otherWsMember}, nil)
				userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{
					memberUserID: {ID: memberUserID, Email: "other@example.com", Name: "Other User"},
				}, nil)
			},
			checkOut: func(t *testing.T, out *workspace.GetWorkspaceDetailOutput) {
				t.Helper()
				assert.Equal(t, workspaceID, out.Workspace.Workspace.ID)
				assert.Equal(t, entity.WorkspaceRoleAdmin, out.Workspace.UserRole)
				require.Len(t, out.Workspace.Members, 1)
				assert.Equal(t, memberUserID, out.Workspace.Members[0].UserID)
				assert.Equal(t, "other@example.com", out.Workspace.Members[0].Email)
				assert.Equal(t, entity.WorkspaceRoleMember, out.Workspace.Members[0].Role)
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

			tt.setupMocks(wsRepo, wsMemberRepo, userRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, mocks.NewMockBoardRepository(t), userRepo, activityRepo, mocks.NewMockBroadcaster(t))
			out, err := uc.GetWorkspaceDetail(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, out)
				return
			}
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
