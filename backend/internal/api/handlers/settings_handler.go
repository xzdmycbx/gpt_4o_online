package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/service"
)

// SettingsHandler handles user settings endpoints
type SettingsHandler struct {
	settingsService *service.UserSettingsService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService *service.UserSettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

// Get retrieves user settings
func (h *SettingsHandler) Get(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	settings, err := h.settingsService.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// Update updates user settings
func (h *SettingsHandler) Update(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	var req model.UserSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	settings, err := h.settingsService.UpdateSettings(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// Sync syncs settings across devices
func (h *SettingsHandler) Sync(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	var localSettings model.UserSettings
	if err := c.ShouldBindJSON(&localSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	settings, pulled, err := h.settingsService.SyncSettings(c.Request.Context(), userID, &localSettings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	action := "pushed"
	if pulled {
		action = "pulled"
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"action":   action,
	})
}

// getUserID extracts user ID from context
func (h *SettingsHandler) getUserID(c *gin.Context) uuid.UUID {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return uuid.Nil
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return uuid.Nil
	}

	return userID
}
