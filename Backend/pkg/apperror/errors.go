package apperror

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Details    any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

func WithDetails(err *AppError, details any) *AppError {
	return &AppError{
		Code:       err.Code,
		Message:    err.Message,
		StatusCode: err.StatusCode,
		Details:    details,
	}
}

func Wrap(err error, appErr *AppError) *AppError {
	return &AppError{
		Code:       appErr.Code,
		Message:    fmt.Sprintf("%s: %v", appErr.Message, err),
		StatusCode: appErr.StatusCode,
	}
}

var (
	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "Bad request",
		StatusCode: http.StatusBadRequest,
	}

	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Unauthorized",
		StatusCode: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Forbidden",
		StatusCode: http.StatusForbidden,
	}

	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "Resource already exists",
		StatusCode: http.StatusConflict,
	}

	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		StatusCode: http.StatusUnprocessableEntity,
	}

	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "Too many requests",
		StatusCode: http.StatusTooManyRequests,
	}

	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrEmailNotVerified = &AppError{
		Code:       "EMAIL_NOT_VERIFIED",
		Message:    "Please verify your email first",
		StatusCode: http.StatusForbidden,
	}

	ErrAccountDisabled = &AppError{
		Code:       "ACCOUNT_DISABLED",
		Message:    "Account has been disabled",
		StatusCode: http.StatusForbidden,
	}

	ErrEmailAlreadyExists = &AppError{
		Code:       "EMAIL_ALREADY_EXISTS",
		Message:    "Email already registered",
		StatusCode: http.StatusConflict,
	}

	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		StatusCode: http.StatusUnauthorized,
	}

	ErrTokenExpired = &AppError{
		Code:       "TOKEN_EXPIRED",
		Message:    "Token has expired",
		StatusCode: http.StatusUnauthorized,
	}

	ErrTokenRevoked = &AppError{
		Code:       "TOKEN_REVOKED",
		Message:    "Token has been revoked",
		StatusCode: http.StatusUnauthorized,
	}

	ErrOTPExpired = &AppError{
		Code:       "OTP_EXPIRED",
		Message:    "OTP has expired",
		StatusCode: http.StatusBadRequest,
	}

	ErrOTPInvalid = &AppError{
		Code:       "OTP_INVALID",
		Message:    "Invalid OTP",
		StatusCode: http.StatusBadRequest,
	}

	ErrOTPMaxAttempts = &AppError{
		Code:       "OTP_MAX_ATTEMPTS",
		Message:    "Too many OTP attempts",
		StatusCode: http.StatusTooManyRequests,
	}

	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "User not found",
		StatusCode: http.StatusNotFound,
	}
)
