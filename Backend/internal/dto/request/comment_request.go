package request

type CreateCommentRequest struct {
	CardID   string  `json:"-"`
	Content  string  `json:"content" binding:"required,min=1,max=10000"`
	ParentID *string `json:"parent_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}
