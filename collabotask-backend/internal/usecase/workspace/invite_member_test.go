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

func TestInviteMember(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()
	userAID := uuid.New()
	userBID := uuid.New()
	userCID := uuid.New()

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

	validInput := workspace.InviteMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		Emails:      []string{"a@example.com"},
	}

	tests := []struct {
		name        string
		input       workspace.InviteMemberInput
		setupMocks  func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository)
		wantErr     error
		wantErrMsg  string
		wantResults []workspace.InviteResult
	}{
		{
			name:  "invalid input (empty Emails) → validation error",
			input: workspace.InviteMemberInput{RequesterID: requesterID, WorkspaceID: workspaceID, Emails: []string{}},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
			},
			wantErrMsg: "validation",
		},
		{
			name:  "GetByWorkspaceAndUser returns DB error → ErrNotWorkspaceAdmin",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErr: domain.ErrNotWorkspaceAdmin,
		},
		{
			name:  "requester found but not admin (IsAdmin() = false) → ErrNotWorkspaceAdmin",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
			},
			wantErr: domain.ErrNotWorkspaceAdmin,
		},
		{
			name:  "email not found → per-email result NOT_FOUND",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(nil, nil)
			},
			wantResults: []workspace.InviteResult{
				{Email: "a@example.com", Success: false, ErrorCode: "NOT_FOUND"},
			},
		},
		{
			name:  "user already in workspace → per-email result CONFLICT",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(&entity.User{ID: userAID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
			},
			wantResults: []workspace.InviteResult{
				{Email: "a@example.com", Success: false, ErrorCode: "CONFLICT"},
			},
		},
		{
			name:  "IsUserExists returns DB error → top-level wrapped error (aborts batch)",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(&entity.User{ID: userAID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(false, errors.New("db error"))
			},
			wantErrMsg: "failed to check member existence",
		},
		{
			name:  "workspaceMemberRepo.Create fails → top-level wrapped error (aborts batch)",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(&entity.User{ID: userAID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(false, nil)
				wsMemberRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to add member to workspace",
		},
		{
			name:  "success — single email added",
			input: validInput,
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(&entity.User{ID: userAID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(false, nil)
				wsMemberRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(m *entity.WorkspaceMember) bool {
					return m.WorkspaceID == workspaceID && m.UserID == userAID && m.Role == entity.WorkspaceRoleMember
				})).Return(nil)
			},
			wantResults: []workspace.InviteResult{
				{Email: "a@example.com", Success: true, ErrorCode: ""},
			},
		},
		{
			name: "email normalization — mixed case lowercased before GetByEmail and reflected in result Email field",
			input: workspace.InviteMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				Emails:      []string{"Alice@Example.COM"},
			},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "alice@example.com").Return(&entity.User{ID: userAID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(false, nil)
				wsMemberRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(m *entity.WorkspaceMember) bool {
					return m.UserID == userAID
				})).Return(nil)
			},
			wantResults: []workspace.InviteResult{
				{Email: "alice@example.com", Success: true, ErrorCode: ""},
			},
		},
		{
			name: "success — mixed batch (NOT_FOUND + CONFLICT + success) → all three entries in order",
			input: workspace.InviteMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				Emails:      []string{"a@example.com", "b@example.com", "c@example.com"},
			},
			setupMocks: func(wsRepo *mocks.MockWorkspaceRepository, wsMemberRepo *mocks.MockWorkspaceMemberRepository, userRepo *mocks.MockUserRepository) {
				wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				// a@: not found
				userRepo.EXPECT().GetByEmail(mock.Anything, "a@example.com").Return(nil, nil)
				// b@: found but already a member
				userRepo.EXPECT().GetByEmail(mock.Anything, "b@example.com").Return(&entity.User{ID: userBID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userBID).Return(true, nil)
				// c@: success
				userRepo.EXPECT().GetByEmail(mock.Anything, "c@example.com").Return(&entity.User{ID: userCID}, nil)
				wsMemberRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userCID).Return(false, nil)
				wsMemberRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(m *entity.WorkspaceMember) bool {
					return m.UserID == userCID
				})).Return(nil)
			},
			wantResults: []workspace.InviteResult{
				{Email: "a@example.com", Success: false, ErrorCode: "NOT_FOUND"},
				{Email: "b@example.com", Success: false, ErrorCode: "CONFLICT"},
				{Email: "c@example.com", Success: true, ErrorCode: ""},
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

			tt.setupMocks(wsRepo, wsMemberRepo, userRepo)

			uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo)
			out, err := uc.InviteMember(context.Background(), tt.input)

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
			assert.Equal(t, tt.wantResults, out.Results)
		})
	}
}
