package request

import "github.com/codewebkhongkho/trello-agent/internal/domain"

type CreateBoardRequest struct {
	Title           string                `json:"title" binding:"required,min=1,max=255"`
	Description     *string               `json:"description" binding:"omitempty,max=1000"`
	BackgroundColor string                `json:"background_color" binding:"omitempty,hexcolor"`
	Visibility      domain.BoardVisibility `json:"visibility" binding:"omitempty,oneof=private workspace public"`
}

type UpdateBoardRequest struct {
	Title           *string                `json:"title" binding:"omitempty,min=1,max=255"`
	Description     *string                `json:"description" binding:"omitempty,max=1000"`
	BackgroundColor *string                `json:"background_color" binding:"omitempty,hexcolor"`
	Visibility      *domain.BoardVisibility `json:"visibility" binding:"omitempty,oneof=private workspace public"`
}

type InviteBoardMemberRequest struct {
	Email   string           `json:"email" binding:"required,email"`
	Role    domain.BoardRole `json:"role" binding:"required,oneof=admin member viewer"`
	Message *string          `json:"message" binding:"omitempty,max=500"`
}

type UpdateBoardMemberRoleRequest struct {
	Role domain.BoardRole `json:"role" binding:"required,oneof=admin member viewer"`
}
