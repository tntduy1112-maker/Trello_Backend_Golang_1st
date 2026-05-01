package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/middleware"
	"github.com/codewebkhongkho/trello-agent/internal/service"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, gin.H{
		"user":    user,
		"message": "Registration successful. Please check your email for verification code.",
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req request.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Email verified successfully")
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req request.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.authService.ResendVerification(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Verification code sent if email exists")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	authResp, err := h.authService.Login(c.Request.Context(), &req, deviceInfo, ipAddress)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, authResp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if req.RefreshToken == "" {
		response.ErrorResponse(c, apperror.New("INVALID_REQUEST", "refresh_token is required", http.StatusBadRequest))
		return
	}

	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	authResp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, deviceInfo, ipAddress)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, authResp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var accessToken string
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		accessToken = strings.TrimPrefix(authHeader, "Bearer ")
	}

	var req request.RefreshTokenRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.authService.Logout(c.Request.Context(), accessToken, req.RefreshToken); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Logged out successfully")
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Logged out from all devices")
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.authService.ForgotPassword(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Password reset email sent if account exists")
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.SuccessMessage(c, http.StatusOK, "Password reset successfully")
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.ErrorResponse(c, apperror.ErrUnauthorized)
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			response.ErrorResponse(c, appErr)
			return
		}
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, user)
}
