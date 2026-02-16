package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/ai-chat/backend/internal/api"
	"github.com/ai-chat/backend/internal/api/handlers"
	"github.com/ai-chat/backend/internal/api/middleware"
	"github.com/ai-chat/backend/internal/config"
	"github.com/ai-chat/backend/internal/database"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/crypto"
	"github.com/ai-chat/backend/internal/pkg/email"
	"github.com/ai-chat/backend/internal/pkg/geo"
	"github.com/ai-chat/backend/internal/pkg/jwt"
	"github.com/ai-chat/backend/internal/pkg/oauth2"
	"github.com/ai-chat/backend/internal/pkg/ratelimit"
	"github.com/ai-chat/backend/internal/repository"
	"github.com/ai-chat/backend/internal/service"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting AI Chat Server in %s mode...", cfg.Env)

	// Initialize database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Load dynamic configuration from database
	if err := cfg.LoadDynamicConfig(db.DB); err != nil {
		log.Printf("Warning: Failed to load dynamic config from database: %v", err)
		log.Println("Using environment variables as configuration source")
	} else {
		log.Println("Dynamic configuration loaded from database")
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected successfully")

	// Initialize IP checker (GeoIP)
	var ipChecker *geo.IPChecker
	if cfg.GeoIP.BlockChina {
		ipChecker, err = geo.NewIPChecker(cfg.GeoIP.DBPath, cfg.GeoIP.BlockChina)
		if err != nil {
			log.Printf("Warning: Failed to initialize GeoIP checker: %v", err)
		} else {
			defer ipChecker.Close()
			log.Println("GeoIP checker initialized")
		}
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration)

	// Initialize rate limiter with default limit
	rateLimiter := ratelimit.NewLimiter(redisClient, cfg.RateLimit.DefaultPerMinute)

	// Initialize OAuth2 clients
	var twitterOAuth2 *oauth2.TwitterOAuth2Client
	if cfg.OAuth2.TwitterClientID != "" {
		twitterOAuth2 = oauth2.NewTwitterOAuth2Client(
			cfg.OAuth2.TwitterClientID,
			cfg.OAuth2.TwitterClientSecret,
			cfg.OAuth2.TwitterRedirectURL,
		)
		log.Println("Twitter OAuth2 initialized")
	}

	// Initialize email sender
	var emailSender *email.Sender
	if cfg.Email.Provider == "smtp" && cfg.Email.SMTPHost != "" {
		emailSender = email.NewSender(
			cfg.Email.Provider,
			cfg.Email.SMTPHost,
			cfg.Email.SMTPPort,
			cfg.Email.SMTPUser,
			cfg.Email.SMTPPassword,
			cfg.Email.EmailFrom,
			"",
			"",
		)
		log.Println("SMTP email sender initialized")
	} else if cfg.Email.Provider == "resend" && cfg.Email.ResendAPIKey != "" {
		emailSender = email.NewSender(
			cfg.Email.Provider,
			"",
			0,
			"",
			"",
			cfg.Email.EmailFrom,
			"",
			cfg.Email.ResendAPIKey,
		)
		log.Println("Resend email sender initialized")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	convRepo := repository.NewConversationRepository(db.DB)
	msgRepo := repository.NewMessageRepository(db.DB)
	memoryRepo := repository.NewMemoryRepository(db.DB)
	modelRepo := repository.NewAIModelRepository(db.DB)
	settingsRepo := repository.NewUserSettingsRepository(db.DB)
	auditRepo := repository.NewAuditLogRepository(db.DB)
	tokenUsageRepo := repository.NewTokenUsageRepository(db.DB)
	systemSettingsRepo := repository.NewSystemSettingsRepository(db.DB)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db.DB)

	// Initialize services
	aiProxyService := service.NewAIProxyService(modelRepo, cfg.Encryption.Key)
	memoryService := service.NewMemoryService(
		memoryRepo,
		msgRepo,
		aiProxyService,
		modelRepo,
		cfg.AI.MemoryExtractionEnabled,
		cfg.AI.DefaultMemoryModel,
	)
	chatService := service.NewChatService(
		convRepo,
		msgRepo,
		modelRepo,
		tokenUsageRepo,
		aiProxyService,
		memoryService,
	)
	authService := service.NewAuthService(userRepo, jwtManager, twitterOAuth2)
	emailService := service.NewEmailService(emailSender, userRepo, resetTokenRepo, cfg.Server.FrontendURL)
	settingsService := service.NewUserSettingsService(settingsRepo)
	systemSettingsService := service.NewSystemSettingsService(systemSettingsRepo, cfg.Encryption.Key)
	adminService := service.NewAdminService(userRepo, modelRepo, auditRepo, tokenUsageRepo, convRepo, msgRepo, cfg.Encryption.Key)

	// Load default rate limit from database (override env var if exists)
	if defaultLimit, err := systemSettingsService.GetRateLimitDefault(ctx); err == nil && defaultLimit > 0 {
		rateLimiter = ratelimit.NewLimiter(redisClient, defaultLimit)
		log.Printf("Loaded rate limit from database: %d/min", defaultLimit)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, emailService, cfg.Server.FrontendURL)
	chatHandler := handlers.NewChatHandler(chatService)
	memoryHandler := handlers.NewMemoryHandler(memoryService)
	settingsHandler := handlers.NewSettingsHandler(settingsService)
	adminHandler := handlers.NewAdminHandler(adminService, systemSettingsService)

	// Setup router
	routerConfig := &api.RouterConfig{
		JWTManager:  jwtManager,
		RateLimiter: rateLimiter,
		IPChecker:   ipChecker,
		Config:      cfg,

		UserRepo:       userRepo,
		ConvRepo:       convRepo,
		MsgRepo:        msgRepo,
		MemoryRepo:     memoryRepo,
		ModelRepo:      modelRepo,
		SettingsRepo:   settingsRepo,
		AuditRepo:      auditRepo,
		TokenUsageRepo: tokenUsageRepo,

		AuthHandler:     authHandler,
		ChatHandler:     chatHandler,
		MemoryHandler:   memoryHandler,
		AdminHandler:    adminHandler,
		SettingsHandler: settingsHandler,
	}

	router := setupRouter(cfg, routerConfig)

	// Create super admin if configured (first time only)
	createSuperAdminIfNeeded(ctx, userRepo)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Server started on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server stopped")
}

func setupRouter(cfg *config.Config, routerCfg *api.RouterConfig) *gin.Engine {
	router := api.SetupRouter(routerCfg)

	// Connect handlers to routes
	v1 := router.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		// Apply IP-based rate limiting to auth endpoints to prevent brute force
		// Limit: 10 requests per minute per IP address
		authRateLimit := middleware.IPRateLimitMiddleware(routerCfg.RateLimiter, 10)

		auth.POST("/login", authRateLimit, routerCfg.AuthHandler.Login)
		auth.POST("/register", authRateLimit, routerCfg.AuthHandler.Register)
		auth.GET("/oauth2/twitter", routerCfg.AuthHandler.TwitterOAuth2)
		auth.GET("/oauth2/callback", routerCfg.AuthHandler.OAuth2Callback)
		auth.POST("/forgot-password", authRateLimit, routerCfg.AuthHandler.ForgotPassword)
		auth.POST("/reset-password", authRateLimit, routerCfg.AuthHandler.ResetPassword)
		auth.POST("/refresh", authRateLimit, routerCfg.AuthHandler.RefreshToken)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(routerCfg.JWTManager, routerCfg.UserRepo))
	protected.Use(middleware.AuditMiddleware(routerCfg.AuditRepo))
	{
		protected.GET("/me", routerCfg.AuthHandler.GetCurrentUser)
		protected.POST("/logout", routerCfg.AuthHandler.Logout)

		// Conversations
		conversations := protected.Group("/conversations")
		{
			conversations.GET("", routerCfg.ChatHandler.ListConversations)
			conversations.POST("", routerCfg.ChatHandler.CreateConversation)
			conversations.GET("/:id", routerCfg.ChatHandler.GetConversation)
			conversations.PUT("/:id", routerCfg.ChatHandler.UpdateConversation)
			conversations.DELETE("/:id", routerCfg.ChatHandler.DeleteConversation)
			conversations.GET("/:id/messages", routerCfg.ChatHandler.GetMessages)

			// Apply rate limiting ONLY to chat message sending
			conversations.POST("/:id/messages",
				middleware.RateLimitMiddleware(routerCfg.RateLimiter),
				routerCfg.ChatHandler.SendMessage,
			)
		}

		// Memories
		memories := protected.Group("/memories")
		{
			memories.GET("", routerCfg.MemoryHandler.List)
			memories.POST("", routerCfg.MemoryHandler.Create)
			memories.PUT("/:id", routerCfg.MemoryHandler.Update)
			memories.DELETE("/:id", routerCfg.MemoryHandler.Delete)
		}

		// User settings
		settings := protected.Group("/user/settings")
		{
			settings.GET("", routerCfg.SettingsHandler.Get)
			settings.PUT("", routerCfg.SettingsHandler.Update)
			settings.POST("/sync", routerCfg.SettingsHandler.Sync)
		}

		// User password management
		protected.PUT("/user/password", routerCfg.AuthHandler.ChangePassword)

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			users := admin.Group("/users")
			{
				users.GET("", routerCfg.AdminHandler.ListUsers)
				users.GET("/:id", routerCfg.AdminHandler.GetUser)
				users.PUT("/:id", routerCfg.AdminHandler.UpdateUser)
				users.PUT("/:id/ban", routerCfg.AdminHandler.BanUser)
				users.PUT("/:id/unban", routerCfg.AdminHandler.UnbanUser)
				users.DELETE("/:id", routerCfg.AdminHandler.DeleteUser)
				users.PUT("/:id/rate-limit", routerCfg.AdminHandler.SetUserRateLimit)
			}

			models := admin.Group("/models")
			{
				models.GET("", routerCfg.AdminHandler.ListModels)
				models.POST("", routerCfg.AdminHandler.CreateModel)
				models.PUT("/:id", routerCfg.AdminHandler.UpdateModel)
				models.DELETE("/:id", routerCfg.AdminHandler.DeleteModel)
				models.PUT("/:id/default", routerCfg.AdminHandler.SetDefaultModel)
			}

			stats := admin.Group("/statistics")
			{
				stats.GET("/tokens", routerCfg.AdminHandler.TokenLeaderboard)
				stats.GET("/overview", routerCfg.AdminHandler.SystemOverview)
			}

			admin.GET("/audit-logs", routerCfg.AdminHandler.ListAuditLogs)

			// System settings
			admin.GET("/system/settings", routerCfg.AdminHandler.GetSystemSettings)
			admin.PUT("/system/settings", routerCfg.AdminHandler.UpdateSystemSettings)
			admin.POST("/system/test-email", routerCfg.AdminHandler.TestEmailConfiguration)
		}

		// Super admin routes
		superAdmin := protected.Group("/super-admin")
		superAdmin.Use(middleware.RequireSuperAdmin())
		{
			superAdmin.PUT("/users/:id/role", routerCfg.AdminHandler.ChangeUserRole)
		}

		// WebSocket streaming - without rate limiting (rate limit only on message send)
		protected.GET("/chat/stream",
			routerCfg.ChatHandler.StreamChat,
		)
	}

	// Serve static files (frontend)
	serveStaticFiles(router, cfg)

	return router
}

func serveStaticFiles(router *gin.Engine, cfg *config.Config) {
	// Serve embedded static files
	staticFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		log.Printf("Warning: Failed to load static files: %v", err)
		return
	}

	// Serve assets directory
	assetsFS, err := fs.Sub(staticFS, "assets")
	if err == nil {
		router.StaticFS("/assets", http.FS(assetsFS))
	}

	// Serve index.html for all non-API routes (SPA routing)
	router.NoRoute(func(c *gin.Context) {
		// If it's an API route that doesn't exist, return 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Serve index.html for frontend routes
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not found")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}

func createSuperAdminIfNeeded(ctx context.Context, userRepo *repository.UserRepository) {
	username := os.Getenv("SUPER_ADMIN_USERNAME")
	password := os.Getenv("SUPER_ADMIN_PASSWORD")
	email := os.Getenv("SUPER_ADMIN_EMAIL")

	if username == "" || password == "" {
		log.Println("Warning: SUPER_ADMIN_USERNAME or SUPER_ADMIN_PASSWORD not set")
		log.Println("Please configure super admin in .env file")
		return
	}

	// Check if user already exists
	existingUser, _ := userRepo.GetByUsername(ctx, username)
	if existingUser != nil {
		// User exists - verify it's already a super admin
		if existingUser.Role == model.RoleSuperAdmin {
			// Already super admin, nothing to do
			return
		}
		// SECURITY: Do NOT auto-promote existing users to super admin
		// This prevents privilege escalation if someone registers the super admin username first
		log.Printf("ERROR: User '%s' already exists but is not a super admin", username)
		log.Printf("This is a security risk - the configured super admin username is taken by another user")
		log.Printf("Please either: 1) Use a different SUPER_ADMIN_USERNAME, or 2) Manually promote user in database")
		return
	}

	// Create new super admin
	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		log.Printf("Error: Failed to hash password: %v", err)
		return
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	user := &model.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        emailPtr,
		PasswordHash: &passwordHash,
		Role:         model.RoleSuperAdmin,
	}

	if err := userRepo.Create(ctx, user); err != nil {
		log.Printf("Warning: Failed to create super admin: %v", err)
		log.Println("Super admin may already exist or database not ready")
	} else {
		log.Printf("✓ Super admin created: %s", username)
	}
}
