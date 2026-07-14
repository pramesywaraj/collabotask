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
	"collabotask/internal/usecase/common"
)

func TestTransferOwnership(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	targetID := uuid.New()

	workspaceBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		Visibility:  entity.BoardVisibilityWorkspace,
	}
	privateBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		Visibility:  entity.BoardVisibilityPrivate,
	}

	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}
	plainBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
	targetBoardMember := &entity.BoardMember{BoardID: boardID, UserID: targetID, Role: entity.BoardRoleMember}
	targetAlreadyOwner := &entity.BoardMember{BoardID: boardID, UserID: targetID, Role: entity.BoardRoleOwner}

	adminWsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	plainWsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.TransferOwnershipInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		ToUserID:    targetID,
	}

	tests := []struct {
		name       string
		input      board.TransferOwnershipInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "missing ToUserID → validation error",
			input:      board.TransferOwnershipInput{RequesterID: requesterID, WorkspaceID: workspaceID, BoardID: boardID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			// Cases 9 & 10 (checker-denied): archived, not-found, and PRIVATE-hidden
			// boards all surface from CheckMutateAccess as ErrBoardNotFound; the
			// usecase simply propagates it. Those distinct branches are exercised in
			// board_access_check_test.go — here we only assert propagation, so the
			// three collapse into one case at this layer.
			name:  "CheckMutateAccess denies (archived/not-found/hidden) → ErrBoardNotFound propagated",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			// Case 10 (workspace/board mismatch): board resolves but belongs to a
			// different workspace than the path → 404. Distinct usecase branch.
			name: "workspace mismatch → ErrBoardNotFound",
			input: board.TransferOwnershipInput{
				RequesterID: requesterID,
				WorkspaceID: uuid.New(), // different from board's workspace
				BoardID:     boardID,
				ToUserID:    targetID,
			},
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard, // workspaceBoard.WorkspaceID != input.WorkspaceID
					WorkspaceMember: adminWsMember,
					BoardMember:     ownerBoardMember,
				}, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			// Case 3: non-joined admin on PRIVATE board → 403 BOARD_JOIN_REQUIRED
			name:  "non-joined admin on PRIVATE board → ErrBoardJoinRequired",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardJoinRequired)
			},
			wantErr: domain.ErrBoardJoinRequired,
		},
		{
			// Case 8: plain board member (non-owner, non-admin) → 403
			name:  "plain board member → ErrBoardPermissionDenied",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: plainWsMember,
					BoardMember:     plainBoardMember,
				}, nil)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			// Case 8: ws-member not on a WORKSPACE board → 403 (CheckMutateAccess returns ErrBoardAccessDenied)
			name:  "ws-member not on board (WORKSPACE visibility) → ErrBoardAccessDenied from checker",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			// Case 5: target not a board member → 400
			name:  "target not a board member → ErrTransferTargetNotBoardMember",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: plainWsMember,
					BoardMember:     ownerBoardMember,
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrTransferTargetNotBoardMember,
		},
		{
			name:  "unexpected error fetching target member → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: plainWsMember,
					BoardMember:     ownerBoardMember,
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch target member",
		},
		{
			// Case 6: target already owns it → 200 no-op, TransferOwnership NOT called
			name:  "target already owner → no-op success (TransferOwnership not called)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: plainWsMember,
					BoardMember:     ownerBoardMember,
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(targetAlreadyOwner, nil)
				// TransferOwnership must NOT be called
			},
		},
		{
			// Case 1: owner → board member, target becomes BOARD_OWNER
			name:  "owner transfers to board member → success",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				fromID := requesterID
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: plainWsMember,
					BoardMember:     ownerBoardMember,
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(targetBoardMember, nil)
				d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, targetID).Return(&fromID, nil)
			},
		},
		{
			// Case 2: workspace admin transfers on a WORKSPACE board (no join required)
			name:  "workspace admin transfers on WORKSPACE board → success",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				fromID := uuid.New()
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: adminWsMember,
					BoardMember:     nil, // not joined
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(targetBoardMember, nil)
				d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, targetID).Return(&fromID, nil)
			},
		},
		{
			// Case 4: admin who has joined the private board → success
			name:  "admin joined PRIVATE board → success",
			input: board.TransferOwnershipInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				BoardID:     boardID,
				ToUserID:    targetID,
			},
			setupMocks: func(d boardTestDeps) {
				fromID := uuid.New()
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           privateBoard,
					WorkspaceMember: adminWsMember,
					BoardMember:     ownerBoardMember, // joined
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(targetBoardMember, nil)
				d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, targetID).Return(&fromID, nil)
			},
		},
		{
			// Case 7: orphan board (no owner) + admin appoints target → success, fromUserID nil
			name:  "orphan board + admin appoints target → success with nil fromUserID",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: adminWsMember,
					BoardMember:     nil,
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(targetBoardMember, nil)
				d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, targetID).Return(nil, nil)
			},
		},
		{
			name:  "repo TransferOwnership fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{
					Board:           workspaceBoard,
					WorkspaceMember: plainWsMember,
					BoardMember:     ownerBoardMember,
				}, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, targetID).Return(targetBoardMember, nil)
				d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, targetID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to transfer ownership",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.TransferOwnership(context.Background(), tt.input)

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
