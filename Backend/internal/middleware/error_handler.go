package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			response.ErrorResponse(c, appErr)
			return
		}

		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			details := make([]map[string]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				details = append(details, map[string]string{
					"field":   e.Field(),
					"message": getValidationErrorMessage(e),
				})
			}
			response.ErrorResponse(c, apperror.WithDetails(apperror.ErrValidation, details))
			return
		}

		log.Error().Err(err).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("Unhandled error")

		c.JSON(http.StatusInternalServerError, response.Response{
			Success: false,
			Error: &response.Error{
				Code:    "INTERNAL_ERROR",
				Message: "An unexpected error occurred",
			},
		})
	}
}

func getValidationErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "len":
		return "Invalid length"
	default:
		return "Invalid value"
	}
}
