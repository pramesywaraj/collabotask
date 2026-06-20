package response

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/board"
	"time"

	"github.com/google/uuid"
)

type BoardResponse struct {
	ID              uuid.UUID `json:"id"`
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description"`
	CreatedBy       uuid.UUID `json:"created_by"`
	IsArchived      bool      `json:"is_archived"`
	BackgroundColor string    `json:"background_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type BoardWithMetaResponse struct {
	BoardResponse

	UserRole     *entity.BoardRole        `json:"user_role"`
	AccessStatus entity.BoardAccessStatus `json:"access_status"`
	MemberCount  uint                     `json:"member_count"`
}

type BoardMemberResponse struct {
	UserID    uuid.UUID        `json:"id"`
	Email     string           `json:"email"`
	Name      string           `json:"name"`
	AvatarURL *string          `json:"avatar_url"`
	Role      entity.BoardRole `json:"role"`
	JoinedAt  time.Time        `json:"joined_at"`
}

type BoardDetailResponse struct {
	BoardResponse

	UserRole     *entity.BoardRole        `json:"user_role"`
	AccessStatus entity.BoardAccessStatus `json:"access_status"`
	Members      []BoardMemberResponse    `json:"members"`
}

type BoardInviteeResponse struct {
	UserID        uuid.UUID            `json:"user_id"`
	Email         string               `json:"email"`
	Name          string               `json:"name"`
	AvatarURL     *string              `json:"avatar_url"`
	WorkspaceRole entity.WorkspaceRole `json:"workspace_role"`
	IsBoardMember bool                 `json:"is_board_member"`
}

type BoardInviteeListResponse struct {
	Members []BoardInviteeResponse `json:"members"`
}

type BoardKanbanResponse struct {
	Columns []ColumnWithCardsResponse `json:"columns"`
}

func BoardToResponse(b *entity.Board) BoardResponse {
	return BoardResponse{
		ID:              b.ID,
		WorkspaceID:     b.WorkspaceID,
		Title:           b.Title,
		Description:     b.Description,
		BackgroundColor: b.BackgroundColor,
		CreatedBy:       b.CreatedBy,
		IsArchived:      b.IsArchived,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

func BoardWithMetaToResponse(b board.BoardWithMeta) BoardWithMetaResponse {
	return BoardWithMetaResponse{
		BoardResponse: BoardToResponse(b.Board),
		UserRole:      b.UserRole,
		AccessStatus:  b.AccessStatus,
		MemberCount:   b.MemberCount,
	}
}

func BoardMemberToResponse(member board.BoardMember) BoardMemberResponse {
	return BoardMemberResponse{
		UserID:    member.UserID,
		Email:     member.Email,
		Name:      member.Name,
		AvatarURL: member.AvatarURL,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	}
}

func BoardDetailToResponse(b board.BoardDetail) BoardDetailResponse {
	members := make([]BoardMemberResponse, 0, len(b.Members))
	for _, member := range b.Members {
		members = append(members, BoardMemberToResponse(member))
	}

	return BoardDetailResponse{
		BoardResponse: BoardToResponse(b.Board),
		UserRole:      b.UserRole,
		AccessStatus:  b.AccessStatus,
		Members:       members,
	}
}

func BoardInviteeToResponse(invitee board.BoardInvitee) BoardInviteeResponse {
	return BoardInviteeResponse{
		UserID:        invitee.UserID,
		Email:         invitee.Email,
		Name:          invitee.Name,
		AvatarURL:     invitee.AvatarURL,
		WorkspaceRole: invitee.WorkspaceRole,
		IsBoardMember: invitee.IsBoardMember,
	}
}

func BoardKanbanToResponse(columns []board.ColumnWithCards) BoardKanbanResponse {
	out := make([]ColumnWithCardsResponse, 0, len(columns))
	for _, col := range columns {
		cards := make([]CardResponse, 0, len(col.Cards))
		for _, c := range col.Cards {
			cards = append(cards, CardToResponse(c.Card, c.Assignee))
		}
		out = append(out, ColumnWithCardsResponse{
			ColumnResponse: ColumnToResponse(col.Column),
			Cards:          cards,
		})
	}

	return BoardKanbanResponse{Columns: out}
}
