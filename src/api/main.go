package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/briandenicola/ancient-coins-api/config"
	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/handlers"
	"github.com/briandenicola/ancient-coins-api/middleware"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database.Connect(cfg.DBPath)

	// Initialize logger from DB settings
	services.SyncLogLevel()
	logger := services.AppLogger

	logger.Info("startup", "Application starting")
	logger.Info("startup", "Database connected: %s", cfg.DBPath)

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
	logger.Debug("startup", "Upload directory: %s", cfg.UploadDir)

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve uploaded images
	r.Static("/uploads", cfg.UploadDir)

	// Serve Vue SPA from wwwroot
	wwwroot := filepath.Join(".", "wwwroot")
	if _, err := os.Stat(wwwroot); err == nil {
		r.Static("/assets", filepath.Join(wwwroot, "assets"))
		r.StaticFile("/coin-logo.jpg", filepath.Join(wwwroot, "coin-logo.jpg"))
		r.StaticFile("/manifest.webmanifest", filepath.Join(wwwroot, "manifest.webmanifest"))
		r.StaticFile("/sw.js", filepath.Join(wwwroot, "sw.js"))
		r.StaticFile("/registerSW.js", filepath.Join(wwwroot, "registerSW.js"))

		// SPA fallback
		r.NoRoute(func(c *gin.Context) {
			// Don't serve index.html for API routes
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(404, gin.H{"error": "Not found"})
				return
			}
			if len(c.Request.URL.Path) >= 8 && c.Request.URL.Path[:8] == "/uploads" {
				c.JSON(404, gin.H{"error": "Not found"})
				return
			}
			c.File(filepath.Join(wwwroot, "index.html"))
		})
	}

	// Auth routes (public)
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)
	api := r.Group("/api")
	{
		api.GET("/auth/setup", authHandler.NeedsSetup)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		coinHandler := handlers.NewCoinHandler()
		protected.GET("/coins", coinHandler.List)
		protected.GET("/coins/:id", coinHandler.Get)
		protected.POST("/coins", coinHandler.Create)
		protected.PUT("/coins/:id", coinHandler.Update)
		protected.DELETE("/coins/:id", coinHandler.Delete)

		protected.GET("/stats", coinHandler.Stats)
		protected.GET("/suggestions", coinHandler.Suggestions)

		imageHandler := handlers.NewImageHandler(cfg.UploadDir)
		protected.POST("/coins/:id/images", imageHandler.Upload)
		protected.DELETE("/coins/:id/images/:imageId", imageHandler.Delete)

		analysisHandler := handlers.NewAnalysisHandler()
		protected.POST("/coins/:id/analyze", analysisHandler.Analyze)
		protected.POST("/extract-text", analysisHandler.ExtractText)
		protected.GET("/ollama-status", analysisHandler.OllamaStatus)

		// User self-service routes
		userHandler := handlers.NewUserHandler()
		protected.GET("/auth/me", userHandler.GetMe)
		protected.POST("/auth/change-password", userHandler.ChangePassword)
		protected.GET("/user/export", userHandler.ExportCollection)
		protected.POST("/user/import", userHandler.ImportCollection)
	}

	// Admin-only routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret))
	admin.Use(handlers.AdminRequired())
	{
		adminHandler := handlers.NewAdminHandler()
		admin.GET("/users", adminHandler.ListUsers)
		admin.DELETE("/users/:id", adminHandler.DeleteUser)
		admin.POST("/users/:id/reset-password", adminHandler.ResetPassword)
		admin.GET("/settings", adminHandler.GetSettings)
		admin.GET("/settings/defaults", adminHandler.GetSettingDefaults)
		admin.PUT("/settings", adminHandler.UpdateSettings)
		admin.GET("/logs", adminHandler.GetLogs)
	}

	log.Printf("Starting server on :%s", cfg.Port)
	logger.Info("startup", "Server starting on port %s", cfg.Port)
	logger.Info("startup", "Log level: %s", logger.GetLevel())

	// Check Ollama connectivity at startup (blocks until complete)
	func() {
		ollamaURL := services.GetSetting(services.SettingOllamaURL)
		ollamaModel := services.GetSetting(services.SettingOllamaModel)
		svc := services.NewOllamaService(ollamaURL, 10)
		available, msg := svc.CheckModel(ollamaModel)
		if available {
			logger.Info("startup", "Ollama: %s", msg)
		} else {
			logger.Warn("startup", "Ollama: %s — AI features will be unavailable until resolved", msg)
		}
	}()

	logger.Info("startup", "Application ready")
	log.Println("Application ready")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
