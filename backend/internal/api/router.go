package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ai-chat/backend/internal/api/handlers"
	"github.com/ai-chat/backend/internal/api/middleware"
	"github.com/ai-chat/backend/internal/config"
	"github.com/ai-chat/backend/internal/pkg/geo"
	"github.com/ai-chat/backend/internal/pkg/jwt"
	"github.com/ai-chat/backend/internal/pkg/ratelimit"
	"github.com/ai-chat/backend/internal/repository"
)

// RouterConfig holds dependencies for router setup
type RouterConfig struct {
	JWTManager  *jwt.Manager
	RateLimiter *ratelimit.Limiter
	IPChecker   *geo.IPChecker
	Config      *config.Config

	// Repositories
	UserRepo         *repository.UserRepository
	ConvRepo         *repository.ConversationRepository
	MsgRepo          *repository.MessageRepository
	MemoryRepo       *repository.MemoryRepository
	ModelRepo        *repository.AIModelRepository
	SettingsRepo     *repository.UserSettingsRepository
	AuditRepo        *repository.AuditLogRepository
	TokenUsageRepo   *repository.TokenUsageRepository

	// Handlers (will be initialized in Stage 4)
	AuthHandler     *handlers.AuthHandler
	ChatHandler     *handlers.ChatHandler
	MemoryHandler   *handlers.MemoryHandler
	AdminHandler    *handlers.AdminHandler
	SettingsHandler *handlers.SettingsHandler
}

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	// Set Gin mode
	if cfg.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Configure trusted proxies for secure IP detection
	// This is critical for IPRateLimitMiddleware and GeoFilterMiddleware to work correctly
	var err error
	if len(cfg.Config.Server.TrustedProxies) > 0 {
		// Use configured trusted proxies from environment variable
		err = router.SetTrustedProxies(cfg.Config.Server.TrustedProxies)
		if err != nil {
			log.Printf("Warning: Failed to set trusted proxies from config: %v. Using defaults.", err)
			// Fall through to defaults
		}
	}

	if err != nil || len(cfg.Config.Server.TrustedProxies) == 0 {
		if cfg.Config.IsProduction() {
			// In production without explicit config, use conservative defaults (localhost only)
			// This prevents IP spoofing but may need adjustment for reverse proxy setups
			if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
				log.Printf("Warning: Failed to set default trusted proxies: %v", err)
			}
		} else {
			// In development, trust all proxies for easier testing
			// NEVER use this in production!
			if err := router.SetTrustedProxies(nil); err != nil {
				log.Printf("Warning: Failed to set trusted proxies to nil: %v", err)
			}
		}
	}

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.MaintenanceModeMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.InputValidationMiddleware())

	// CORS
	allowOrigins := []string{cfg.Config.Server.FrontendURL}
	if cfg.Config.IsDevelopment() {
		allowOrigins = append(allowOrigins, "http://localhost:3000", "http://localhost:8080")
	}
	router.Use(middleware.CORSMiddleware(allowOrigins))

	// Geo filtering (if enabled)
	if cfg.IPChecker != nil {
		router.Use(middleware.GeoFilterMiddleware(cfg.IPChecker))
		router.Use(middleware.GeoInfo(cfg.IPChecker))
	}

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// CSRF token endpoint (no auth required)
	router.GET("/api/v1/csrf-token", middleware.GenerateCSRFToken)

	// Add CSRF protection for state-changing requests
	router.Use(middleware.CSRFMiddleware())

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no auth)
		_ = v1.Group("/auth")
		// Note: Auth routes are defined in cmd/server/main.go setupRouter()

		// Protected routes (require auth)
		_ = v1.Group("")
		// Note: Protected routes are defined in cmd/server/main.go setupRouter()
	}

	// Serve static files (frontend) - will be implemented in Stage 4
	// router.NoRoute(...)

	return router
}
