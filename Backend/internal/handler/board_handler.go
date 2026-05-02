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

type BoardHandler struct {
	boardService *service.BoardService
}

func NewBoardHandler(boardService *service.BoardService) *BoardHandler {
	return &BoardHandler{boardService: boardService}
}

func (h *BoardHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	var req request.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	board, err := h.boardService.Create(c.Request.Context(), userID, slug, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, board)
}

func (h *BoardHandler) ListByOrg(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	slug := c.Param("slug")
	includeClosed := c.Query("include_closed") == "true"

	boards, err := h.boardService.ListByOrg(c.Request.Context(), userID, slug, includeClosed)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, boards)
}

func (h *BoardHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	board, err := h.boardService.GetByID(c.Request.Context(), userID, boardID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, board)
}

func (h *BoardHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	var req request.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	board, err := h.boardService.Update(c.Request.Context(), userID, boardID, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, board)
}

func (h *BoardHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	if err := h.boardService.Delete(c.Request.Context(), userID, boardID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Board deleted successfully")
}

func (h *BoardHandler) Close(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	if err := h.boardService.Close(c.Request.Context(), userID, boardID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Board closed successfully")
}

func (h *BoardHandler) Reopen(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	if err := h.boardService.Reopen(c.Request.Context(), userID, boardID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Board reopened successfully")
}

func (h *BoardHandler) ListMembers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	members, err := h.boardService.ListMembers(c.Request.Context(), userID, boardID)
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

func (h *BoardHandler) Invite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	var req request.InviteBoardMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	inv, err := h.boardService.Invite(c.Request.Context(), userID, boardID, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, inv)
}

func (h *BoardHandler) ListInvitations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	invitations, err := h.boardService.ListInvitations(c.Request.Context(), userID, boardID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, invitations)
}

func (h *BoardHandler) RevokeInvitation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	invitationID := c.Param("invitationId")
	if err := h.boardService.RevokeInvitation(c.Request.Context(), userID, invitationID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Invitation revoked")
}

func (h *BoardHandler) UpdateMemberRole(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	targetUserID := c.Param("userId")

	var req request.UpdateBoardMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.boardService.UpdateMemberRole(c.Request.Context(), userID, boardID, targetUserID, req.Role); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Member role updated successfully")
}

func (h *BoardHandler) RemoveMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	targetUserID := c.Param("userId")

	if err := h.boardService.RemoveMember(c.Request.Context(), userID, boardID, targetUserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Member removed successfully")
}

func (h *BoardHandler) Leave(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	if err := h.boardService.Leave(c.Request.Context(), userID, boardID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Left board successfully")
}
