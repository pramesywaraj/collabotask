package request

type CreateWorkspaceRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Description *string `json:"description"`
}

type InviteMemberRequest struct {
	Emails []string `json:"emails" binding:"required,min=1,dive,email"`
}
