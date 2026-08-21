package board_test

// Broadcast contract tests for the board use case (Part C / ADR-009 / SRS §5.2).
// Three contract points per mutation: (1) happy path — Broadcast called with the
// right event type and board_id, (2) no-op silence — guards that prevent the write
// also prevent the broadcast, (3) best-effort — Broadcast has no return value.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
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

// --- RemoveMember (Part D) ---

func TestRemoveMemberBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	targetID := uuid.New()
	cardID := uuid.New()
	colID := uuid.New()

	existingBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}

	validInput := board.RemoveMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserID:      targetID,
	}

	wireAccess := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).
			Return(nil, domain.ErrBoardMemberNotFound)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d boardTestDeps)
	}{
		{
			name: "happy path — evict + MEMBER_REMOVED + CARD_UPDATED per cleared card (evict-first ordering)",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, targetID).
					Return([]repository.AffectedCard{{CardID: cardID, ColumnID: colID}}, nil)

				// 1. Evict target with non-empty reason (involuntary → ACCESS_REVOKED sent by adapter)
				d.broadcaster.EXPECT().EvictUser(boardID, targetID, mock.MatchedBy(func(r common.EvictReason) bool {
					return r != common.EvictReasonSilent
				}))
				// 2. MEMBER_REMOVED broadcast to room
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.MemberRemoved)
					return ok && ev.BoardID == boardID && ev.UserID == targetID
				}))
				// 3. CARD_UPDATED for each cleared card (assigned_to: null)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardUpdated)
					return ok && ev.Card.ID == cardID &&
						len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "assigned_to" &&
						ev.Assignee == nil
				}))
			},
		},
		{
			name: "no affected cards — evict + MEMBER_REMOVED, no CARD_UPDATED",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d)
				d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, targetID).
					Return(nil, nil)

				d.broadcaster.EXPECT().EvictUser(boardID, targetID, mock.MatchedBy(func(r common.EvictReason) bool {
					return r != common.EvictReasonSilent
				}))
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					_, ok := e.(common.MemberRemoved)
					return ok
				}))
				// broadcaster.Broadcast not called for CARD_UPDATED
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.RemoveMember(context.Background(), validInput)
			require.NoError(t, err)
		})
	}
}

// --- LeaveBoard cascade-error (Part D review) ---

func TestRemoveMemberCascadeError(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	targetID := uuid.New()

	existingBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}

	validInput := board.RemoveMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserID:      targetID,
	}

	t.Run("cascade error — no EvictUser, no broadcast", func(t *testing.T) {
		d := newDeps(t)
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).
			Return(nil, domain.ErrBoardMemberNotFound)
		d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, targetID).
			Return(nil, errors.New("db error"))
		// broadcaster must NOT be called when the cascade errors (return before broadcast)

		err := d.uc.RemoveMember(context.Background(), validInput)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to remove member from board")
	})
}

// --- LeaveBoard (Part D) ---

func TestLeaveBoardBroadcast(t *testing.T) {
	boardID := uuid.New()
	requesterID := uuid.New()
	cardID := uuid.New()

	existingBoard := &entity.Board{ID: boardID}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}

	validInput := board.LeaveBoardInput{
		RequesterID: requesterID,
		BoardID:     boardID,
	}

	wireBase := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	tests := []struct {
		name       string
		wantErr    bool
		setupMocks func(d boardTestDeps)
	}{
		{
			name: "voluntary leave — silent EvictUser (empty reason, no ACCESS_REVOKED)",
			setupMocks: func(d boardTestDeps) {
				wireBase(d)
				d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, requesterID).Return(nil, nil)
				d.broadcaster.EXPECT().EvictUser(boardID, requesterID, common.EvictReasonSilent)
				// no MEMBER_REMOVED (voluntary); no CARD_UPDATED (no cleared cards)
			},
		},
		{
			name: "voluntary leave with cleared cards — CARD_UPDATED per cleared card",
			setupMocks: func(d boardTestDeps) {
				wireBase(d)
				d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, requesterID).
					Return([]repository.AffectedCard{{CardID: cardID}}, nil)
				d.broadcaster.EXPECT().EvictUser(boardID, requesterID, common.EvictReasonSilent)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardUpdated)
					return ok && ev.Card.ID == cardID &&
						len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "assigned_to" &&
						ev.Assignee == nil
				}))
			},
		},
		{
			name:    "cascade error — no EvictUser, no broadcast",
			wantErr: true,
			setupMocks: func(d boardTestDeps) {
				wireBase(d)
				d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, requesterID).
					Return(nil, errors.New("db error"))
				// broadcaster must NOT be called when the cascade errors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.LeaveBoard(context.Background(), validInput)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// --- UpdateBoard →PRIVATE flip (Part D) ---

func TestUpdateBoardPrivateEviction(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	memberID := uuid.New()
	adminID := uuid.New()
	nonMemberID := uuid.New()

	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	wireAccess := func(d boardTestDeps, b *entity.Board) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	_ = nonMemberID // used indirectly via EvictExcept matcher

	tests := []struct {
		name       string
		setupMocks func(d boardTestDeps)
		input      board.UpdateBoardInput
	}{
		{
			name: "WORKSPACE→PRIVATE flip — BOARD_UPDATED + EvictExcept with non-empty reason",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Board",
					Visibility: entity.BoardVisibilityWorkspace})

				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.BoardUpdated)
					return ok && len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "visibility"
				}))

				// board members: member + admin (requester is also admin in workspace)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{
					{BoardID: boardID, UserID: memberID, Role: entity.BoardRoleMember},
					{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner},
				}, nil)
				d.wsMbrRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{
					{WorkspaceID: workspaceID, UserID: adminID, Role: entity.WorkspaceRoleAdmin},
					{WorkspaceID: workspaceID, UserID: nonMemberID, Role: entity.WorkspaceRoleMember},
				}, nil)

				// EvictExcept: allowed = board members + workspace admins; reason non-empty
				d.broadcaster.EXPECT().EvictExcept(boardID, mock.MatchedBy(func(allowed []uuid.UUID) bool {
					set := make(map[uuid.UUID]struct{})
					for _, id := range allowed {
						set[id] = struct{}{}
					}
					_, hasMember := set[memberID]
					_, hasRequester := set[requesterID]
					_, hasAdmin := set[adminID]
					_, hasNonMember := set[nonMemberID]
					return hasMember && hasRequester && hasAdmin && !hasNonMember
				}), mock.MatchedBy(func(r common.EvictReason) bool { return r != common.EvictReasonSilent }))
			},
			input: board.UpdateBoardInput{
				RequesterID: requesterID,
				BoardID:     boardID,
				Visibility:  strPtr("PRIVATE"),
			},
		},
		{
			name: "PRIVATE→WORKSPACE flip — BOARD_UPDATED broadcast, no EvictExcept",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Board",
					Visibility: entity.BoardVisibilityPrivate})

				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.BoardUpdated)
					return ok && len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "visibility"
				}))
				// EvictExcept NOT called (access only widens)
			},
			input: board.UpdateBoardInput{
				RequesterID: requesterID,
				BoardID:     boardID,
				Visibility:  strPtr("WORKSPACE"),
			},
		},
		{
			name: "title-only update — no EvictExcept",
			setupMocks: func(d boardTestDeps) {
				wireAccess(d, &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Old",
					Visibility: entity.BoardVisibilityWorkspace})

				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.BoardUpdated)
					return ok && len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "title"
				}))
				// EvictExcept NOT called
			},
			input: board.UpdateBoardInput{
				RequesterID: requesterID,
				BoardID:     boardID,
				Title:       strPtr("New Title"),
			},
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
