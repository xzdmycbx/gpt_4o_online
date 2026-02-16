package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ai-chat/backend/internal/pkg/jwt"
	"github.com/ai-chat/backend/internal/repository"
)

// AuthMiddleware validates JWT tokens and attaches user info to context
func AuthMiddleware(jwtManager *jwt.Manager, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Extract token from "Bearer <token>" format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If no Authorization header, try to get from HttpOnly cookie (OAuth2)
		if tokenString == "" {
			var err error
			tokenString, err = c.Cookie("auth_token")
			if err != nil || tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
				c.Abort()
				return
			}
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get user from database to check if banned
		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user is banned
		if user.IsBanned {
			c.JSON(http.StatusForbidden, gin.H{"error": "Your account has been banned", "reason": user.BanReason})
			c.Abort()
			return
		}

		// Attach user info to context
		c.Set("user_id", claims.UserID.String())
		c.Set("user", user)
		c.Set("username", claims.Username)
		c.Set("role", user.Role) // Use role from database, not stale JWT claims

		c.Next()
	}
}

// OptionalAuthMiddleware validates token if present but doesn't require it
func OptionalAuthMiddleware(jwtManager *jwt.Manager, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		if !user.IsBanned {
			c.Set("user_id", claims.UserID.String())
			c.Set("user", user)
			c.Set("username", claims.Username)
			c.Set("role", user.Role) // Use role from database, not stale JWT claims
		}

		c.Next()
	}
}
