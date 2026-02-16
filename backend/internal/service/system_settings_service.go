package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/repository"
)

// SystemSettingsService handles system settings business logic
type SystemSettingsService struct {
	settingsRepo *repository.SystemSettingsRepository
}

// NewSystemSettingsService creates a new system settings service
func NewSystemSettingsService(settingsRepo *repository.SystemSettingsRepository) *SystemSettingsService {
	return &SystemSettingsService{
		settingsRepo: settingsRepo,
	}
}

// GetSettings retrieves all system settings as DTO
func (s *SystemSettingsService) GetSettings(ctx context.Context) (*model.SystemSettingsDTO, error) {
	settings, err := s.settingsRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	dto := &model.SystemSettingsDTO{}

	for _, setting := range settings {
		switch setting.SettingKey {
		case "rate_limit_default_per_minute":
			val, _ := strconv.Atoi(setting.SettingValue)
			dto.RateLimitDefaultPerMinute = val
		case "system_name":
			dto.SystemName = setting.SettingValue
		case "maintenance_mode":
			dto.MaintenanceMode = setting.SettingValue == "true"
		}
	}

	return dto, nil
}

// UpdateSettings updates system settings
func (s *SystemSettingsService) UpdateSettings(ctx context.Context, dto *model.SystemSettingsDTO) error {
	updates := make(map[string]string)

	if dto.RateLimitDefaultPerMinute > 0 {
		updates["rate_limit_default_per_minute"] = strconv.Itoa(dto.RateLimitDefaultPerMinute)
	}

	if dto.SystemName != "" {
		updates["system_name"] = dto.SystemName
	}

	updates["maintenance_mode"] = strconv.FormatBool(dto.MaintenanceMode)

	return s.settingsRepo.UpdateMultiple(ctx, updates)
}

// GetRateLimitDefault retrieves the default rate limit setting
func (s *SystemSettingsService) GetRateLimitDefault(ctx context.Context) (int, error) {
	setting, err := s.settingsRepo.GetByKey(ctx, "rate_limit_default_per_minute")
	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit default: %w", err)
	}

	val, err := strconv.Atoi(setting.SettingValue)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit value: %w", err)
	}

	return val, nil
}
