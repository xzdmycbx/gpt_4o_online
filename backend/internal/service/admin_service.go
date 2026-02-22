package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/crypto"
	"github.com/ai-chat/backend/internal/repository"
)

// AdminService handles admin operations
type AdminService struct {
	userRepo         *repository.UserRepository
	modelRepo        *repository.AIModelRepository
	providerRepo     *repository.AIProviderRepository
	auditRepo        *repository.AuditLogRepository
	tokenUsageRepo   *repository.TokenUsageRepository
	conversationRepo *repository.ConversationRepository
	messageRepo      *repository.MessageRepository
	encryptionKey    string
	startTime        time.Time // Track server start time
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo *repository.UserRepository,
	modelRepo *repository.AIModelRepository,
	providerRepo *repository.AIProviderRepository,
	auditRepo *repository.AuditLogRepository,
	tokenUsageRepo *repository.TokenUsageRepository,
	conversationRepo *repository.ConversationRepository,
	messageRepo *repository.MessageRepository,
	encryptionKey string,
) *AdminService {
	return &AdminService{
		userRepo:         userRepo,
		modelRepo:        modelRepo,
		providerRepo:     providerRepo,
		auditRepo:        auditRepo,
		tokenUsageRepo:   tokenUsageRepo,
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		encryptionKey:    encryptionKey,
		startTime:        time.Now(),
	}
}

// ListUsers lists all users with pagination
func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, int64, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUser retrieves a user by ID
func (s *AdminService) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateUser updates a user
func (s *AdminService) UpdateUser(ctx context.Context, adminUser *model.User, targetUserID uuid.UUID, req *model.UserUpdateRequest) (*model.User, error) {
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if admin can manage this user
	if !adminUser.CanManageUser(targetUser) {
		return nil, fmt.Errorf("insufficient permissions to manage this user")
	}

	// Update allowed fields
	if req.Email != nil {
		targetUser.Email = req.Email
	}
	if req.DisplayName != nil {
		targetUser.DisplayName = req.DisplayName
	}
	if req.AvatarURL != nil {
		targetUser.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, targetUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return targetUser, nil
}

// BanUser bans a user
func (s *AdminService) BanUser(ctx context.Context, adminUser *model.User, targetUserID uuid.UUID, reason string) error {
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if admin can manage this user
	if !adminUser.CanManageUser(targetUser) {
		return fmt.Errorf("insufficient permissions to ban this user")
	}

	now := time.Now()
	targetUser.IsBanned = true
	targetUser.BanReason = &reason
	targetUser.BannedAt = &now
	targetUser.BannedBy = &adminUser.ID

	return s.userRepo.Update(ctx, targetUser)
}

// UnbanUser unbans a user
func (s *AdminService) UnbanUser(ctx context.Context, adminUser *model.User, targetUserID uuid.UUID) error {
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !adminUser.CanManageUser(targetUser) {
		return fmt.Errorf("insufficient permissions to unban this user")
	}

	targetUser.IsBanned = false
	targetUser.BanReason = nil
	targetUser.BannedAt = nil
	targetUser.BannedBy = nil

	return s.userRepo.Update(ctx, targetUser)
}

// DeleteUser deletes a user
func (s *AdminService) DeleteUser(ctx context.Context, adminUser *model.User, targetUserID uuid.UUID) error {
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !adminUser.CanManageUser(targetUser) {
		return fmt.Errorf("insufficient permissions to delete this user")
	}

	return s.userRepo.Delete(ctx, targetUserID)
}

// ChangeUserRole changes a user's role (super admin only)
func (s *AdminService) ChangeUserRole(ctx context.Context, adminUser *model.User, targetUserID uuid.UUID, newRole model.UserRole) error {
	// Only super admin can change roles
	if !adminUser.IsSuperAdmin() {
		return fmt.Errorf("only super admin can change user roles")
	}

	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Cannot set super admin role through API
	if newRole == model.RoleSuperAdmin {
		return fmt.Errorf("cannot set super admin role through API")
	}

	targetUser.Role = newRole
	return s.userRepo.Update(ctx, targetUser)
}

// ListAIModels lists all AI models
func (s *AdminService) ListAIModels(ctx context.Context, activeOnly bool) ([]*model.AIModel, error) {
	return s.modelRepo.List(ctx, activeOnly)
}

// CreateAIModel creates a new AI model
func (s *AdminService) CreateAIModel(ctx context.Context, adminUserID uuid.UUID, req *model.AIModelCreateRequest) (*model.AIModel, error) {
	var encryptedKey string

	// If no provider_id, API key is required
	if req.ProviderID == nil {
		if req.APIKey == "" {
			return nil, fmt.Errorf("api_key is required when no provider is specified")
		}
		if req.APIEndpoint == "" {
			return nil, fmt.Errorf("api_endpoint is required when no provider is specified")
		}
		var err error
		encryptedKey, err = crypto.Encrypt(req.APIKey, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
	} else if req.APIKey != "" {
		// Optional: model-level key override
		var err error
		encryptedKey, err = crypto.Encrypt(req.APIKey, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
	}

	aiModel := &model.AIModel{
		ID:                uuid.New(),
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Provider:          req.Provider,
		APIEndpoint:       req.APIEndpoint,
		APIKeyEncrypted:   encryptedKey,
		ModelIdentifier:   req.ModelIdentifier,
		ProviderID:        req.ProviderID,
		SupportsStreaming: req.SupportsStreaming,
		SupportsFunctions: req.SupportsFunctions,
		MaxTokens:         req.MaxTokens,
		InputPricePer1k:   req.InputPricePer1k,
		OutputPricePer1k:  req.OutputPricePer1k,
		IsActive:          true,
		IsDefault:         false,
		Description:       req.Description,
		CreatedBy:         &adminUserID,
	}

	if err := s.modelRepo.Create(ctx, aiModel); err != nil {
		return nil, fmt.Errorf("failed to create AI model: %w", err)
	}

	return aiModel, nil
}

// UpdateAIModel updates an AI model
func (s *AdminService) UpdateAIModel(ctx context.Context, modelID uuid.UUID, req *model.AIModelUpdateRequest) (*model.AIModel, error) {
	aiModel, err := s.modelRepo.GetByID(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("model not found")
	}

	// Update fields
	if req.DisplayName != nil {
		aiModel.DisplayName = *req.DisplayName
	}
	if req.APIEndpoint != nil {
		aiModel.APIEndpoint = *req.APIEndpoint
	}
	if req.APIKey != nil {
		encryptedKey, err := crypto.Encrypt(*req.APIKey, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		aiModel.APIKeyEncrypted = encryptedKey
	}
	if req.SupportsStreaming != nil {
		aiModel.SupportsStreaming = *req.SupportsStreaming
	}
	if req.SupportsFunctions != nil {
		aiModel.SupportsFunctions = *req.SupportsFunctions
	}
	if req.MaxTokens != nil {
		aiModel.MaxTokens = *req.MaxTokens
	}
	if req.InputPricePer1k != nil {
		aiModel.InputPricePer1k = req.InputPricePer1k
	}
	if req.OutputPricePer1k != nil {
		aiModel.OutputPricePer1k = req.OutputPricePer1k
	}
	if req.Description != nil {
		aiModel.Description = *req.Description
	}
	if req.IsActive != nil {
		aiModel.IsActive = *req.IsActive
	}

	if err := s.modelRepo.Update(ctx, aiModel); err != nil {
		return nil, fmt.Errorf("failed to update AI model: %w", err)
	}

	return aiModel, nil
}

// DeleteAIModel deletes an AI model
func (s *AdminService) DeleteAIModel(ctx context.Context, modelID uuid.UUID) error {
	return s.modelRepo.Delete(ctx, modelID)
}

// SetDefaultModel sets a model as default
func (s *AdminService) SetDefaultModel(ctx context.Context, modelID uuid.UUID) error {
	return s.modelRepo.SetDefault(ctx, modelID)
}

// ─── Provider CRUD ────────────────────────────────────────────────────────────

// ListProviders lists all AI providers
func (s *AdminService) ListProviders(ctx context.Context, activeOnly bool) ([]*model.AIProvider, error) {
	return s.providerRepo.List(ctx, activeOnly)
}

// CreateProvider creates a new AI provider
func (s *AdminService) CreateProvider(ctx context.Context, adminUserID uuid.UUID, req *model.AIProviderCreateRequest) (*model.AIProvider, error) {
	encryptedKey, err := crypto.Encrypt(req.APIKey, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %w", err)
	}

	provider := &model.AIProvider{
		ID:              uuid.New(),
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		ProviderType:    req.ProviderType,
		APIEndpoint:     req.APIEndpoint,
		APIKeyEncrypted: encryptedKey,
		IsActive:        true,
		Description:     req.Description,
		CreatedBy:       &adminUserID,
	}

	if err := s.providerRepo.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// UpdateProvider updates an AI provider
func (s *AdminService) UpdateProvider(ctx context.Context, providerID uuid.UUID, req *model.AIProviderUpdateRequest) (*model.AIProvider, error) {
	provider, err := s.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}

	if req.DisplayName != nil {
		provider.DisplayName = *req.DisplayName
	}
	if req.ProviderType != nil {
		provider.ProviderType = *req.ProviderType
	}
	if req.APIEndpoint != nil {
		provider.APIEndpoint = *req.APIEndpoint
	}
	if req.APIKey != nil && *req.APIKey != "" {
		encryptedKey, err := crypto.Encrypt(*req.APIKey, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		provider.APIKeyEncrypted = encryptedKey
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}
	if req.Description != nil {
		provider.Description = *req.Description
	}

	if err := s.providerRepo.Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	return provider, nil
}

// DeleteProvider deletes an AI provider (only if no models are linked)
func (s *AdminService) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	count, err := s.providerRepo.CountModels(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to check linked models: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete provider: %d model(s) are linked to it", count)
	}

	return s.providerRepo.Delete(ctx, providerID)
}

// GetTokenLeaderboard retrieves the token usage leaderboard
func (s *AdminService) GetTokenLeaderboard(ctx context.Context, limit int) ([]*model.TokenLeaderboard, error) {
	return s.tokenUsageRepo.GetLeaderboard(ctx, limit)
}

// GetSystemOverview retrieves system statistics
func (s *AdminService) GetSystemOverview(ctx context.Context) (map[string]interface{}, error) {
	// Get total users
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get total conversations
	totalConversations, err := s.conversationRepo.Count(ctx)
	if err != nil {
		totalConversations = 0 // Non-critical, continue
	}

	// Get total messages
	totalMessages, err := s.messageRepo.Count(ctx)
	if err != nil {
		totalMessages = 0
	}

	// Get total tokens used (sum of all token usage)
	totalTokensUsed, err := s.tokenUsageRepo.GetTotalTokens(ctx)
	if err != nil {
		totalTokensUsed = 0
	}

	// Get active users today
	activeUsersToday, err := s.userRepo.CountActiveUsers(ctx, 1)
	if err != nil {
		activeUsersToday = 0
	}

	// Get active users this week
	activeUsersWeek, err := s.userRepo.CountActiveUsers(ctx, 7)
	if err != nil {
		activeUsersWeek = 0
	}

	// Get messages today
	messagesToday, err := s.messageRepo.CountToday(ctx)
	if err != nil {
		messagesToday = 0
	}

	// Get messages this week
	messagesWeek, err := s.messageRepo.CountLastDays(ctx, 7)
	if err != nil {
		messagesWeek = 0
	}

	// Calculate average tokens per message
	averageTokensPerMessage := float64(0)
	if totalMessages > 0 {
		averageTokensPerMessage = float64(totalTokensUsed) / float64(totalMessages)
	}

	// Calculate system uptime
	uptime := time.Since(s.startTime)
	systemUptime := formatUptime(uptime)

	return map[string]interface{}{
		"total_users":               totalUsers,
		"total_conversations":       totalConversations,
		"total_messages":            totalMessages,
		"total_tokens_used":         totalTokensUsed,
		"active_users_today":        activeUsersToday,
		"active_users_week":         activeUsersWeek,
		"messages_today":            messagesToday,
		"messages_week":             messagesWeek,
		"average_tokens_per_message": averageTokensPerMessage,
		"system_uptime":             systemUptime,
	}, nil
}

// formatUptime formats duration into human-readable string
func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天 %d小时 %d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时 %d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}

// ListAuditLogs lists audit logs
func (s *AdminService) ListAuditLogs(ctx context.Context, limit, offset int) ([]*model.AuditLog, error) {
	return s.auditRepo.List(ctx, limit, offset)
}

// SetUserRateLimit sets custom rate limit for a user
func (s *AdminService) SetUserRateLimit(ctx context.Context, adminUser *model.User, userID uuid.UUID, limit *int, exempt bool) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if admin can manage this user (prevents regular admins from modifying other admins)
	if !adminUser.CanManageUser(user) {
		return fmt.Errorf("insufficient permissions to manage this user")
	}

	user.CustomRateLimit = limit
	user.RateLimitExempt = exempt

	return s.userRepo.Update(ctx, user)
}
