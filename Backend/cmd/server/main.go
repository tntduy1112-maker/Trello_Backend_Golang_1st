package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/codewebkhongkho/trello-agent/internal/config"
	"github.com/codewebkhongkho/trello-agent/internal/handler"
	"github.com/codewebkhongkho/trello-agent/internal/middleware"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/internal/service"
	"github.com/codewebkhongkho/trello-agent/pkg/cache"
	"github.com/codewebkhongkho/trello-agent/pkg/database"
	"github.com/codewebkhongkho/trello-agent/pkg/email"
	"github.com/codewebkhongkho/trello-agent/pkg/jwt"
)

var Version = "dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if cfg.App.Env == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Connected to PostgreSQL")

	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("Connected to Redis")

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	verificationRepo := repository.NewVerificationRepository(db)

	jwtManager := jwt.NewManager(jwt.Config{
		AccessSecret:     cfg.JWT.AccessSecret,
		RefreshSecret:    cfg.JWT.RefreshSecret,
		AccessExpiresIn:  cfg.JWT.AccessExpiresIn,
		RefreshExpiresIn: cfg.JWT.RefreshExpiresIn,
	})

	emailService := email.NewService(email.Config{
		Host: cfg.SMTP.Host,
		Port: cfg.SMTP.Port,
		User: cfg.SMTP.User,
		Pass: cfg.SMTP.Pass,
		From: cfg.SMTP.From,
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	authService := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		VerificationRepo: verificationRepo,
		JWTManager:       jwtManager,
		EmailService:     emailService,
		Cache:            redisClient,
		FrontendURL:      frontendURL,
	})

	authHandler := handler.NewAuthHandler(authService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())
	r.Use(middleware.ErrorHandler())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175", cfg.App.URL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": Version,
		})
	})

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 3,
					Window:      time.Hour,
					KeyPrefix:   "register",
				}),
				authHandler.Register,
			)

			auth.POST("/verify-email",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 10,
					Window:      15 * time.Minute,
					KeyPrefix:   "verify_email",
				}),
				authHandler.VerifyEmail,
			)

			auth.POST("/resend-verification",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 3,
					Window:      time.Hour,
					KeyPrefix:   "resend_verification",
				}),
				authHandler.ResendVerification,
			)

			auth.POST("/login",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 5,
					Window:      15 * time.Minute,
					KeyPrefix:   "login",
				}),
				authHandler.Login,
			)

			auth.POST("/refresh",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 30,
					Window:      15 * time.Minute,
					KeyPrefix:   "refresh",
				}),
				authHandler.Refresh,
			)

			auth.POST("/logout", authHandler.Logout)

			auth.POST("/logout-all",
				middleware.Auth(jwtManager, authService),
				authHandler.LogoutAll,
			)

			auth.POST("/forgot-password",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 5,
					Window:      15 * time.Minute,
					KeyPrefix:   "forgot_password",
				}),
				authHandler.ForgotPassword,
			)

			auth.POST("/reset-password",
				middleware.RateLimit(redisClient, middleware.RateLimitConfig{
					MaxRequests: 5,
					Window:      15 * time.Minute,
					KeyPrefix:   "reset_password",
				}),
				authHandler.ResetPassword,
			)

			auth.GET("/me",
				middleware.Auth(jwtManager, authService),
				authHandler.GetMe,
			)

			auth.PUT("/me",
				middleware.Auth(jwtManager, authService),
				authHandler.UpdateMe,
			)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("")
	}
}
