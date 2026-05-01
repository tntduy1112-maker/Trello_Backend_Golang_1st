package domain

import (
	"time"
)

type VerificationType string

const (
	VerificationTypeEmail         VerificationType = "verify_email"
	VerificationTypePasswordReset VerificationType = "reset_password"
)

type EmailVerification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Token     string           `json:"-"`
	Type      VerificationType `json:"type"`
	ExpiresAt time.Time        `json:"expires_at"`
	UsedAt    *time.Time       `json:"used_at,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.ExpiresAt)
}

func (ev *EmailVerification) IsUsed() bool {
	return ev.UsedAt != nil
}

func (ev *EmailVerification) IsValid() bool {
	return !ev.IsExpired() && !ev.IsUsed()
}
