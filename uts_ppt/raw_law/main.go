package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"raw-law-api/config"
	"raw-law-api/handlers"
	"raw-law-api/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// ── Database ──────────────────────────────────────────────
	dbCfg := config.DefaultDBConfig()
	db, err := config.InitDB(dbCfg)
	if err != nil {
		log.Fatalf("❌ Database initialization failed: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// ── Router ────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Global: inject DB into every request context + health check
	r.Use(middleware.InjectDB(db))
	r.Use(middleware.DBHealthMiddleware(db))

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	// ── Root & 404 Handlers ───────────────────────────────────
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Welcome to Raw Law API! The server is running smoothly.",
			"version": "1.0",
		})
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Endpoint not found. Please check the URL and HTTP method.",
		})
	})

	// ── Health ────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error", "message": "Database unreachable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"service":  "Raw Law API",
			"database": dbCfg.DBName,
			"dsn_host": dbCfg.Host + ":" + dbCfg.Port,
		})
	})

	api := r.Group("/api")

	// ── Public Routes ─────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Public lawyer browsing
	api.GET("/lawyers", handlers.ListLawyers)
	api.GET("/lawyers/:id", handlers.GetLawyer)
	api.GET("/lawyers/:id/reviews", handlers.GetLawyerReviews)

	// ── Protected Routes ──────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Profile
		protected.GET("/profile", handlers.GetProfile)

		// Consultations
		protected.POST("/consultations", handlers.BookConsultation)
		protected.GET("/consultations", handlers.GetMyConsultations)
		protected.GET("/consultations/:id/status", handlers.GetConsultationStatus)
		protected.PATCH("/consultations/:id/cancel", handlers.CancelConsultation)
		protected.GET("/consultations/:id/payment", handlers.GetPaymentByConsultation)

		// Payment proof upload
		protected.POST("/payments/:id/upload", handlers.UploadPaymentProof)

		// Lawyer profile (lawyer role only — checked inside handler)
		protected.POST("/lawyers/profile", handlers.CreateLawyerProfile)
	}

	// ── HTTP Server ───────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Raw Law API running on http://localhost:%s", port)
		log.Printf("   DB: %s@%s:%s/%s", dbCfg.User, dbCfg.Host, dbCfg.Port, dbCfg.DBName)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Raw Law API...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server stopped gracefully.")
}
