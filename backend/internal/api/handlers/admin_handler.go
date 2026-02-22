package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/service"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	adminService           *service.AdminService
	systemSettingsService *service.SystemSettingsService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *service.AdminService, systemSettingsService *service.SystemSettingsService) *AdminHandler {
	return &AdminHandler{
		adminService:           adminService,
		systemSettingsService: systemSettingsService,
	}
}

// ListUsers lists all users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.adminService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	userResponses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  userResponses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUser retrieves a user by ID
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.adminService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateUser updates a user
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	adminUser := h.getAdminUser(c)
	if adminUser == nil {
		return
	}

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req model.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.adminService.UpdateUser(c.Request.Context(), adminUser, targetUserID, &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditUserUpdated)
	c.Set("audit_target_user_id", targetUserID.String())

	c.JSON(http.StatusOK, user.ToResponse())
}

// BanUser bans a user
func (h *AdminHandler) BanUser(c *gin.Context) {
	adminUser := h.getAdminUser(c)
	if adminUser == nil {
		return
	}

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}

	if err := h.adminService.BanUser(c.Request.Context(), adminUser, targetUserID, req.Reason); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditUserBanned)
	c.Set("audit_target_user_id", targetUserID.String())
	c.Set("audit_details", map[string]interface{}{"reason": req.Reason})

	c.JSON(http.StatusOK, gin.H{"message": "User banned successfully"})
}

// UnbanUser unbans a user
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	adminUser := h.getAdminUser(c)
	if adminUser == nil {
		return
	}

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.adminService.UnbanUser(c.Request.Context(), adminUser, targetUserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditUserUnbanned)
	c.Set("audit_target_user_id", targetUserID.String())

	c.JSON(http.StatusOK, gin.H{"message": "User unbanned successfully"})
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminUser := h.getAdminUser(c)
	if adminUser == nil {
		return
	}

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), adminUser, targetUserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ChangeUserRole changes a user's role (super admin only)
func (h *AdminHandler) ChangeUserRole(c *gin.Context) {
	adminUser := h.getAdminUser(c)
	if adminUser == nil {
		return
	}

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Role model.UserRole `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.adminService.ChangeUserRole(c.Request.Context(), adminUser, targetUserID, req.Role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditPermissionChanged)
	c.Set("audit_target_user_id", targetUserID.String())
	c.Set("audit_details", map[string]interface{}{"new_role": req.Role})

	c.JSON(http.StatusOK, gin.H{"message": "User role changed successfully"})
}

// ListModels lists AI models
func (h *AdminHandler) ListModels(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	models, err := h.adminService.ListAIModels(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if models == nil {
		models = []*model.AIModel{}
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// CreateModel creates a new AI model
func (h *AdminHandler) CreateModel(c *gin.Context) {
	adminUserID := h.getAdminUserID(c)
	if adminUserID == uuid.Nil {
		return
	}

	var req model.AIModelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	aiModel, err := h.adminService.CreateAIModel(c.Request.Context(), adminUserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditModelCreated)
	c.Set("audit_details", map[string]interface{}{"model_name": aiModel.Name})

	c.JSON(http.StatusCreated, aiModel)
}

// UpdateModel updates an AI model
func (h *AdminHandler) UpdateModel(c *gin.Context) {
	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	var req model.AIModelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	aiModel, err := h.adminService.UpdateAIModel(c.Request.Context(), modelID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditModelUpdated)
	c.Set("audit_details", map[string]interface{}{"model_id": modelID})

	c.JSON(http.StatusOK, aiModel)
}

// DeleteModel deletes an AI model
func (h *AdminHandler) DeleteModel(c *gin.Context) {
	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	if err := h.adminService.DeleteAIModel(c.Request.Context(), modelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set audit action
	c.Set("audit_action", model.AuditModelDeleted)
	c.Set("audit_details", map[string]interface{}{"model_id": modelID})

	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}

// SetDefaultModel sets a model as default
func (h *AdminHandler) SetDefaultModel(c *gin.Context) {
	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	if err := h.adminService.SetDefaultModel(c.Request.Context(), modelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default model set successfully"})
}

// TokenLeaderboard retrieves token usage leaderboard
func (h *AdminHandler) TokenLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	leaderboard, err := h.adminService.GetTokenLeaderboard(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if leaderboard == nil {
		leaderboard = []*model.TokenLeaderboard{}
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
	})
}

// SystemOverview retrieves system statistics
func (h *AdminHandler) SystemOverview(c *gin.Context) {
	overview, err := h.adminService.GetSystemOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// ListAuditLogs lists audit logs
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.adminService.ListAuditLogs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if logs == nil {
		logs = []*model.AuditLog{}
	}

	// Estimate total (actual count would require additional query)
	total := offset + len(logs)
	if len(logs) >= limit {
		total = offset + len(logs) + 1 // Indicate there are more
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// SetUserRateLimit sets custom rate limit for a user
func (h *AdminHandler) SetUserRateLimit(c *gin.Context) {
	adminUser := h.getAdminUser(c)
	if adminUser == nil {
		return
	}

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Limit  *int `json:"limit"`
		Exempt bool `json:"exempt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.adminService.SetUserRateLimit(c.Request.Context(), adminUser, targetUserID, req.Limit, req.Exempt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rate limit updated successfully"})
}

// getAdminUser retrieves the admin user from context
func (h *AdminHandler) getAdminUser(c *gin.Context) *model.User {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil
	}

	return user.(*model.User)
}

// getAdminUserID retrieves the admin user ID from context
func (h *AdminHandler) getAdminUserID(c *gin.Context) uuid.UUID {
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

// GetSystemSettings retrieves current system settings
func (h *AdminHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.systemSettingsService.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system settings"})
		return
	}

	// 掩码敏感信息
	settings.MaskSensitiveData()

	c.JSON(http.StatusOK, settings)
}

// UpdateSystemSettings updates system settings
func (h *AdminHandler) UpdateSystemSettings(c *gin.Context) {
	var dto model.SystemSettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate settings
	if dto.RateLimitDefaultPerMinute < 1 || dto.RateLimitDefaultPerMinute > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rate limit must be between 1 and 1000"})
		return
	}

	// 验证邮件配置
	if err := h.systemSettingsService.ValidateEmailConfig(c.Request.Context(), &dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 OAuth2 配置
	if err := h.systemSettingsService.ValidateOAuth2Config(c.Request.Context(), &dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.systemSettingsService.UpdateSettings(c.Request.Context(), &dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// TestEmailConfiguration 测试邮件配置
func (h *AdminHandler) TestEmailConfiguration(c *gin.Context) {
	var req struct {
		TestEmail string `json:"test_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}

	// 测试邮件发送
	if err := h.systemSettingsService.TestEmailConfiguration(c.Request.Context(), req.TestEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send test email: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully"})
}

// ─── Provider handlers ────────────────────────────────────────────────────────

// ListProviders lists all AI providers
func (h *AdminHandler) ListProviders(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	providers, err := h.adminService.ListProviders(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if providers == nil {
		providers = []*model.AIProvider{}
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// CreateProvider creates a new AI provider
func (h *AdminHandler) CreateProvider(c *gin.Context) {
	adminUserID := h.getAdminUserID(c)
	if adminUserID == uuid.Nil {
		return
	}

	var req model.AIProviderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	provider, err := h.adminService.CreateProvider(c.Request.Context(), adminUserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// UpdateProvider updates an AI provider
func (h *AdminHandler) UpdateProvider(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var req model.AIProviderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	provider, err := h.adminService.UpdateProvider(c.Request.Context(), providerID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProvider deletes an AI provider
func (h *AdminHandler) DeleteProvider(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	if err := h.adminService.DeleteProvider(c.Request.Context(), providerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}
