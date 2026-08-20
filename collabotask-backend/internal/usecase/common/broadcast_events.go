package common

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

// Cards

type CardCreated struct {
	Card     *entity.Card
	Assignee *entity.User
}

type CardMoved struct {
	CardID       uuid.UUID
	FromColumnID uuid.UUID
	ToColumnID   uuid.UUID
	Position     float64
}

type CardUpdated struct {
	Card          *entity.Card
	Assignee      *entity.User
	ChangedFields []string
}

type CardDeleted struct {
	CardID   uuid.UUID
	ColumnID uuid.UUID
}

// Columns

type ColumnCreated struct {
	Column *entity.Column
}

type ColumnUpdated struct {
	ColumnID uuid.UUID
	Title    string
}

type ColumnDeleted struct {
	ColumnID uuid.UUID
}

type ColumnMoved struct {
	ColumnID uuid.UUID
	Position float64
}

// Board

type MemberAdded struct {
	BoardID uuid.UUID
	User    *entity.User
	Role    entity.BoardRole
}

type OwnershipTransferred struct {
	BoardID    uuid.UUID
	FromUserID *uuid.UUID
	ToUserID   uuid.UUID
}

type BoardUpdated struct {
	Board         *entity.Board
	ChangedFields []string
}

// BoardArchivedSet encodes both BOARD_ARCHIVED and BOARD_UNARCHIVED depending on Archived.
type BoardArchivedSet struct {
	BoardID  uuid.UUID
	Archived bool
}

func (CardCreated) FrameType() FrameType        { return FrameCardCreated }
func (CardMoved) FrameType() FrameType          { return FrameCardMoved }
func (CardUpdated) FrameType() FrameType        { return FrameCardUpdated }
func (CardDeleted) FrameType() FrameType        { return FrameCardDeleted }
func (ColumnCreated) FrameType() FrameType      { return FrameColumnCreated }
func (ColumnUpdated) FrameType() FrameType      { return FrameColumnUpdated }
func (ColumnDeleted) FrameType() FrameType      { return FrameColumnDeleted }
func (ColumnMoved) FrameType() FrameType        { return FrameColumnMoved }
func (MemberAdded) FrameType() FrameType        { return FrameMemberAdded }
func (OwnershipTransferred) FrameType() FrameType { return FrameOwnershipTransferred }
func (BoardUpdated) FrameType() FrameType       { return FrameBoardUpdated }
func (e BoardArchivedSet) FrameType() FrameType {
	if e.Archived {
		return FrameBoardArchived
	}
	return FrameBoardUnarchived
}
