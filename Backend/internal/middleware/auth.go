package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/service"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UserEmailKey        = "user_email"
	JTIKey              = "jti"
)

func Auth(jwtManager *jwt.Manager, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Check Authorization header first
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) {
			tokenString = strings.TrimPrefix(authHeader, BearerPrefix)
		}

		// Fallback to query parameter (for SSE/EventSource)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			response.ErrorResponse(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				response.ErrorResponse(c, apperror.ErrTokenExpired)
			} else {
				response.ErrorResponse(c, apperror.ErrInvalidToken)
			}
			c.Abort()
			return
		}

		isBlacklisted, err := authService.IsTokenBlacklisted(c.Request.Context(), claims.ID)
		if err != nil {
			response.ErrorResponse(c, apperror.ErrInternal)
			c.Abort()
			return
		}
		if isBlacklisted {
			response.ErrorResponse(c, apperror.ErrTokenRevoked)
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(JTIKey, claims.ID)
		c.Next()
	}
}

func OptionalAuth(jwtManager *jwt.Manager, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		isBlacklisted, _ := authService.IsTokenBlacklisted(c.Request.Context(), claims.ID)
		if isBlacklisted {
			c.Next()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(JTIKey, claims.ID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	return userID.(string)
}

func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return ""
	}
	return email.(string)
}
