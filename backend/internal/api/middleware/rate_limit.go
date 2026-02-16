package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/ratelimit"
)

// RateLimitMiddleware applies rate limiting to chat API requests
// Note: This middleware should only be applied to chat message sending endpoints,
// not to all protected routes
func RateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Check user-specific rate limit (authenticated users only)
		userID, exists := c.Get("user_id")
		if !exists {
			// If not authenticated, skip rate limiting (auth middleware should have blocked this)
			c.Next()
			return
		}

		user, userExists := c.Get("user")
		if !userExists {
			c.Next()
			return
		}

		currentUser := user.(*model.User)

		// Skip rate limit check if user is exempt
		if currentUser.RateLimitExempt {
			c.Next()
			return
		}

		// Check user rate limit
		exceeded, err := limiter.CheckUserLimit(ctx, userID.(string), currentUser.CustomRateLimit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user rate limit"})
			c.Abort()
			return
		}

		if exceeded {
			remaining, _ := limiter.GetRemainingRequests(ctx, userID.(string), currentUser.CustomRateLimit)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "您发送消息过于频繁，请稍后再试",
				"remaining": remaining,
			})
			c.Abort()
			return
		}

		// Add rate limit info to response headers
		remaining, _ := limiter.GetRemainingRequests(ctx, userID.(string), currentUser.CustomRateLimit)
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

// IPRateLimitMiddleware applies IP-based rate limiting for unauthenticated endpoints
// This prevents brute force attacks on login, register, and password reset endpoints
func IPRateLimitMiddleware(limiter *ratelimit.Limiter, maxRequestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get client IP address
		clientIP := c.ClientIP()

		// Use IP as the rate limit key
		key := "ip:" + clientIP

		// Check IP-based rate limit
		exceeded, err := limiter.CheckUserLimit(ctx, key, &maxRequestsPerMinute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check rate limit"})
			c.Abort()
			return
		}

		if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests from this IP address. Please try again later.",
				"error_zh": "您的请求过于频繁，请稍后再试。",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitInfo middleware adds rate limit information to response headers
func RateLimitInfo(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		userIDStr, exists := c.Get("user_id")
		if !exists {
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			return
		}

		user, exists := c.Get("user")
		if !exists {
			return
		}

		currentUser := user.(*model.User)
		if currentUser.RateLimitExempt {
			c.Header("X-RateLimit-Remaining", "unlimited")
			return
		}

		count, err := limiter.GetUserRequestCount(c.Request.Context(), userID.String())
		if err == nil {
			c.Header("X-RateLimit-Used", strconv.FormatInt(count, 10))
		}
	}
}
