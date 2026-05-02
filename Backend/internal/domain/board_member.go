package domain

import (
	"time"
)

type BoardRole string

const (
	BoardRoleOwner  BoardRole = "owner"
	BoardRoleAdmin  BoardRole = "admin"
	BoardRoleMember BoardRole = "member"
	BoardRoleViewer BoardRole = "viewer"
)

func (r BoardRole) IsValid() bool {
	switch r {
	case BoardRoleOwner, BoardRoleAdmin, BoardRoleMember, BoardRoleViewer:
		return true
	}
	return false
}

func (r BoardRole) Level() int {
	switch r {
	case BoardRoleOwner:
		return 4
	case BoardRoleAdmin:
		return 3
	case BoardRoleMember:
		return 2
	case BoardRoleViewer:
		return 1
	}
	return 0
}

func (r BoardRole) HasPermission(required BoardRole) bool {
	return r.Level() >= required.Level()
}

func (r BoardRole) CanEdit() bool {
	return r.HasPermission(BoardRoleMember)
}

func (r BoardRole) CanEditLists() bool {
	return r.HasPermission(BoardRoleAdmin)
}

func (r BoardRole) CanManage() bool {
	return r.HasPermission(BoardRoleAdmin)
}

func (r BoardRole) CanInvite() bool {
	return r.HasPermission(BoardRoleAdmin)
}

type BoardMember struct {
	ID       string    `json:"id"`
	BoardID  string    `json:"board_id"`
	UserID   string    `json:"user_id"`
	Role     BoardRole `json:"role"`
	JoinedAt time.Time `json:"joined_at"`

	User  *User  `json:"user,omitempty"`
	Board *Board `json:"board,omitempty"`
}
