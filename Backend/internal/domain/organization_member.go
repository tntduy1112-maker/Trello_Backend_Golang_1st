package domain

import (
	"time"
)

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
)

func (r OrgRole) IsValid() bool {
	switch r {
	case OrgRoleOwner, OrgRoleAdmin, OrgRoleMember:
		return true
	}
	return false
}

func (r OrgRole) Level() int {
	switch r {
	case OrgRoleOwner:
		return 3
	case OrgRoleAdmin:
		return 2
	case OrgRoleMember:
		return 1
	}
	return 0
}

func (r OrgRole) HasPermission(required OrgRole) bool {
	return r.Level() >= required.Level()
}

type OrganizationMember struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           OrgRole   `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`

	User         *User         `json:"user,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
}
