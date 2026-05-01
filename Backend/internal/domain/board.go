package domain

import (
	"time"
)

type BoardVisibility string

const (
	VisibilityPrivate   BoardVisibility = "private"
	VisibilityWorkspace BoardVisibility = "workspace"
	VisibilityPublic    BoardVisibility = "public"
)

func (v BoardVisibility) IsValid() bool {
	switch v {
	case VisibilityPrivate, VisibilityWorkspace, VisibilityPublic:
		return true
	}
	return false
}

type Board struct {
	ID              string          `json:"id"`
	OrganizationID  string          `json:"organization_id"`
	Title           string          `json:"title"`
	Description     *string         `json:"description,omitempty"`
	BackgroundColor string          `json:"background_color"`
	BackgroundImage *string         `json:"background_image,omitempty"`
	Visibility      BoardVisibility `json:"visibility"`
	IsClosed        bool            `json:"is_closed"`
	OwnerID         string          `json:"owner_id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	ClosedAt        *time.Time      `json:"closed_at,omitempty"`
	DeletedAt       *time.Time      `json:"-"`

	Organization *Organization `json:"organization,omitempty"`
}

func (b *Board) IsDeleted() bool {
	return b.DeletedAt != nil
}

const DefaultBackgroundColor = "#0079bf"
