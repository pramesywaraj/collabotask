package request

type CreateColumnRequest struct {
	Title string `json:"title" binding:"required,min=1,max=255"`
}

type UpdateColumnRequest struct {
	Title string `json:"title" binding:"required,min=1,max=255"`
}

type UpdateColumnPosition struct {
	Position int `json:"position" binding:"min=0"`
}
