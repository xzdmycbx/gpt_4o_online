package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ai-chat/backend/internal/model"
)

// MaintenanceModeMiddleware blocks access during maintenance except for admins
func MaintenanceModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if maintenance mode is enabled via environment variable
		maintenanceMode := os.Getenv("MAINTENANCE_MODE")
		if maintenanceMode != "true" {
			c.Next()
			return
		}

		// Allow health check endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Check if user is authenticated and is admin/super_admin
		role, exists := c.Get("role")
		if exists {
			userRole := role.(model.UserRole)
			if userRole == model.RoleAdmin || userRole == model.RoleSuperAdmin {
				c.Next()
				return
			}
		}

		// Return maintenance mode message for all other users
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service Unavailable",
			"message": "System is currently under maintenance. Please try again later.",
			"message_zh": "系统正在维护中，请稍后再试。",
		})
		c.Abort()
	}
}
