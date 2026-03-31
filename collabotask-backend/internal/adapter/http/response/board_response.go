package response

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
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

func BoardDTOToResponse(board dto.BoardDTO) BoardResponse {
	return BoardResponse{
		ID:              board.ID,
		WorkspaceID:     board.WorkspaceID,
		Title:           board.Title,
		Description:     board.Description,
		BackgroundColor: board.BackgroundColor,
		CreatedBy:       board.CreatedBy,
		IsArchived:      board.IsArchived,
		CreatedAt:       board.CreatedAt,
		UpdatedAt:       board.UpdatedAt,
	}
}

func BoardWithMetaDTOToResponse(board dto.BoardWithMetaDTO) BoardWithMetaResponse {
	return BoardWithMetaResponse{
		BoardResponse: BoardDTOToResponse(board.BoardDTO),
		UserRole:      board.UserRole,
		AccessStatus:  board.AccessStatus,
		MemberCount:   board.MemberCount,
	}
}

func BoardMemberDTOToResponse(member dto.BoardMemberDTO) BoardMemberResponse {
	return BoardMemberResponse{
		UserID:    member.UserID,
		Email:     member.Email,
		Name:      member.Name,
		AvatarURL: member.AvatarURL,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	}
}

func BoardDetailDTOToResponse(board dto.BoardDetailDTO) BoardDetailResponse {
	members := make([]BoardMemberResponse, 0, len(board.Members))
	for _, member := range board.Members {
		members = append(members, BoardMemberDTOToResponse(member))
	}

	return BoardDetailResponse{
		BoardResponse: BoardDTOToResponse(board.BoardDTO),
		UserRole:      board.UserRole,
		AccessStatus:  board.AccessStatus,
		Members:       members,
	}
}

func BoardInviteeDTOToResponse(invitee dto.BoardInviteeDTO) BoardInviteeResponse {
	return BoardInviteeResponse{
		UserID:        invitee.UserID,
		Email:         invitee.Email,
		Name:          invitee.Name,
		AvatarURL:     invitee.AvatarURL,
		WorkspaceRole: invitee.WorkspaceRole,
		IsBoardMember: invitee.IsBoardMember,
	}
}

func BoardKanbanToResponse(columns []dto.ColumnWithCardsDTO) BoardKanbanResponse {
	out := make([]ColumnWithCardsResponse, 0, len(columns))
	for _, col := range columns {
		out = append(out, ColumnWithCardsDTOToResponse(col))
	}

	return BoardKanbanResponse{Columns: out}
}
