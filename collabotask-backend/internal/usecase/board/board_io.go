package board

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// Result types for the board use cases. Plain boards are returned as
// *entity.Board directly; the types below carry genuine transformations
// (joined/aggregated data) and therefore live next to the use case that
// produces them, not in a global dto package.

type BoardWithMeta struct {
	Board        *entity.Board
	UserRole     *entity.BoardRole
	AccessStatus entity.BoardAccessStatus
	MemberCount  uint
	CardCount    uint
}

type BoardMember struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	AvatarURL *string
	Role      entity.BoardRole
	JoinedAt  time.Time
}

type BoardDetail struct {
	Board        *entity.Board
	UserRole     *entity.BoardRole
	AccessStatus entity.BoardAccessStatus
	Members      []BoardMember
}

type BoardInvitee struct {
	UserID        uuid.UUID
	Email         string
	Name          string
	AvatarURL     *string
	WorkspaceRole entity.WorkspaceRole
	IsBoardMember bool
}

type CardWithAssignee struct {
	Card     *entity.Card
	Assignee *entity.User
}

type ColumnWithCards struct {
	Column *entity.Column
	Cards  []CardWithAssignee
}

type CreateBoardInput struct {
	WorkspaceID     uuid.UUID `validate:"required"`
	Title           string    `validate:"required,min=3,max=255"`
	Description     *string   `validate:"omitempty,max=1000"`
	RequesterID     uuid.UUID `validate:"required"`
	BackgroundColor *string   `validate:"omitempty,min=4,max=8"`
	Visibility      *string   `validate:"omitempty,oneof=PRIVATE WORKSPACE"`
}

type CreateBoardOutput struct {
	Board *entity.Board
}

type GetBoardDetailInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
}

type GetBoardDetailOutput struct {
	Board BoardDetail
}

type GetBoardsInput struct {
	WorkspaceID uuid.UUID `validate:"required"`
	RequesterID uuid.UUID `validate:"required"`
}

type GetBoardsOutput struct {
	Boards []BoardWithMeta
}

type InviteMemberInput struct {
	RequesterID uuid.UUID   `validate:"required"`
	WorkspaceID uuid.UUID   `validate:"required"`
	BoardID     uuid.UUID   `validate:"required"`
	UserIDs     []uuid.UUID `validate:"required,min=1,dive"`
}

type RemoveMemberInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
	UserID      uuid.UUID `validate:"required"`
}

type GetWorkspaceInviteesForBoardInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
}

type GetWorkspaceInviteesForBoardOutput struct {
	Members []BoardInvitee
}

type LeaveBoardInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
}

type SelfJoinBoardInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
}

type SelfJoinBoardOutput struct {
	// Joined is true when the requester was newly added, false when they were
	// already a member (idempotent re-join). It also gates the future activity
	// log / WebSocket broadcast so a no-op re-join stays silent.
	Joined bool
}

type UpdateBoardInput struct {
	RequesterID        uuid.UUID `validate:"required"`
	BoardID            uuid.UUID `validate:"required"`
	BackgroundColor    *string   `validate:"omitempty,min=4,max=8"`
	Description        *string   `validate:"omitempty,max=1000"`
	DescriptionPresent bool
	Title              *string `validate:"omitempty,min=3,max=255"`
	Visibility         *string `validate:"omitempty,oneof=PRIVATE WORKSPACE"`
}

type UpdateBoardOutput struct {
	Board *entity.Board
}

type SetArchivedInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
	IsArchived  *bool     `validate:"required"`
}

type SetArchivedOutput struct {
	Board *entity.Board
}

type GetBoardKanbanInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
}

type GetBoardKanbanOutput struct {
	Columns []ColumnWithCards
}
