package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cache"
)

type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
	KeyPrefix   string
}

func RateLimit(cache *cache.RedisClient, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", config.KeyPrefix, c.ClientIP())

		count, err := cache.Incr(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			_ = cache.Expire(c.Request.Context(), key, config.Window)
		}

		if count > int64(config.MaxRequests) {
			response.ErrorResponse(c, apperror.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitByUser(cache *cache.RedisClient, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:%s:user:%s", config.KeyPrefix, userID)

		count, err := cache.Incr(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			_ = cache.Expire(c.Request.Context(), key, config.Window)
		}

		if count > int64(config.MaxRequests) {
			response.ErrorResponse(c, apperror.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitByEmail(cache *cache.RedisClient, config RateLimitConfig, getEmail func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := getEmail(c)
		if email == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:%s:email:%s", config.KeyPrefix, email)

		count, err := cache.Incr(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			_ = cache.Expire(c.Request.Context(), key, config.Window)
		}

		if count > int64(config.MaxRequests) {
			response.ErrorResponse(c, apperror.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
