package request

import "time"

type CreateChecklistRequest struct {
	CardID string `json:"-"`
	Title  string `json:"title" binding:"required,min=1,max=255"`
}

type UpdateChecklistRequest struct {
	Title *string `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
}

type CreateChecklistItemRequest struct {
	ChecklistID string     `json:"-"`
	Title       string     `json:"title" binding:"required,min=1,max=500"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type UpdateChecklistItemRequest struct {
	Title      *string    `json:"title,omitempty" binding:"omitempty,min=1,max=500"`
	AssigneeID *string    `json:"assignee_id,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
}
