package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/middleware"
	"github.com/codewebkhongkho/trello-agent/internal/service"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
	sseManager          *service.SSEManager
}

func NewNotificationHandler(
	notificationService *service.NotificationService,
	sseManager *service.SSEManager,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		sseManager:          sseManager,
	}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unreadOnly := c.Query("unread_only") == "true"

	result, err := h.notificationService.List(c.Request.Context(), userID, page, limit, unreadOnly)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"count": count})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	notificationID := c.Param("id")
	if err := h.notificationService.MarkAsRead(c.Request.Context(), userID, notificationID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Notification marked as read")
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "All notifications marked as read")
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	notificationID := c.Param("id")
	if err := h.notificationService.Delete(c.Request.Context(), userID, notificationID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Notification deleted")
}

func (h *NotificationHandler) Stream(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	clientChan := make(chan service.SSEEvent, 100)
	clientID := cuid.New()

	h.sseManager.Register(userID, clientID, clientChan)

	c.SSEvent("connected", gin.H{"client_id": clientID})
	c.Writer.Flush()

	ctx := c.Request.Context()

	go func() {
		<-ctx.Done()
		h.sseManager.Unregister(userID, clientID)
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-clientChan:
			if !ok {
				return
			}
			c.SSEvent(event.Type, event.Data)
			c.Writer.Flush()

		case <-ticker.C:
			c.SSEvent("ping", gin.H{"timestamp": time.Now().Unix()})
			c.Writer.Flush()
		}
	}
}
