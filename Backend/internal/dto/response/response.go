package response

import (
	"github.com/gin-gonic/gin"

	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
)

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

func SuccessMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context, err *apperror.AppError) {
	c.JSON(err.StatusCode, Response{
		Success: false,
		Error: &Error{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
	})
}

func Paginated(c *gin.Context, statusCode int, data any, pagination *Pagination) {
	c.JSON(statusCode, PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	})
}

func SuccessPaginated(c *gin.Context, statusCode int, data any, page, limit, total int) {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}
	c.JSON(statusCode, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: &Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
