package router

import (
	"database/sql"
	"net/http"

	"legal-consultation-api/internal/config"
	"legal-consultation-api/internal/handler"
	"legal-consultation-api/internal/middleware"
	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/repository"
	"legal-consultation-api/internal/service"

	"github.com/gin-gonic/gin"
)

func Setup(db *sql.DB) *gin.Engine {
	if config.AppConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware(config.AppConfig.AllowedOrigins))

	// Serve uploaded files
	r.Static("/uploads", config.AppConfig.UploadDir)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": config.AppConfig.AppName})
	})

	// ── Repositories ──────────────────────────────────────────
	userRepo     := repository.NewUserRepository(db)
	lawyerRepo   := repository.NewLawyerRepository(db)
	consultRepo  := repository.NewConsultationRepository(db)
	paymentRepo  := repository.NewPaymentRepository(db)
	chatRepo     := repository.NewChatRepository(db)
	reviewRepo   := repository.NewReviewRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)

	// ── Services ──────────────────────────────────────────────
	authService     := service.NewAuthService(userRepo, lawyerRepo)
	lawyerService   := service.NewLawyerService(lawyerRepo, userRepo)
	consultService  := service.NewConsultationService(consultRepo, lawyerRepo, paymentRepo)
	paymentService  := service.NewPaymentService(paymentRepo, consultRepo)
	chatService     := service.NewChatService(chatRepo, consultRepo)
	reviewService   := service.NewReviewService(reviewRepo, consultRepo, lawyerRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, lawyerRepo)

	// ── Handlers ──────────────────────────────────────────────
	authHandler         := handler.NewAuthHandler(authService)
	lawyerHandler       := handler.NewLawyerHandler(lawyerService, scheduleService, reviewService)
	consultationHandler := handler.NewConsultationHandler(consultService, reviewService)
	paymentHandler      := handler.NewPaymentHandler(paymentService)
	chatHandler         := handler.NewChatHandler(chatService)

	api := r.Group("/api")

	// ── Public routes ─────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Public lawyer search
	api.GET("/lawyers", lawyerHandler.SearchLawyers)
	api.GET("/lawyers/:id", lawyerHandler.GetLawyer)
	api.GET("/lawyers/:id/schedules", lawyerHandler.GetSchedules)
	api.GET("/lawyers/:id/reviews", lawyerHandler.GetReviews)

	// ── Authenticated routes ───────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Profile
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)

		// Consultations (client + lawyer)
		protected.POST("/consultations", consultationHandler.Book)
		protected.GET("/consultations", consultationHandler.GetMyConsultations)
		protected.GET("/consultations/:id", consultationHandler.GetByID)
		protected.GET("/consultations/:id/status", consultationHandler.GetStatus)
		protected.PATCH("/consultations/:id/cancel", consultationHandler.Cancel)

		// Consultation actions
		protected.PATCH("/consultations/:id/confirm",
			middleware.RoleMiddleware(models.RoleLawyer),
			consultationHandler.Confirm)
		protected.PATCH("/consultations/:id/complete",
			middleware.RoleMiddleware(models.RoleLawyer),
			consultationHandler.Complete)

		// Reviews (client only)
		protected.POST("/consultations/:id/reviews",
			middleware.RoleMiddleware(models.RoleClient),
			consultationHandler.CreateReview)

		// Chat
		protected.GET("/consultations/:id/messages", chatHandler.GetMessages)
		protected.POST("/consultations/:id/messages", chatHandler.SendMessage)
		protected.GET("/consultations/:id/ws", chatHandler.WebSocket)

		// Payment
		protected.GET("/consultations/:id/payment", paymentHandler.GetByConsultation)
		protected.POST("/payments/:id/upload", paymentHandler.UploadProof)

		// Lawyer profile management
		lawyerRoutes := protected.Group("")
		lawyerRoutes.Use(middleware.RoleMiddleware(models.RoleLawyer))
		{
			lawyerRoutes.POST("/lawyers/profile", lawyerHandler.CreateProfile)
			lawyerRoutes.PUT("/lawyers/profile", lawyerHandler.UpdateProfile)
			lawyerRoutes.PATCH("/lawyers/availability", lawyerHandler.SetAvailability)
			lawyerRoutes.POST("/schedules", lawyerHandler.CreateSchedule)
			lawyerRoutes.DELETE("/schedules/:id", lawyerHandler.DeleteSchedule)
		}

		// Admin routes
		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(middleware.RoleMiddleware(models.RoleAdmin))
		{
			adminRoutes.PATCH("/payments/:id/verify", paymentHandler.Verify)
			adminRoutes.PATCH("/payments/:id/reject", paymentHandler.Reject)
		}
	}

	return r
}
