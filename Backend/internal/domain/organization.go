package domain

import (
	"time"
)

type Organization struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	LogoURL     *string    `json:"logo_url,omitempty"`
	OwnerID     string     `json:"owner_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}

func (o *Organization) IsDeleted() bool {
	return o.DeletedAt != nil
}
