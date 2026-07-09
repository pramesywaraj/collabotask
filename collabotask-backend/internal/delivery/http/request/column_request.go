package request

type CreateColumnRequest struct {
	Title string `json:"title" binding:"required,min=1,max=255"`
}

type UpdateColumnRequest struct {
	Title string `json:"title" binding:"required,min=1,max=255"`
}

type UpdateColumnPosition struct {
	// Pointer enforces presence (omitted or null → 400). Any finite float64 value
	// is valid; the JSON decoder rejects NaN/±Inf before binding runs.
	Position *float64 `json:"position" binding:"required"`
}
