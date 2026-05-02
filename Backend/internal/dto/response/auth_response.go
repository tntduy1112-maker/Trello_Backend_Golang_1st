package response

import (
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
)

type UserResponse struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

func ToUserResponse(user *domain.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FullName:   user.FullName,
		AvatarURL:  user.AvatarURL,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
	}
}

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int64         `json:"expires_in"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type AcceptInvitationResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int64         `json:"expires_in"`
	BoardID      string        `json:"board_id"`
}
