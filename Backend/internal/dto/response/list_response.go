package response

import "time"

type ListResponse struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Position   float64   `json:"position"`
	IsArchived bool      `json:"is_archived"`
	CardsCount int       `json:"cards_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListWithCardsResponse struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Position float64       `json:"position"`
	Cards    []CardSummary `json:"cards"`
}

type ListSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
