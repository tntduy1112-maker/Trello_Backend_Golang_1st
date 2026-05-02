package main

import (
	"context"
	_ "embed"
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
	"github.com/codewebkhongkho/trello-agent/pkg/storage"
)

//go:embed swagger.yaml
var swaggerYAML []byte

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
	orgRepo := repository.NewOrganizationRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	invRepo := repository.NewInvitationRepository(db)
	listRepo := repository.NewListRepository(db)
	cardRepo := repository.NewCardRepository(db)
	labelRepo := repository.NewLabelRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	checklistRepo := repository.NewChecklistRepository(db)
	attachmentRepo := repository.NewAttachmentRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// Initialize MinIO storage
	storageService, err := storage.NewMinioStorage(storage.MinioConfig{
		Endpoint:   cfg.MinIO.Endpoint,
		AccessKey:  cfg.MinIO.AccessKey,
		SecretKey:  cfg.MinIO.SecretKey,
		Bucket:     cfg.MinIO.Bucket,
		UseSSL:     cfg.MinIO.UseSSL,
		PublicHost: cfg.MinIO.PublicHost,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MinIO")
	}
	log.Info().Msg("Connected to MinIO")

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

	orgService := service.NewOrganizationService(orgRepo, userRepo, boardRepo)

	boardService := service.NewBoardService(service.BoardServiceConfig{
		BoardRepo:   boardRepo,
		OrgRepo:     orgRepo,
		InvRepo:     invRepo,
		UserRepo:    userRepo,
		ListRepo:    listRepo,
		LabelRepo:   labelRepo,
		EmailSvc:    emailService,
		FrontendURL: frontendURL,
	})

	sseManager := service.NewSSEManager()
	notificationService := service.NewNotificationService(notificationRepo, sseManager)

	activityService := service.NewActivityService(activityRepo, cardRepo, boardRepo)
	listService := service.NewListService(listRepo, boardRepo, cardRepo)
	cardService := service.NewCardService(cardRepo, listRepo, boardRepo, labelRepo, userRepo, notificationService, activityService)
	labelService := service.NewLabelService(labelRepo, boardRepo, cardRepo)
	commentService := service.NewCommentService(commentRepo, cardRepo, boardRepo, activityRepo, userRepo, notificationService)
	checklistService := service.NewChecklistService(checklistRepo, cardRepo, boardRepo, activityRepo)
	attachmentService := service.NewAttachmentService(attachmentRepo, cardRepo, boardRepo, activityRepo, storageService)
	invitationService := service.NewInvitationService(invRepo, boardRepo, userRepo, orgRepo, jwtManager)

	authHandler := handler.NewAuthHandler(authService)
	orgHandler := handler.NewOrganizationHandler(orgService)
	boardHandler := handler.NewBoardHandler(boardService)
	listHandler := handler.NewListHandler(listService)
	cardHandler := handler.NewCardHandler(cardService)
	labelHandler := handler.NewLabelHandler(labelService)
	commentHandler := handler.NewCommentHandler(commentService)
	checklistHandler := handler.NewChecklistHandler(checklistService)
	attachmentHandler := handler.NewAttachmentHandler(attachmentService)
	activityHandler := handler.NewActivityHandler(activityService)
	notificationHandler := handler.NewNotificationHandler(notificationService, sseManager)
	invitationHandler := handler.NewInvitationHandler(invitationService)

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

	// Swagger API Documentation
	swaggerHandler := handler.NewSwaggerHandler(swaggerYAML)
	docs := r.Group("/api/docs")
	{
		docs.GET("", swaggerHandler.ServeUI)
		docs.GET("/", swaggerHandler.ServeUI)
		docs.GET("/swagger.yaml", swaggerHandler.ServeYAML)
	}

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
					MaxRequests: 50,
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

		// Public invitation endpoints (no auth required)
		api.GET("/invitations/:token", invitationHandler.GetByToken)
		api.POST("/invitations/:token/accept-with-password", invitationHandler.AcceptWithPassword)

		protected := api.Group("")
		protected.Use(middleware.Auth(jwtManager, authService))
		{
			// Protected invitation endpoints
			invitations := protected.Group("/invitations")
			{
				invitations.POST("/:token/accept", invitationHandler.Accept)
				invitations.POST("/:token/decline", invitationHandler.Decline)
			}

			orgs := protected.Group("/organizations")
			{
				orgs.POST("", orgHandler.Create)
				orgs.GET("", orgHandler.List)
				orgs.GET("/:slug", orgHandler.GetBySlug)
				orgs.PUT("/:slug", orgHandler.Update)
				orgs.DELETE("/:slug", orgHandler.Delete)
				orgs.GET("/:slug/members", orgHandler.ListMembers)
				orgs.GET("/:slug/board-members", orgHandler.ListBoardMembers)
				orgs.PUT("/:slug/boards/:boardId/members/:userId", orgHandler.UpdateBoardMemberRole)
				orgs.POST("/:slug/members", orgHandler.InviteMember)
				orgs.PUT("/:slug/members/:userId", orgHandler.UpdateMemberRole)
				orgs.DELETE("/:slug/members/:userId", orgHandler.RemoveMember)
				orgs.POST("/:slug/leave", orgHandler.Leave)

				orgs.POST("/:slug/boards", boardHandler.Create)
				orgs.GET("/:slug/boards", boardHandler.ListByOrg)
			}

			boards := protected.Group("/boards")
			{
				boards.GET("/:id", boardHandler.GetByID)
				boards.PUT("/:id", boardHandler.Update)
				boards.DELETE("/:id", boardHandler.Delete)
				boards.POST("/:id/close", boardHandler.Close)
				boards.POST("/:id/reopen", boardHandler.Reopen)
				boards.GET("/:id/members", boardHandler.ListMembers)
				boards.POST("/:id/members", boardHandler.Invite)
				boards.PUT("/:id/members/:userId", boardHandler.UpdateMemberRole)
				boards.DELETE("/:id/members/:userId", boardHandler.RemoveMember)
				boards.POST("/:id/leave", boardHandler.Leave)
				boards.GET("/:id/invitations", boardHandler.ListInvitations)
				boards.DELETE("/:id/invitations/:invitationId", boardHandler.RevokeInvitation)

				boards.POST("/:id/lists", listHandler.Create)
				boards.GET("/:id/labels", labelHandler.ListByBoard)
				boards.POST("/:id/labels", labelHandler.Create)
				boards.GET("/:id/activity", activityHandler.ListByBoard)
			}

			lists := protected.Group("/lists")
			{
				lists.PUT("/:id", listHandler.Update)
				lists.POST("/:id/archive", listHandler.Archive)
				lists.POST("/:id/restore", listHandler.Restore)
				lists.POST("/:id/move", listHandler.Move)
				lists.POST("/:id/copy", listHandler.Copy)
				lists.POST("/:id/cards", cardHandler.Create)
			}

			cards := protected.Group("/cards")
			{
				cards.GET("/:id", cardHandler.GetByID)
				cards.PUT("/:id", cardHandler.Update)
				cards.POST("/:id/archive", cardHandler.Archive)
				cards.POST("/:id/restore", cardHandler.Restore)
				cards.POST("/:id/move", cardHandler.Move)
				cards.POST("/:id/assign", cardHandler.Assign)
				cards.POST("/:id/unassign", cardHandler.Unassign)
				cards.POST("/:id/complete", cardHandler.MarkComplete)
				cards.POST("/:id/incomplete", cardHandler.MarkIncomplete)
				cards.POST("/:id/labels/:labelId", labelHandler.AssignToCard)
				cards.DELETE("/:id/labels/:labelId", labelHandler.RemoveFromCard)

				// Comments
				cards.GET("/:id/comments", commentHandler.List)
				cards.POST("/:id/comments", commentHandler.Create)

				// Checklists
				cards.GET("/:id/checklists", checklistHandler.List)
				cards.POST("/:id/checklists", checklistHandler.Create)

				// Attachments
				cards.GET("/:id/attachments", attachmentHandler.List)
				cards.POST("/:id/attachments", attachmentHandler.Upload)
				cards.DELETE("/:id/cover", attachmentHandler.RemoveCover)

				// Activity
				cards.GET("/:id/activity", activityHandler.ListByCard)
			}

			labels := protected.Group("/labels")
			{
				labels.PUT("/:id", labelHandler.Update)
				labels.DELETE("/:id", labelHandler.Delete)
			}

			comments := protected.Group("/comments")
			{
				comments.PUT("/:id", commentHandler.Update)
				comments.DELETE("/:id", commentHandler.Delete)
			}

			checklists := protected.Group("/checklists")
			{
				checklists.PUT("/:id", checklistHandler.Update)
				checklists.DELETE("/:id", checklistHandler.Delete)
				checklists.POST("/:id/items", checklistHandler.CreateItem)
			}

			checklistItems := protected.Group("/checklist-items")
			{
				checklistItems.PUT("/:id", checklistHandler.UpdateItem)
				checklistItems.DELETE("/:id", checklistHandler.DeleteItem)
				checklistItems.POST("/:id/toggle", checklistHandler.ToggleItem)
			}

			attachments := protected.Group("/attachments")
			{
				attachments.DELETE("/:id", attachmentHandler.Delete)
				attachments.POST("/:id/cover", attachmentHandler.SetCover)
				attachments.GET("/:id/download", attachmentHandler.Download)
			}

			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.List)
				notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifications.GET("/stream", notificationHandler.Stream)
				notifications.POST("/:id/read", notificationHandler.MarkAsRead)
				notifications.POST("/read-all", notificationHandler.MarkAllAsRead)
				notifications.DELETE("/:id", notificationHandler.Delete)
			}
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
