package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/briandenicola/ancient-coins-api/config"
	"github.com/briandenicola/ancient-coins-api/database"
	_ "github.com/briandenicola/ancient-coins-api/docs"
	"github.com/briandenicola/ancient-coins-api/handlers"
	"github.com/briandenicola/ancient-coins-api/middleware"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title						Ancient Coins API
//	@version					1.0
//	@description				REST API for managing an ancient coin collection. Supports coin CRUD, image uploads, AI-powered analysis via Ollama, user management, and admin features.
//	@BasePath					/api
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter your JWT token with the Bearer prefix, e.g. "Bearer eyJhbGci..."

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				Enter your API key, e.g. "ak_a1b2c3d4..."

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

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes (public)
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)
	webauthnHandler, err := handlers.NewWebAuthnHandler(cfg.WebAuthnID, cfg.WebAuthnOrigin, authHandler)
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn: %v", err)
	}

	api := r.Group("/api")
	{
		api.GET("/auth/setup", authHandler.NeedsSetup)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)

		// WebAuthn public routes (login ceremony)
		api.POST("/auth/webauthn/login/begin", webauthnHandler.LoginBegin)
		api.POST("/auth/webauthn/login/finish", webauthnHandler.LoginFinish)
		api.GET("/auth/webauthn/check", webauthnHandler.CheckCredentials)
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
		protected.POST("/coins/:id/purchase", coinHandler.Purchase)
		protected.DELETE("/coins/:id", coinHandler.Delete)

		journalHandler := handlers.NewJournalHandler()
		protected.GET("/coins/:id/journal", journalHandler.ListEntries)
		protected.POST("/coins/:id/journal", journalHandler.AddEntry)
		protected.DELETE("/coins/:id/journal/:entryId", journalHandler.DeleteEntry)

		protected.GET("/stats", coinHandler.Stats)
		protected.GET("/value-history", coinHandler.ValueHistory)
		protected.GET("/suggestions", coinHandler.Suggestions)

		imageHandler := handlers.NewImageHandler(cfg.UploadDir)
		protected.POST("/coins/:id/images", imageHandler.Upload)
		protected.POST("/coins/:id/images/base64", imageHandler.UploadBase64)
		protected.DELETE("/coins/:id/images/:imageId", imageHandler.Delete)
		protected.GET("/proxy-image", imageHandler.ProxyImage)

		analysisHandler := handlers.NewAnalysisHandler()
		protected.POST("/coins/:id/analyze", analysisHandler.Analyze)
		protected.DELETE("/coins/:id/analyze", analysisHandler.DeleteAnalysis)
		protected.POST("/extract-text", analysisHandler.ExtractText)
		protected.GET("/ollama-status", analysisHandler.OllamaStatus)

		numistaHandler := handlers.NewNumistaHandler()
		protected.GET("/numista/search", numistaHandler.Search)

		// User self-service routes
		userHandler := handlers.NewUserHandler(cfg.UploadDir)
		protected.GET("/auth/me", userHandler.GetMe)
		protected.POST("/auth/change-password", userHandler.ChangePassword)
		protected.GET("/user/export", userHandler.ExportCollection)
		protected.POST("/user/import", userHandler.ImportCollection)

		// API key management
		apiKeyHandler := handlers.NewApiKeyHandler()
		protected.POST("/auth/api-keys", apiKeyHandler.Generate)
		protected.GET("/auth/api-keys", apiKeyHandler.List)
		protected.DELETE("/auth/api-keys/:id", apiKeyHandler.Revoke)

		// WebAuthn registration (requires auth)
		protected.POST("/auth/webauthn/register/begin", webauthnHandler.RegisterBegin)
		protected.POST("/auth/webauthn/register/finish", webauthnHandler.RegisterFinish)
		protected.GET("/auth/webauthn/credentials", webauthnHandler.ListCredentials)
		protected.DELETE("/auth/webauthn/credentials/:id", webauthnHandler.DeleteCredential)
	}

	// Admin-only routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret))
	admin.Use(handlers.AdminRequired())
	{
		adminHandler := handlers.NewAdminHandler(cfg.UploadDir)
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
