package response

import (
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
)

type BoardSummary struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	BackgroundColor string                 `json:"background_color"`
	Visibility      domain.BoardVisibility `json:"visibility"`
	IsClosed        bool                   `json:"is_closed"`
	ListsCount      int                    `json:"lists_count"`
	CardsCount      int                    `json:"cards_count"`
	MyRole          domain.BoardRole       `json:"my_role"`
}

type BoardDetail struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     *string                `json:"description,omitempty"`
	BackgroundColor string                 `json:"background_color"`
	Visibility      domain.BoardVisibility `json:"visibility"`
	IsClosed        bool                   `json:"is_closed"`
	MyRole          domain.BoardRole       `json:"my_role"`
	Organization    *OrgSummary            `json:"organization"`
	Members         []*BoardMemberResponse `json:"members"`
	Lists           []*ListWithCards       `json:"lists"`
	Labels          []*LabelSummary        `json:"labels"`
	CreatedAt       time.Time              `json:"created_at"`
}

type ListWithCards struct {
	ID         string        `json:"id"`
	Title      string        `json:"title"`
	Position   float64       `json:"position"`
	CardsCount int           `json:"cards_count"`
	Cards      []CardSummary `json:"cards"`
}

type OrgSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type BoardMemberResponse struct {
	ID       string           `json:"id"`
	User     *UserSummary     `json:"user"`
	Role     domain.BoardRole `json:"role"`
	JoinedAt time.Time        `json:"joined_at"`
}

type InvitationResponse struct {
	ID           string                  `json:"id"`
	Board        *BoardSummaryForInvite  `json:"board"`
	Inviter      *UserSummary            `json:"inviter"`
	InviteeEmail string                  `json:"invitee_email"`
	Role         domain.BoardRole        `json:"role"`
	Message      *string                 `json:"message,omitempty"`
	Status       domain.InvitationStatus `json:"status"`
	ExpiresAt    time.Time               `json:"expires_at"`
	CreatedAt    time.Time               `json:"created_at"`
	Token        string                  `json:"token,omitempty"`
	InviteURL    string                  `json:"invite_url,omitempty"`
	IsNewUser    bool                    `json:"is_new_user"`
}

type BoardSummaryForInvite struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func ToBoardMemberResponse(member *domain.BoardMember) *BoardMemberResponse {
	if member == nil {
		return nil
	}
	return &BoardMemberResponse{
		ID:       member.ID,
		User:     ToUserSummary(member.User),
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
	}
}

func ToInvitationResponse(inv *domain.BoardInvitation) *InvitationResponse {
	return ToInvitationResponseWithURL(inv, "")
}

func ToInvitationResponseWithURL(inv *domain.BoardInvitation, frontendURL string) *InvitationResponse {
	if inv == nil {
		return nil
	}
	resp := &InvitationResponse{
		ID:           inv.ID,
		InviteeEmail: inv.InviteeEmail,
		Role:         inv.Role,
		Message:      inv.Message,
		Status:       inv.Status,
		ExpiresAt:    inv.ExpiresAt,
		CreatedAt:    inv.CreatedAt,
		Token:        inv.Token,
	}
	if frontendURL != "" && inv.Token != "" {
		resp.InviteURL = frontendURL + "/invite/" + inv.Token
	}
	if inv.Board != nil {
		resp.Board = &BoardSummaryForInvite{
			ID:    inv.Board.ID,
			Title: inv.Board.Title,
		}
	}
	if inv.Inviter != nil {
		resp.Inviter = ToUserSummary(inv.Inviter)
	}
	return resp
}
