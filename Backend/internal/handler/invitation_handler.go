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

type InvitationHandler struct {
	invitationService *service.InvitationService
}

func NewInvitationHandler(invitationService *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

func (h *InvitationHandler) GetByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.ErrorResponse(c, apperror.New("INVALID_TOKEN", "Token is required", 400))
		return
	}

	inv, err := h.invitationService.GetByToken(c.Request.Context(), token)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, inv)
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	token := c.Param("token")
	if token == "" {
		response.ErrorResponse(c, apperror.New("INVALID_TOKEN", "Token is required", 400))
		return
	}

	if err := h.invitationService.Accept(c.Request.Context(), userID, token); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Successfully joined the board")
}

func (h *InvitationHandler) Decline(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	token := c.Param("token")
	if token == "" {
		response.ErrorResponse(c, apperror.New("INVALID_TOKEN", "Token is required", 400))
		return
	}

	if err := h.invitationService.Decline(c.Request.Context(), userID, token); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Invitation declined")
}

func (h *InvitationHandler) AcceptWithPassword(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.ErrorResponse(c, apperror.New("INVALID_TOKEN", "Token is required", 400))
		return
	}

	var req request.AcceptInvitationWithPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	result, err := h.invitationService.AcceptWithPassword(c.Request.Context(), token, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, result)
}
