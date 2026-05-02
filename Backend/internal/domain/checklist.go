package domain

import "time"

type Checklist struct {
	ID        string          `json:"id"`
	CardID    string          `json:"card_id"`
	Title     string          `json:"title"`
	Position  float64         `json:"position"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`

	Items    []ChecklistItem   `json:"items,omitempty"`
	Progress *ChecklistProgress `json:"progress,omitempty"`
}

type ChecklistItem struct {
	ID          string     `json:"id"`
	ChecklistID string     `json:"checklist_id"`
	Title       string     `json:"title"`
	Position    float64    `json:"position"`
	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CompletedBy *string    `json:"completed_by,omitempty"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Assignee *User `json:"assignee,omitempty"`
}
