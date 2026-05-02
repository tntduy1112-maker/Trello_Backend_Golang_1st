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

type ListHandler struct {
	listService *service.ListService
}

func NewListHandler(listService *service.ListService) *ListHandler {
	return &ListHandler{listService: listService}
}

func (h *ListHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	boardID := c.Param("id")
	var req request.CreateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	list, err := h.listService.Create(c.Request.Context(), userID, boardID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, list)
}

func (h *ListHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	listID := c.Param("id")
	var req request.UpdateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	list, err := h.listService.Update(c.Request.Context(), userID, listID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, list)
}

func (h *ListHandler) Archive(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	listID := c.Param("id")
	if err := h.listService.Archive(c.Request.Context(), userID, listID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "List archived successfully")
}

func (h *ListHandler) Restore(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	listID := c.Param("id")
	if err := h.listService.Restore(c.Request.Context(), userID, listID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "List restored successfully")
}

func (h *ListHandler) Move(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	listID := c.Param("id")
	var req request.MoveListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	list, err := h.listService.Move(c.Request.Context(), userID, listID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, list)
}

func (h *ListHandler) Copy(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	listID := c.Param("id")
	var req request.CopyListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	list, err := h.listService.Copy(c.Request.Context(), userID, listID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, list)
}
