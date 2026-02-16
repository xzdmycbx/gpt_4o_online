package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRF token store (in-memory, should use Redis in production)
type csrfStore struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var globalCSRFStore = &csrfStore{
	tokens: make(map[string]time.Time),
}

// CSRFMiddleware provides CSRF protection for state-changing requests
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF check for auth endpoints (they use other protection)
		if isAuthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get CSRF token from header
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			c.Abort()
			return
		}

		// Validate token
		if !globalCSRFStore.validate(token) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a new CSRF token for the session
func GenerateCSRFToken(c *gin.Context) {
	token := generateRandomToken()

	// Store token with expiration (24 hours)
	globalCSRFStore.store(token, 24*time.Hour)

	// Send token in response header and cookie
	c.Header("X-CSRF-Token", token)

	// Set cookie (not HttpOnly so JavaScript can read it)
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("csrf_token", token, 86400, "/", "", isSecure, false)

	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}

// generateRandomToken generates a cryptographically secure random token
func generateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// store stores a CSRF token with expiration
func (s *csrfStore) store(token string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean expired tokens periodically
	now := time.Now()
	for t, expiry := range s.tokens {
		if now.After(expiry) {
			delete(s.tokens, t)
		}
	}

	s.tokens[token] = now.Add(duration)
}

// validate checks if a CSRF token is valid and not expired
func (s *csrfStore) validate(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expiry, exists := s.tokens[token]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		return false
	}

	return true
}

// isAuthEndpoint checks if the path is an auth endpoint
func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
		"/api/v1/auth/refresh",
		"/api/v1/auth/oauth2/twitter",
		"/api/v1/auth/oauth2/callback",
		"/api/v1/logout", // Logout doesn't need CSRF (safe operation)
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}

	return false
}
