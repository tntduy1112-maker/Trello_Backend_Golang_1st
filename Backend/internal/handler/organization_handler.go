package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/middleware"
	"github.com/codewebkhongkho/trello-agent/internal/service"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
)

type OrganizationHandler struct {
	orgService *service.OrganizationService
}

func NewOrganizationHandler(orgService *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	org, err := h.orgService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, org)
}

func (h *OrganizationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	orgs, err := h.orgService.List(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, orgs)
}

func (h *OrganizationHandler) GetBySlug(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	org, err := h.orgService.GetBySlug(c.Request.Context(), userID, slug)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, org)
}

func (h *OrganizationHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	var req request.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	org, err := h.orgService.Update(c.Request.Context(), userID, slug, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, org)
}

func (h *OrganizationHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	if err := h.orgService.Delete(c.Request.Context(), userID, slug); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Organization deleted successfully")
}

func (h *OrganizationHandler) ListMembers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	members, err := h.orgService.ListMembers(c.Request.Context(), userID, slug)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, members)
}

func (h *OrganizationHandler) InviteMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	var req request.InviteOrgMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	member, err := h.orgService.InviteMember(c.Request.Context(), userID, slug, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, member)
}

func (h *OrganizationHandler) ListBoardMembers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	boardMembers, err := h.orgService.ListBoardMembers(c.Request.Context(), userID, slug)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, boardMembers)
}

func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	targetUserID := c.Param("userId")

	var req request.UpdateOrgMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.orgService.UpdateMemberRole(c.Request.Context(), userID, slug, targetUserID, req.Role); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Member role updated successfully")
}

func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	targetUserID := c.Param("userId")

	if err := h.orgService.RemoveMember(c.Request.Context(), userID, slug, targetUserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Member removed successfully")
}

func (h *OrganizationHandler) Leave(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	if err := h.orgService.Leave(c.Request.Context(), userID, slug); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Left organization successfully")
}

func (h *OrganizationHandler) UpdateBoardMemberRole(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	boardID := c.Param("boardId")
	targetUserID := c.Param("userId")

	var req request.UpdateBoardMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.orgService.UpdateBoardMemberRole(c.Request.Context(), userID, slug, boardID, targetUserID, req.Role); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Board member role updated successfully")
}
