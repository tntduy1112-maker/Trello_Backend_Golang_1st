package request

import (
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
)

type CreateCardRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

type UpdateCardRequest struct {
	Title       *string              `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string              `json:"description" validate:"omitempty,max=10000"`
	Priority    *domain.CardPriority `json:"priority" validate:"omitempty,oneof=none low medium high"`
	DueDate     *time.Time           `json:"due_date"`
}

type MoveCardRequest struct {
	ListID   string  `json:"list_id" validate:"required"`
	Position float64 `json:"position" validate:"required,gt=0"`
}

type AssignCardRequest struct {
	UserID string `json:"user_id" validate:"required"`
}
