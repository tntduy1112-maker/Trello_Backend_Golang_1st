package domain

import (
	"time"
)

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationDeclined InvitationStatus = "declined"
	InvitationExpired  InvitationStatus = "expired"
)

func (s InvitationStatus) IsValid() bool {
	switch s {
	case InvitationPending, InvitationAccepted, InvitationDeclined, InvitationExpired:
		return true
	}
	return false
}

type BoardInvitation struct {
	ID           string           `json:"id"`
	BoardID      string           `json:"board_id"`
	InviterID    string           `json:"inviter_id"`
	InviteeID    *string          `json:"invitee_id,omitempty"`
	InviteeEmail string           `json:"invitee_email"`
	Role         BoardRole        `json:"role"`
	Token        string           `json:"-"`
	Message      *string          `json:"message,omitempty"`
	Status       InvitationStatus `json:"status"`
	ExpiresAt    time.Time        `json:"expires_at"`
	CreatedAt    time.Time        `json:"created_at"`
	RespondedAt  *time.Time       `json:"responded_at,omitempty"`

	Board   *Board `json:"board,omitempty"`
	Inviter *User  `json:"inviter,omitempty"`
}

func (i *BoardInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

func (i *BoardInvitation) IsPending() bool {
	return i.Status == InvitationPending && !i.IsExpired()
}

const InvitationExpiresIn = 3 * 24 * time.Hour
