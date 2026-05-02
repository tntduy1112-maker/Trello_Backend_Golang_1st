package domain

import "time"

type List struct {
	ID         string     `json:"id"`
	BoardID    string     `json:"board_id"`
	Title      string     `json:"title"`
	Position   float64    `json:"position"`
	IsArchived bool       `json:"is_archived"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`

	Cards      []*Card `json:"cards,omitempty"`
	CardsCount int     `json:"cards_count"`
}
