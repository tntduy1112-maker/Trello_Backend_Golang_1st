package domain

import "time"

type Label struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      *string   `json:"name,omitempty"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var DefaultLabelColors = []string{
	"#61bd4f", // Green
	"#f2d600", // Yellow
	"#ff9f1a", // Orange
	"#eb5a46", // Red
	"#c377e0", // Purple
	"#0079bf", // Blue
	"#00c2e0", // Sky
	"#51e898", // Lime
	"#ff78cb", // Pink
	"#344563", // Dark
}

type CardLabel struct {
	ID         string
	CardID     string
	LabelID    string
	AssignedAt time.Time
}
