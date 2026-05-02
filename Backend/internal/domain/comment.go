package domain

import "time"

type Comment struct {
	ID        string     `json:"id"`
	CardID    string     `json:"card_id"`
	AuthorID  string     `json:"author_id"`
	Content   string     `json:"content"`
	ParentID  *string    `json:"parent_id,omitempty"`
	IsEdited  bool       `json:"is_edited"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`

	Author  *User     `json:"author,omitempty"`
	Replies []Comment `json:"replies,omitempty"`
}
