package board

import (
	"collabotask/internal/dto"
	"context"

	"github.com/google/uuid"
)

type BoardUseCase interface {
	CreateBoard(ctx context.Context, input CreateBoardInput) (*CreateBoardOutput, error)
	BoardDetail(ctx context.Context, input BoardDetailInput) (*BoardDetailOutput, error)
	ListBoardsInWorkspace(ctx context.Context, input ListBoardsInput) (*ListBoardsOutput, error)
	InviteMember(ctx context.Context, input InviteMemberInput) error
	RemoveMember(ctx context.Context, input RemoveMemberInput) error
	ListWorkspaceInviteesForBoard(ctx context.Context, input ListWorkspaceInviteesForBoardInput) (*ListWorkspaceInviteesForBoardOutput, error)
	LeaveBoard(ctx context.Context, input LeaveBoardInput) error
	SelfJoinBoard(ctx context.Context, input SelfJoinBoardInput) error
	UpdateBoard(ctx context.Context, input UpdateBoardInput) (*UpdateBoardOutput, error)
	SetArchived(ctx context.Context, input SetArchivedInput) (*SetArchivedOutput, error)
}

type CreateBoardInput struct {
	WorkspaceID     uuid.UUID `validate:"required"`
	Title           string    `validate:"required,min=3,max=255"`
	Description     *string   `validate:"omitempty,max=1000"`
	RequesterID     uuid.UUID `validate:"required"`
	BackgroundColor string    `validate:"min=4,max=8"`
}

type CreateBoardOutput struct {
	Board dto.BoardDTO
}

type BoardDetailInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
}

type BoardDetailOutput struct {
	Board dto.BoardDetailDTO
}

type ListBoardsInput struct {
	WorkspaceID uuid.UUID `validate:"required"`
	RequesterID uuid.UUID `validate:"required"`
}

type ListBoardsOutput struct {
	Boards []dto.BoardWithMetaDTO
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

type ListWorkspaceInviteesForBoardInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
}

type ListWorkspaceInviteesForBoardOutput struct {
	Members []dto.BoardInviteeDTO
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

type UpdateBoardInput struct {
	RequesterID     uuid.UUID `validate:"required"`
	BoardID         uuid.UUID `validate:"required"`
	BackgroundColor *string   `validate:"omitempty,min=4,max=8"`
	Description     *string   `validate:"omitempty,max=1000"`
	Title           *string   `validate:"omitempty,min=3,max=255"`
}

type UpdateBoardOutput struct {
	Board dto.BoardDTO
}

type SetArchivedInput struct {
	RequesterID uuid.UUID `validate:"required"`
	BoardID     uuid.UUID `validate:"required"`
	IsArchived  *bool     `validate:"required"`
}

type SetArchivedOutput struct {
	Board dto.BoardDTO
}
