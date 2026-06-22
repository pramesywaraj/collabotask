package response

type failureDocBase struct {
	Success bool `json:"success" example:"false"`
}

type Failure400ValidationDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"400"`
	Message    string `json:"message" example:"Validation failed"`
	Error      *struct {
		Code    string   `json:"code" example:"VALIDATION_ERROR"`
		Message string   `json:"message" example:"Request validation failed"`
		Details []string `json:"details" example:"email must be a valid email address,password must be at least 8 characters"`
	} `json:"error"`
}

type Failure400BadRequestDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"400"`
	Message    string `json:"message" example:"Invalid something id"`
	Error      *struct {
		Code    string `json:"code" example:"VALIDATION_ERROR"`
		Message string `json:"message" example:"Invalid something id"`
	} `json:"error"`
}

type Failure401UnauthorizedDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"401"`
	Message    string `json:"message" example:"Invalid or expired token"`
	Error      *struct {
		Code    string `json:"code" example:"UNAUTHORIZED"`
		Message string `json:"message" example:"Invalid or expired token"`
	} `json:"error"`
}

type Failure401LoginDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"401"`
	Message    string `json:"message" example:"invalid email or password"`
	Error      *struct {
		Code    string `json:"code" example:"UNAUTHORIZED"`
		Message string `json:"message" example:"invalid email or password"`
	} `json:"error"`
}

type Failure403ForbiddenDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"403"`
	Message    string `json:"message" example:"user not in something"`
	Error      *struct {
		Code    string `json:"code" example:"FORBIDDEN"`
		Message string `json:"message" example:"user not in something"`
	} `json:"error"`
}

type Failure404NotFoundDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"404"`
	Message    string `json:"message" example:"something not found"`
	Error      *struct {
		Code    string `json:"code" example:"NOT_FOUND"`
		Message string `json:"message" example:"something not found"`
	} `json:"error"`
}

type Failure409ConflictDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"409"`
	Message    string `json:"message" example:"user already in something"`
	Error      *struct {
		Code    string `json:"code" example:"CONFLICT"`
		Message string `json:"message" example:"user already in something"`
	} `json:"error"`
}

type Failure500InternalDoc struct {
	failureDocBase
	StatusCode int    `json:"status_code" example:"500"`
	Message    string `json:"message" example:"Internal server error"`
	Error      *struct {
		Code    string `json:"code" example:"INTERNAL_ERROR"`
		Message string `json:"message" example:"Internal server error"`
	} `json:"error"`
}

type successDocBase struct {
	Success bool `json:"success" example:"true"`
}

// AUTH
type AuthRegisterSuccessDoc struct {
	successDocBase
	StatusCode int          `json:"status_code" example:"201"`
	Message    string       `json:"message" example:"User registered successfully"`
	Data       AuthResponse `json:"data"`
}

type AuthLoginSuccessDoc struct {
	successDocBase
	StatusCode int          `json:"status_code" example:"200"`
	Message    string       `json:"message" example:"Successfully logged in"`
	Data       AuthResponse `json:"data"`
}

// USER
type UserProfileSuccessDoc struct {
	successDocBase
	StatusCode int          `json:"status_code" example:"200"`
	Message    string       `json:"message" example:"Profile retrieved successfully"`
	Data       UserResponse `json:"data"`
}

// WORKSPACE
type WorkspaceCreateSuccessDoc struct {
	successDocBase
	StatusCode int               `json:"status_code" example:"201"`
	Message    string            `json:"message" example:"Workspace created successfully"`
	Data       WorkspaceResponse `json:"data"`
}

type WorkspaceListSuccessDoc struct {
	successDocBase
	StatusCode int                         `json:"status_code" example:"200"`
	Message    string                      `json:"message" example:"Workspace retrieved successfully"`
	Data       []WorkspaceWithMetaResponse `json:"data"`
}

type WorkspaceDetailSuccessDoc struct {
	successDocBase
	StatusCode int                     `json:"status_code" example:"200"`
	Message    string                  `json:"message" example:"Workspace detail successfully fetched"`
	Data       WorkspaceDetailResponse `json:"data"`
}

type WorkspaceInviteSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Invitations sent"`
	Data       interface{} `json:"data"`
}

type WorkspaceRemoveMemberSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Member removed successfully"`
	Data       interface{} `json:"data"`
}

// BOARD
type BoardCreateSuccessDoc struct {
	successDocBase
	StatusCode int           `json:"status_code" example:"201"`
	Message    string        `json:"message" example:"Board created successfully"`
	Data       BoardResponse `json:"data"`
}

type BoardListInWorkspaceSuccessDoc struct {
	successDocBase
	StatusCode int                     `json:"status_code" example:"200"`
	Message    string                  `json:"message" example:"List boards in workspace fetched successfully"`
	Data       []BoardWithMetaResponse `json:"data"`
}

type BoardDetailSuccessDoc struct {
	successDocBase
	StatusCode int                 `json:"status_code" example:"200"`
	Message    string              `json:"message" example:"Board detail fetched successfully"`
	Data       BoardDetailResponse `json:"data"`
}

type BoardKanbanSuccessDoc struct {
	successDocBase
	StatusCode int                 `json:"status_code" example:"200"`
	Message    string              `json:"message" example:"Board kanban fetched successfully"`
	Data       BoardKanbanResponse `json:"data"`
}

type BoardUpdateSuccessDoc struct {
	successDocBase
	StatusCode int           `json:"status_code" example:"200"`
	Message    string        `json:"message" example:"Board successfully updated"`
	Data       BoardResponse `json:"data"`
}

type BoardArchiveSuccessDoc struct {
	successDocBase
	StatusCode int           `json:"status_code" example:"200"`
	Message    string        `json:"message" example:"Board archived status changed"`
	Data       BoardResponse `json:"data"`
}

type BoardInviteMembersSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Members have been invited to the board"`
	Data       interface{} `json:"data"`
}

type BoardRemoveMemberSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Member has been removed from the board"`
	Data       interface{} `json:"data"`
}

type BoardInviteesListSuccessDoc struct {
	successDocBase
	StatusCode int                    `json:"status_code" example:"200"`
	Message    string                 `json:"message" example:"Workspace invitees for board fetched successfully"`
	Data       []BoardInviteeResponse `json:"data"`
}

type BoardSelfJoinSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Successfully joined the board"`
	Data       interface{} `json:"data"`
}

type BoardLeaveSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Successfully left the board"`
	Data       interface{} `json:"data"`
}

// COLUMN
type ColumnCreateSuccessDoc struct {
	successDocBase
	StatusCode int            `json:"status_code" example:"201"`
	Message    string         `json:"message" example:"Column created successfully"`
	Data       ColumnResponse `json:"data"`
}

type ColumnUpdateSuccessDoc struct {
	successDocBase
	StatusCode int            `json:"status_code" example:"200"`
	Message    string         `json:"message" example:"Column updated successfully"`
	Data       ColumnResponse `json:"data"`
}

type ColumnDeleteSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Column deleted successfully"`
	Data       interface{} `json:"data"`
}

type ColumnPositionSuccessDoc struct {
	successDocBase
	StatusCode int            `json:"status_code" example:"200"`
	Message    string         `json:"message" example:"Column position updated successfully"`
	Data       ColumnResponse `json:"data"`
}

// CARD
type CardCreateSuccessDoc struct {
	successDocBase
	StatusCode int          `json:"status_code" example:"201"`
	Message    string       `json:"message" example:"Card created successfully"`
	Data       CardResponse `json:"data"`
}

type CardUpdateSuccessDoc struct {
	successDocBase
	StatusCode int          `json:"status_code" example:"200"`
	Message    string       `json:"message" example:"Card updated successfully"`
	Data       CardResponse `json:"data"`
}

type CardDeleteSuccessDoc struct {
	successDocBase
	StatusCode int         `json:"status_code" example:"200"`
	Message    string      `json:"message" example:"Card deleted successfully"`
	Data       interface{} `json:"data"`
}

type CardMoveSuccessDoc struct {
	successDocBase
	StatusCode int          `json:"status_code" example:"200"`
	Message    string       `json:"message" example:"Card position updated successfully"`
	Data       CardResponse `json:"data"`
}
