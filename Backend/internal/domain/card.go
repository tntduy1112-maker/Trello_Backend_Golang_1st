package domain

import "time"

type CardPriority string

const (
	PriorityNone   CardPriority = "none"
	PriorityLow    CardPriority = "low"
	PriorityMedium CardPriority = "medium"
	PriorityHigh   CardPriority = "high"
)

type Card struct {
	ID                string       `json:"id"`
	ListID            string       `json:"list_id"`
	Title             string       `json:"title"`
	Description       *string      `json:"description,omitempty"`
	Position          float64      `json:"position"`
	AssigneeID        *string      `json:"assignee_id,omitempty"`
	Priority          CardPriority `json:"priority"`
	DueDate           *time.Time   `json:"due_date,omitempty"`
	IsCompleted       bool         `json:"is_completed"`
	CompletedAt       *time.Time   `json:"completed_at,omitempty"`
	CoverAttachmentID *string      `json:"cover_attachment_id,omitempty"`
	IsArchived        bool         `json:"is_archived"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	ArchivedAt        *time.Time   `json:"archived_at,omitempty"`
	CreatedBy         string       `json:"created_by"`

	Assignee          *User              `json:"assignee,omitempty"`
	Labels            []*Label           `json:"labels,omitempty"`
	CommentsCount     int                `json:"comments_count"`
	AttachmentsCount  int                `json:"attachments_count"`
	ChecklistProgress *ChecklistProgress `json:"checklist_progress,omitempty"`
}

type ChecklistProgress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

type CardMember struct {
	ID      string
	CardID  string
	UserID  string
	AddedAt time.Time

	User *User
}
