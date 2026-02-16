package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/repository"
)

// UserSettingsService handles user settings and multi-device sync
type UserSettingsService struct {
	settingsRepo *repository.UserSettingsRepository
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(settingsRepo *repository.UserSettingsRepository) *UserSettingsService {
	return &UserSettingsService{
		settingsRepo: settingsRepo,
	}
}

// GetSettings retrieves user settings
func (s *UserSettingsService) GetSettings(ctx context.Context, userID uuid.UUID) (*model.UserSettings, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If settings don't exist, create default settings
		defaultSettings := s.createDefaultSettings(userID)
		if createErr := s.settingsRepo.Upsert(ctx, defaultSettings); createErr != nil {
			return nil, fmt.Errorf("failed to create default settings: %w", createErr)
		}
		return defaultSettings, nil
	}

	return settings, nil
}

// UpdateSettings updates user settings
func (s *UserSettingsService) UpdateSettings(ctx context.Context, userID uuid.UUID, req *model.UserSettingsUpdateRequest) (*model.UserSettings, error) {
	// Get existing settings or create default
	settings, err := s.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Theme != nil {
		settings.Theme = *req.Theme
	}
	if req.FontSize != nil {
		settings.FontSize = *req.FontSize
	}
	if req.Language != nil {
		settings.Language = *req.Language
	}
	if req.NotificationsEnabled != nil {
		settings.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.NotificationSound != nil {
		settings.NotificationSound = *req.NotificationSound
	}
	if req.DefaultModelID != nil {
		settings.DefaultModelID = req.DefaultModelID
	}
	if req.StreamResponse != nil {
		settings.StreamResponse = *req.StreamResponse
	}
	if req.ShowTokenCount != nil {
		settings.ShowTokenCount = *req.ShowTokenCount
	}
	if req.AdvancedSettings != nil {
		settings.AdvancedSettings = req.AdvancedSettings
	}
	if req.DeviceID != nil {
		settings.DeviceID = req.DeviceID
	}

	// Upsert settings
	if err := s.settingsRepo.Upsert(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	return settings, nil
}

// SyncSettings syncs settings across devices
func (s *UserSettingsService) SyncSettings(ctx context.Context, userID uuid.UUID, localSettings *model.UserSettings) (*model.UserSettings, bool, error) {
	// Get server settings
	serverSettings, err := s.GetSettings(ctx, userID)
	if err != nil {
		return nil, false, err
	}

	// Compare timestamps to determine which is newer
	// If local is newer, push to server
	// If server is newer, return server settings to pull
	localTime := localSettings.UpdatedAt
	serverTime := serverSettings.UpdatedAt

	if localTime.After(serverTime) {
		// Local is newer, push to server
		localSettings.UserID = userID
		if err := s.settingsRepo.Upsert(ctx, localSettings); err != nil {
			return nil, false, fmt.Errorf("failed to sync settings: %w", err)
		}
		return localSettings, false, nil // false = pushed
	}

	// Server is newer or same, return server settings
	return serverSettings, true, nil // true = pulled
}

// DeleteSettings deletes user settings
func (s *UserSettingsService) DeleteSettings(ctx context.Context, userID uuid.UUID) error {
	return s.settingsRepo.Delete(ctx, userID)
}

// createDefaultSettings creates default settings for a new user
func (s *UserSettingsService) createDefaultSettings(userID uuid.UUID) *model.UserSettings {
	now := time.Now()
	return &model.UserSettings{
		ID:                   uuid.New(),
		UserID:               userID,
		Theme:                "dark",
		FontSize:             "medium",
		Language:             "en",
		NotificationsEnabled: true,
		NotificationSound:    true,
		StreamResponse:       true,
		ShowTokenCount:       false,
		AdvancedSettings:     make(map[string]interface{}),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}
