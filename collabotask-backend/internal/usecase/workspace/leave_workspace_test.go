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
	"collabotask/internal/domain/repository"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/workspace"
)

func TestLeaveWorkspace(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()

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
	ws := &entity.Workspace{
		ID:      workspaceID,
		OwnerID: requesterID,
	}
	wsOtherOwner := &entity.Workspace{
		ID:      workspaceID,
		OwnerID: uuid.New(),
	}

	validInput := workspace.LeaveWorkspaceInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
	}

	tests := []struct {
		name       string
		input      workspace.LeaveWorkspaceInput
		setupMocks func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input (zero RequesterID) → validation error",
			input:      workspace.LeaveWorkspaceInput{WorkspaceID: workspaceID},
			setupMocks: func(_ *mocks.MockWorkspaceRepository, _ *mocks.MockWorkspaceMemberRepository) {},
			wantErrMsg: "validation",
		},
		{
			name:  "requester not a member → ErrMemberNotFound",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, domain.ErrMemberNotFound)
			},
			wantErr: domain.ErrMemberNotFound,
		},
		{
			name:  "owner cannot leave → ErrWorkspaceOwnerCannotLeave",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
			},
			wantErr: domain.ErrWorkspaceOwnerCannotLeave,
		},
		{
			name:  "last admin cannot leave → ErrLastAdminCannotLeave",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(wsOtherOwner, nil)
				wsMemberRepo.EXPECT().CountAdmins(mock.Anything, workspaceID).Return(1, nil)
			},
			wantErr: domain.ErrLastAdminCannotLeave,
		},
		{
			name:  "successful leave as regular member (cascade invoked)",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(wsOtherOwner, nil)
				wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return([]repository.AffectedCard{}, nil)
			},
		},
		{
			name:  "successful leave as admin (not owner, not last admin)",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(wsOtherOwner, nil)
				wsMemberRepo.EXPECT().CountAdmins(mock.Anything, workspaceID).Return(2, nil)
				wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return([]repository.AffectedCard{}, nil)
			},
		},
		{
			name:  "cascade repo error → wrapped error",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(wsOtherOwner, nil)
				wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to leave workspace",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wsRepo := mocks.NewMockWorkspaceRepository(t)
			wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
			userRepo := mocks.NewMockUserRepository(t)

			tt.setupMocks(wsRepo, wsMemberRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo)
			err := uc.LeaveWorkspace(context.Background(), tt.input)

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
