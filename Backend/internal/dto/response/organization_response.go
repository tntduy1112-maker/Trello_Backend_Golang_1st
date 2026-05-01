package response

import (
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
)

type OrganizationSummary struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	LogoURL     *string        `json:"logo_url,omitempty"`
	Role        domain.OrgRole `json:"role"`
	BoardsCount int            `json:"boards_count"`
}

type OrganizationDetail struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	Description  *string        `json:"description,omitempty"`
	LogoURL      *string        `json:"logo_url,omitempty"`
	Owner        *UserSummary   `json:"owner"`
	MembersCount int            `json:"members_count"`
	BoardsCount  int            `json:"boards_count"`
	MyRole       domain.OrgRole `json:"my_role"`
	CreatedAt    time.Time      `json:"created_at"`
}

type UserSummary struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type OrgMemberResponse struct {
	ID       string         `json:"id"`
	User     *UserSummary   `json:"user"`
	Role     domain.OrgRole `json:"role"`
	JoinedAt time.Time      `json:"joined_at"`
}

func ToUserSummary(user *domain.User) *UserSummary {
	if user == nil {
		return nil
	}
	return &UserSummary{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
	}
}

func ToOrgMemberResponse(member *domain.OrganizationMember) *OrgMemberResponse {
	if member == nil {
		return nil
	}
	return &OrgMemberResponse{
		ID:       member.ID,
		User:     ToUserSummary(member.User),
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
	}
}
