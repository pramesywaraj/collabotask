package board_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/board"
)

func TestInviteMemberBoard(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	inviteeID := uuid.New()
	inviteeID2 := uuid.New()

	existingBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		CreatedBy:   uuid.New(),
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.InviteMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserIDs:     []uuid.UUID{inviteeID},
	}

	tests := []struct {
		name       string
		input      board.InviteMemberInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input → validation error",
			input:      board.InviteMemberInput{RequesterID: requesterID, WorkspaceID: workspaceID, BoardID: boardID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "board not found (repo ErrBoardNotFound) → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "unexpected error fetching board → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to verify board existence",
		},
		{
			name:  "board nil/empty → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&entity.Board{}, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "board archived → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				b.IsArchived = true
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "board belongs to different workspace → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				b.WorkspaceID = uuid.New()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "requester not in workspace → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "unexpected error fetching requester's board member → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to check requester permission",
		},
		{
			name:  "not authorized → ErrBoardPermissionDenied",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			name:  "invited user ID not in users map → ErrUserNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{}, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:  "workspaceMemberRepo.IsUserExists returns DB error → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{inviteeID: {ID: inviteeID}}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(false, errors.New("db error"))
			},
			wantErrMsg: "failed to check workspace membership",
		},
		{
			name:  "invitee not in workspace → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{inviteeID: {ID: inviteeID}}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(false, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "boardMemberRepo.IsUserExists returns DB error → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{inviteeID: {ID: inviteeID}}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID).Return(false, errors.New("db error"))
			},
			wantErrMsg: "failed to check board membership",
		},
		{
			name:  "invitee already in board → ErrBoardAlreadyMember",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{inviteeID: {ID: inviteeID}}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID).Return(true, nil)
			},
			wantErr: domain.ErrBoardAlreadyMember,
		},
		{
			name: "only RequesterID in UserIDs (self is skipped, list empty) → ErrBoardNoMembersToInvite",
			input: board.InviteMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				BoardID:     boardID,
				UserIDs:     []uuid.UUID{requesterID},
			},
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{requesterID: {ID: requesterID}}, nil)
			},
			wantErr: domain.ErrBoardNoMembersToInvite,
		},
		{
			name:  "boardMemberRepo.CreateMany fails → error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{inviteeID: {ID: inviteeID}}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID).Return(false, nil)
				d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "success — adds members",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{inviteeID: {ID: inviteeID}}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID).Return(false, nil)
				d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.MatchedBy(func(members []*entity.BoardMember) bool {
					return len(members) == 1 && members[0].UserID == inviteeID && members[0].Role == entity.BoardRoleMember
				})).Return(nil)
			},
		},
		{
			name: "success — adds multiple members (CreateMany receives all)",
			input: board.InviteMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				BoardID:     boardID,
				UserIDs:     []uuid.UUID{inviteeID, inviteeID2},
			},
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{
					inviteeID:  {ID: inviteeID},
					inviteeID2: {ID: inviteeID2},
				}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID).Return(false, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID2).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID2).Return(false, nil)
				d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.MatchedBy(func(members []*entity.BoardMember) bool {
					if len(members) != 2 {
						return false
					}
					got := make(map[uuid.UUID]bool, 2)
					for _, m := range members {
						if m.BoardID != boardID || m.Role != entity.BoardRoleMember {
							return false
						}
						got[m.UserID] = true
					}
					return got[inviteeID] && got[inviteeID2]
				})).Return(nil)
			},
		},
		{
			name: "second invitee already a member → aborts whole batch, CreateMany not called",
			input: board.InviteMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				BoardID:     boardID,
				UserIDs:     []uuid.UUID{inviteeID, inviteeID2},
			},
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{
					inviteeID:  {ID: inviteeID},
					inviteeID2: {ID: inviteeID2},
				}, nil)
				// first invitee is valid and gets queued...
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID).Return(false, nil)
				// ...but the second is already on the board, so the whole call aborts
				// before CreateMany — the first invitee is NOT persisted.
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, inviteeID2).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, inviteeID2).Return(true, nil)
				// No CreateMany expectation: NewMockX(t) fails the test if it is called.
			},
			wantErr: domain.ErrBoardAlreadyMember,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.InviteMember(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}
