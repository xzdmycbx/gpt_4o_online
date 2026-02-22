package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/oauth2"
	"github.com/ai-chat/backend/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService           *service.AuthService
	emailService          *service.EmailService
	systemSettingsService *service.SystemSettingsService
	frontendURL           string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, emailService *service.EmailService, systemSettingsService *service.SystemSettingsService, frontendURL string) *AuthHandler {
	return &AuthHandler{
		authService:           authService,
		emailService:          emailService,
		systemSettingsService: systemSettingsService,
		frontendURL:           frontendURL,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set token in HttpOnly cookie (primary auth method)
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 86400, "/", "", isSecure, true) // 24 hour expiry, HttpOnly

	// Also return token in response for backwards compatibility
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditUserCreated)
	c.Set("audit_target_user_id", user.ID.String())

	// Set token in HttpOnly cookie (primary auth method)
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 86400, "/", "", isSecure, true) // 24 hour expiry, HttpOnly

	// Also return token in response for backwards compatibility
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// TwitterOAuth2 initiates Twitter OAuth2 flow
func (h *AuthHandler) TwitterOAuth2(c *gin.Context) {
	// Generate state
	state, err := oauth2.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Detect if connection is secure for cookie flags
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	// Store state in session/cookie for verification
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", state, 600, "/", "", isSecure, true)

	// Generate auth URL
	authURL, codeVerifier, err := h.authService.GenerateTwitterOAuth2URL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth URL"})
		return
	}

	// Store code verifier for PKCE
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_verifier", codeVerifier, 600, "/", "", isSecure, true)

	// Redirect directly to Twitter OAuth2 URL
	c.Redirect(http.StatusFound, authURL)
}

// OAuth2Callback handles OAuth2 callback
func (h *AuthHandler) OAuth2Callback(c *gin.Context) {
	// Helper function to build redirect URL
	buildRedirectURL := func(path string) string {
		if h.frontendURL != "" {
			return h.frontendURL + path
		}
		return path // Fallback to relative path for same-origin
	}

	// Get parameters
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// Handle OAuth2 error
	if errorParam != "" {
		c.Redirect(http.StatusFound, buildRedirectURL("/login?error=oauth_failed"))
		return
	}

	if code == "" {
		c.Redirect(http.StatusFound, buildRedirectURL("/login?error=missing_code"))
		return
	}

	// Verify state
	storedState, err := c.Cookie("oauth_state")
	if err != nil || !oauth2.ValidateState(storedState, state) {
		c.Redirect(http.StatusFound, buildRedirectURL("/login?error=invalid_state"))
		return
	}

	// Get code verifier
	codeVerifier, err := c.Cookie("oauth_verifier")
	if err != nil {
		c.Redirect(http.StatusFound, buildRedirectURL("/login?error=missing_verifier"))
		return
	}

	// Handle callback
	token, user, err := h.authService.HandleTwitterOAuth2Callback(c.Request.Context(), code, codeVerifier)
	if err != nil {
		c.Redirect(http.StatusFound, buildRedirectURL("/login?error=auth_failed"))
		return
	}

	// Detect if connection is secure for cookie flags
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	// Clear OAuth temporary cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", "", -1, "/", "", isSecure, true)
	c.SetCookie("oauth_verifier", "", -1, "/", "", isSecure, true)

	// SECURITY: Set JWT token in HttpOnly cookie instead of URL
	// This prevents token leakage to browser history, logs, and referer headers
	c.SetCookie("auth_token", token, 86400, "/", "", isSecure, true) // 24 hour expiry, HttpOnly

	// Redirect to frontend OAuth callback page (without token in URL)
	// Frontend will read token from cookie
	redirectURL := buildRedirectURL(fmt.Sprintf("/oauth2/callback?username=%s", user.Username))
	c.Redirect(http.StatusFound, redirectURL)
}

// ForgotPassword initiates password reset flow
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email is required"})
		return
	}

	// Send password reset email (doesn't reveal if email exists)
	_ = h.emailService.SendPasswordResetEmail(c.Request.Context(), req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.emailService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditPasswordReset)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToResponse(),
	})
}

// ChangePassword changes the user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Password must be at least 8 characters"})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditPasswordChange)
	c.Set("audit_resource_type", "user")

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// RefreshToken refreshes an expired JWT token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	newToken, err := h.authService.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}

// Logout clears the authentication token cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear auth token cookie
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", "", -1, "/", "", isSecure, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetEnabledOAuth2Providers returns the list of enabled OAuth2 providers (public endpoint)
func (h *AuthHandler) GetEnabledOAuth2Providers(c *gin.Context) {
	if h.systemSettingsService == nil {
		c.JSON(http.StatusOK, gin.H{"providers": []interface{}{}})
		return
	}

	settings, err := h.systemSettingsService.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"providers": []interface{}{}})
		return
	}

	type ProviderInfo struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		AuthURL     string `json:"auth_url"`
	}

	providers := []ProviderInfo{}

	if settings.OAuth2TwitterEnabled && settings.OAuth2TwitterClientID != "" {
		providers = append(providers, ProviderInfo{
			Name:        "twitter",
			DisplayName: "X (Twitter)",
			AuthURL:     "/api/v1/auth/oauth2/twitter",
		})
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}
