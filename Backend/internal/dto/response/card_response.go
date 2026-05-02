package response

import (
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
)

type CardSummary struct {
	ID                string                    `json:"id"`
	Title             string                    `json:"title"`
	Position          float64                   `json:"position"`
	Description       *string                   `json:"description"`
	Priority          domain.CardPriority       `json:"priority"`
	DueDate           *time.Time                `json:"due_date"`
	IsCompleted       bool                      `json:"is_completed"`
	Assignee          *UserSummary              `json:"assignee"`
	Labels            []LabelSummary            `json:"labels"`
	CommentsCount     int                       `json:"comments_count"`
	AttachmentsCount  int                       `json:"attachments_count"`
	ChecklistProgress *domain.ChecklistProgress `json:"checklists_progress"`
}

type CardDetail struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	Position    float64             `json:"position"`
	Priority    domain.CardPriority `json:"priority"`
	DueDate     *time.Time          `json:"due_date"`
	IsCompleted bool                `json:"is_completed"`
	CoverURL    *string             `json:"cover_url"`
	List        ListSummary         `json:"list"`
	Assignee    *UserSummary        `json:"assignee"`
	Reporter    *UserSummary        `json:"reporter"`
	Labels      []LabelSummary      `json:"labels"`
	Checklists  []interface{}       `json:"checklists"`
	Comments    []interface{}       `json:"comments"`
	Attachments []interface{}       `json:"attachments"`
	Activity    []interface{}       `json:"activity"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type CardResponse struct {
	ID       string  `json:"id"`
	ListID   string  `json:"list_id"`
	Position float64 `json:"position"`
}
