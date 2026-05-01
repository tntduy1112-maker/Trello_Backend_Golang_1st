package request

import "github.com/codewebkhongkho/trello-agent/internal/domain"

type CreateOrganizationRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

type UpdateOrganizationRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

type InviteOrgMemberRequest struct {
	Email string         `json:"email" binding:"required,email"`
	Role  domain.OrgRole `json:"role" binding:"required,oneof=admin member"`
}

type UpdateOrgMemberRoleRequest struct {
	Role domain.OrgRole `json:"role" binding:"required,oneof=admin member"`
}
