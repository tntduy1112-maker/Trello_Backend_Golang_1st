package domain

import (
	"time"
)

type User struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	FullName         string     `json:"full_name"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	IsVerified       bool       `json:"is_verified"`
	IsActive         bool       `json:"is_active"`
	TokensValidAfter time.Time  `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"-"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
