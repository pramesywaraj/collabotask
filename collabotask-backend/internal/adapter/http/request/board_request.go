package request

import "github.com/google/uuid"

type CreateBoardRequest struct {
	Title           string  `json:"title" binding:"required,min=3,max=255"`
	Description     *string `json:"description" binding:"omitempty,max=1000"`
	BackgroundColor *string `json:"background_color" binding:"omitempty,min=4,max=8"`
}

type UpdateBoardRequest struct {
	Title           *string               `json:"title" binding:"omitempty,min=3,max=255"`
	Description     OptionalPatch[string] `json:"description"`
	BackgroundColor *string               `json:"background_color" binding:"omitempty,min=4,max=8"`
}

type InviteMemberBoardRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required,min=1,dive"`
}

type RemoveMemberBoardRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type SetArchivedBoardRequest struct {
	IsArchived *bool `json:"is_archived" binding:"required"`
}
