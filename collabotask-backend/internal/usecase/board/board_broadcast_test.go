package board_test

// Broadcast contract tests for the board use case (Part C / ADR-009 / SRS §5.2).
// Three contract points per mutation: (1) happy path — Broadcast called with the
// right event type and board_id, (2) no-op silence — guards that prevent the write
// also prevent the broadcast, (3) best-effort — Broadcast has no return value.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/board"
	"collabotask/internal/usecase/common"
)

// --- UpdateBoard ---

func TestUpdateBoardBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	// wireAccess wires the read + permission + write path shared by every case.
	wireAccess := func(d boardTestDeps, b *entity.Board) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d boardTestDeps)
		input      board.UpdateBoardInput
	}{
		{
			name: "happy path — BOARD_UPDATED broadcast with changed_fields=[title]",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Old Title"})
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.BoardUpdated)
					return ok &&
						ev.Board.ID == boardID &&
						len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "title"
				}))
			},
			input: board.UpdateBoardInput{RequesterID: requesterID, BoardID: boardID, Title: strPtr("New Title")},
		},
		{
			name: "no-op silence — title unchanged → Broadcast NOT called",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Same Title"})
				// broadcaster.EXPECT not set → any call fails the test
			},
			input: board.UpdateBoardInput{RequesterID: requesterID, BoardID: boardID, Title: strPtr("Same Title")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			_, err := d.uc.UpdateBoard(context.Background(), tt.input)
			require.NoError(t, err)
		})
	}
}

// --- SetArchived ---

func TestSetArchivedBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	// wireRead wires the read + permission path shared by every case. The write
	// (SetArchived) only fires on an actual state change, so it lives per-case.
	wireRead := func(d boardTestDeps, b *entity.Board) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
	}
	wireWrite := func(d boardTestDeps) {
		d.boardRepo.EXPECT().SetArchived(mock.Anything, boardID, mock.AnythingOfType("bool")).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d boardTestDeps)
		archive    bool
	}{
		{
			name: "happy path — BOARD_ARCHIVED broadcast when archiving",
			setupMocks: func(d boardTestDeps) {
				wireRead(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: false})
				wireWrite(d)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.BoardArchivedSet)
					return ok && ev.BoardID == boardID && ev.Archived == true
				}))
			},
			archive: true,
		},
		{
			name: "happy path — BOARD_UNARCHIVED broadcast when unarchiving",
			setupMocks: func(d boardTestDeps) {
				wireRead(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: true})
				wireWrite(d)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.BoardArchivedSet)
					return ok && ev.BoardID == boardID && ev.Archived == false
				}))
			},
			archive: false,
		},
		{
			name: "no-op silence — already archived → Broadcast NOT called",
			setupMocks: func(d boardTestDeps) {
				wireRead(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: true})
				// no write, no broadcast
			},
			archive: true,
		},
		{
			name: "no-op silence — already unarchived → Broadcast NOT called",
			setupMocks: func(d boardTestDeps) {
				wireRead(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: false})
				// no write, no broadcast
			},
			archive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
				RequesterID: requesterID,
				BoardID:     boardID,
				IsArchived:  boolPtr(tt.archive),
			})
			require.NoError(t, err)
		})
	}
}

// --- TransferOwnership ---

func TestTransferOwnershipBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	toUserID := uuid.New()
	fromUserID := uuid.New()

	b := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}
	targetMember := &entity.BoardMember{BoardID: boardID, UserID: toUserID, Role: entity.BoardRoleMember}
	alreadyOwnerTarget := &entity.BoardMember{BoardID: boardID, UserID: toUserID, Role: entity.BoardRoleOwner}

	input := board.TransferOwnershipInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		ToUserID:    toUserID,
	}

	wireAccess := func(d boardTestDeps) {
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{
				Board:           b,
				WorkspaceMember: wsMember,
				BoardMember:     ownerBoardMember,
			}, nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d boardTestDeps)
	}{
		{
			name: "happy path — OWNERSHIP_TRANSFERRED broadcast with from/to user IDs",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, toUserID).Return(targetMember, nil)
				d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, toUserID).Return(&fromUserID, nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.OwnershipTransferred)
					return ok &&
						ev.BoardID == boardID &&
						ev.ToUserID == toUserID &&
						ev.FromUserID != nil && *ev.FromUserID == fromUserID
				}))
			},
		},
		{
			name: "no-op — target already owner → Broadcast NOT called",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, toUserID).Return(alreadyOwnerTarget, nil)
				// broadcaster.EXPECT not set
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.TransferOwnership(context.Background(), input)
			require.NoError(t, err)
		})
	}
}

// --- InviteMember ---

func TestInviteMemberBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	userAID := uuid.New()
	userBID := uuid.New()
	joinedAt := time.Now().UTC().Truncate(time.Second)

	b := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	wireAccess := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(ownerBoardMember, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}
	// stampJoinedAt mimics CreateMany's RETURNING joined_at: the DB timestamp
	// flows back onto the member structs the broadcast then reads.
	stampJoinedAt := func(_ context.Context, members []*entity.BoardMember) {
		for _, m := range members {
			m.JoinedAt = joinedAt
		}
	}

	tests := []struct {
		name       string
		setupMocks func(d boardTestDeps)
		userIDs    []uuid.UUID
		wantErr    error
	}{
		{
			name: "happy path — MEMBER_ADDED broadcast once, carrying user + role + joined_at",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				userA := &entity.User{ID: userAID, Name: "Alice"}
				d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID}).
					Return(map[uuid.UUID]*entity.User{userAID: userA}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
				d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Run(stampJoinedAt).Return(nil)

				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.MemberAdded)
					return ok &&
						ev.BoardID == boardID &&
						ev.User.ID == userAID &&
						ev.Role == entity.BoardRoleMember &&
						ev.JoinedAt.Equal(joinedAt)
				}))
			},
			userIDs: []uuid.UUID{userAID},
		},
		{
			name: "happy path — MEMBER_ADDED broadcast once per each of multiple invitees",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				userA := &entity.User{ID: userAID, Name: "Alice"}
				userB := &entity.User{ID: userBID, Name: "Bob"}
				d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID, userBID}).
					Return(map[uuid.UUID]*entity.User{userAID: userA, userBID: userB}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userBID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userBID).Return(false, nil)
				d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Run(stampJoinedAt).Return(nil)

				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.MemberAdded)
					return ok && ev.User.ID == userAID
				})).Once()
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.MemberAdded)
					return ok && ev.User.ID == userBID
				})).Once()
			},
			userIDs: []uuid.UUID{userAID, userBID},
		},
		{
			name: "duplicate invitee already a board member → error, Broadcast NOT called",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				userA := &entity.User{ID: userAID, Name: "Alice"}
				d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID}).
					Return(map[uuid.UUID]*entity.User{userAID: userA}, nil)
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
				d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(true, nil)
				// already a member → CreateMany + broadcaster never reached
			},
			userIDs: []uuid.UUID{userAID},
			wantErr: domain.ErrBoardAlreadyMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.InviteMember(context.Background(), board.InviteMemberInput{
				RequesterID: requesterID,
				WorkspaceID: workspaceID,
				BoardID:     boardID,
				UserIDs:     tt.userIDs,
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
