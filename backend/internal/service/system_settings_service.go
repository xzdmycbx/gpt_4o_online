package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/crypto"
	"github.com/ai-chat/backend/internal/pkg/email"
	"github.com/ai-chat/backend/internal/repository"
)

// SystemSettingsService handles system settings business logic
type SystemSettingsService struct {
	settingsRepo  *repository.SystemSettingsRepository
	encryptionKey string
}

// NewSystemSettingsService creates a new system settings service
func NewSystemSettingsService(settingsRepo *repository.SystemSettingsRepository, encryptionKey string) *SystemSettingsService {
	return &SystemSettingsService{
		settingsRepo:  settingsRepo,
		encryptionKey: encryptionKey,
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
		// 基础设置
		case "rate_limit_default_per_minute":
			val, _ := strconv.Atoi(setting.SettingValue)
			dto.RateLimitDefaultPerMinute = val
		case "system_name":
			dto.SystemName = setting.SettingValue
		case "maintenance_mode":
			dto.MaintenanceMode = setting.SettingValue == "true"

		// OAuth2 配置
		case "oauth2_twitter_enabled":
			dto.OAuth2TwitterEnabled = setting.SettingValue == "true"
		case "oauth2_twitter_client_id":
			dto.OAuth2TwitterClientID = setting.SettingValue
		case "oauth2_twitter_client_secret":
			// 解密敏感信息
			if setting.SettingValue != "" {
				decrypted, err := crypto.Decrypt(setting.SettingValue, s.encryptionKey)
				if err == nil {
					dto.OAuth2TwitterClientSecret = decrypted
				}
			}
		case "oauth2_twitter_redirect_url":
			dto.OAuth2TwitterRedirectURL = setting.SettingValue

		// 邮件配置
		case "email_enabled":
			dto.EmailEnabled = setting.SettingValue == "true"
		case "email_provider":
			dto.EmailProvider = setting.SettingValue
		case "email_smtp_host":
			dto.EmailSMTPHost = setting.SettingValue
		case "email_smtp_port":
			val, _ := strconv.Atoi(setting.SettingValue)
			dto.EmailSMTPPort = val
		case "email_smtp_user":
			dto.EmailSMTPUser = setting.SettingValue
		case "email_smtp_password":
			// 解密敏感信息
			if setting.SettingValue != "" {
				decrypted, err := crypto.Decrypt(setting.SettingValue, s.encryptionKey)
				if err == nil {
					dto.EmailSMTPPassword = decrypted
				}
			}
		case "email_from":
			dto.EmailFrom = setting.SettingValue
		case "email_from_name":
			dto.EmailFromName = setting.SettingValue
		case "email_resend_api_key":
			// 解密敏感信息
			if setting.SettingValue != "" {
				decrypted, err := crypto.Decrypt(setting.SettingValue, s.encryptionKey)
				if err == nil {
					dto.EmailResendAPIKey = decrypted
				}
			}

		// AI 配置
		case "ai_default_memory_model":
			dto.AIDefaultMemoryModel = setting.SettingValue
		case "ai_memory_extraction_enabled":
			dto.AIMemoryExtractionEnabled = setting.SettingValue == "true"
		}
	}

	return dto, nil
}

// UpdateSettings updates system settings
func (s *SystemSettingsService) UpdateSettings(ctx context.Context, dto *model.SystemSettingsDTO) error {
	updates := make(map[string]string)

	// 基础设置
	if dto.RateLimitDefaultPerMinute > 0 {
		updates["rate_limit_default_per_minute"] = strconv.Itoa(dto.RateLimitDefaultPerMinute)
	}
	if dto.SystemName != "" {
		updates["system_name"] = dto.SystemName
	}
	updates["maintenance_mode"] = strconv.FormatBool(dto.MaintenanceMode)

	// OAuth2 配置
	updates["oauth2_twitter_enabled"] = strconv.FormatBool(dto.OAuth2TwitterEnabled)
	if dto.OAuth2TwitterClientID != "" {
		updates["oauth2_twitter_client_id"] = dto.OAuth2TwitterClientID
	}
	// Client Secret - 只在不为空且不是掩码时才更新
	if dto.OAuth2TwitterClientSecret != "" && !containsMask(dto.OAuth2TwitterClientSecret) {
		encrypted, err := crypto.Encrypt(dto.OAuth2TwitterClientSecret, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt oauth2 client secret: %w", err)
		}
		updates["oauth2_twitter_client_secret"] = encrypted
	}
	if dto.OAuth2TwitterRedirectURL != "" {
		updates["oauth2_twitter_redirect_url"] = dto.OAuth2TwitterRedirectURL
	}

	// 邮件配置
	updates["email_enabled"] = strconv.FormatBool(dto.EmailEnabled)
	if dto.EmailProvider != "" {
		updates["email_provider"] = dto.EmailProvider
	}
	if dto.EmailSMTPHost != "" {
		updates["email_smtp_host"] = dto.EmailSMTPHost
	}
	if dto.EmailSMTPPort > 0 {
		updates["email_smtp_port"] = strconv.Itoa(dto.EmailSMTPPort)
	}
	if dto.EmailSMTPUser != "" {
		updates["email_smtp_user"] = dto.EmailSMTPUser
	}
	// SMTP Password - 只在不为空且不是掩码时才更新
	if dto.EmailSMTPPassword != "" && !containsMask(dto.EmailSMTPPassword) {
		encrypted, err := crypto.Encrypt(dto.EmailSMTPPassword, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt email password: %w", err)
		}
		updates["email_smtp_password"] = encrypted
	}
	if dto.EmailFrom != "" {
		updates["email_from"] = dto.EmailFrom
	}
	if dto.EmailFromName != "" {
		updates["email_from_name"] = dto.EmailFromName
	}
	// Resend API Key - 只在不为空且不是掩码时才更新
	if dto.EmailResendAPIKey != "" && !containsMask(dto.EmailResendAPIKey) {
		encrypted, err := crypto.Encrypt(dto.EmailResendAPIKey, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt resend api key: %w", err)
		}
		updates["email_resend_api_key"] = encrypted
	}

	// AI 配置
	if dto.AIDefaultMemoryModel != "" {
		updates["ai_default_memory_model"] = dto.AIDefaultMemoryModel
	}
	updates["ai_memory_extraction_enabled"] = strconv.FormatBool(dto.AIMemoryExtractionEnabled)

	return s.settingsRepo.UpdateMultiple(ctx, updates)
}

// containsMask 检查字符串是否包含掩码
func containsMask(s string) bool {
	return len(s) >= 8 && s[:8] == "********"
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

// TestEmailConfiguration 测试邮件配置
func (s *SystemSettingsService) TestEmailConfiguration(ctx context.Context, testEmail string) error {
	// 获取当前邮件配置
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	if !settings.EmailEnabled {
		return fmt.Errorf("email service is not enabled")
	}

	// 创建邮件发送器
	var sender *email.Sender
	if settings.EmailProvider == "resend" {
		sender = email.NewSender(
			"resend",
			"", 0, "", "",
			settings.EmailFrom,
			settings.EmailFromName,
			settings.EmailResendAPIKey,
		)
	} else {
		sender = email.NewSender(
			"smtp",
			settings.EmailSMTPHost,
			settings.EmailSMTPPort,
			settings.EmailSMTPUser,
			settings.EmailSMTPPassword,
			settings.EmailFrom,
			settings.EmailFromName,
			"",
		)
	}

	// 发送测试邮件
	subject := "Test Email - AI Chat System"
	htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Test Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2b5278;">邮件配置测试</h2>
        <p>这是一封测试邮件，用于验证您的邮件配置是否正确。</p>
        <p>如果您收到这封邮件，说明邮件服务配置成功！</p>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">
        <p style="color: #666; font-size: 12px;">
            AI Chat System - 系统配置测试<br>
            这是一封自动发送的测试邮件
        </p>
    </div>
</body>
</html>
`
	textBody := `
邮件配置测试

这是一封测试邮件，用于验证您的邮件配置是否正确。
如果您收到这封邮件，说明邮件服务配置成功！

---
AI Chat System - 系统配置测试
这是一封自动发送的测试邮件
`

	return sender.SendEmail(testEmail, subject, htmlBody, textBody)
}

// ValidateEmailConfig validates email configuration.
func (s *SystemSettingsService) ValidateEmailConfig(ctx context.Context, dto *model.SystemSettingsDTO) error {
	if !dto.EmailEnabled {
		return nil
	}

	if dto.EmailFrom == "" {
		return fmt.Errorf("email_from is required")
	}

	if dto.EmailProvider != "smtp" && dto.EmailProvider != "resend" {
		return fmt.Errorf("invalid email_provider: must be 'smtp' or 'resend'")
	}

	if dto.EmailProvider == "smtp" {
		if dto.EmailSMTPHost == "" {
			return fmt.Errorf("email_smtp_host is required for SMTP provider")
		}
		if dto.EmailSMTPPort <= 0 {
			return fmt.Errorf("email_smtp_port must be greater than 0")
		}
	}

	if dto.EmailProvider == "resend" && (dto.EmailResendAPIKey == "" || containsMask(dto.EmailResendAPIKey)) {
		hasStoredKey, err := s.settingsRepo.HasNonEmptyValue(ctx, "email_resend_api_key")
		if err != nil {
			return fmt.Errorf("failed to validate existing resend api key: %w", err)
		}
		if !hasStoredKey {
			return fmt.Errorf("email_resend_api_key is required for Resend provider")
		}
	}

	return nil
}

// ValidateOAuth2Config validates OAuth2 configuration.
func (s *SystemSettingsService) ValidateOAuth2Config(ctx context.Context, dto *model.SystemSettingsDTO) error {
	if !dto.OAuth2TwitterEnabled {
		return nil
	}

	if dto.OAuth2TwitterClientID == "" {
		return fmt.Errorf("oauth2_twitter_client_id is required")
	}

	if dto.OAuth2TwitterClientSecret == "" || containsMask(dto.OAuth2TwitterClientSecret) {
		hasStoredSecret, err := s.settingsRepo.HasNonEmptyValue(ctx, "oauth2_twitter_client_secret")
		if err != nil {
			return fmt.Errorf("failed to validate existing oauth2 client secret: %w", err)
		}
		if !hasStoredSecret {
			return fmt.Errorf("oauth2_twitter_client_secret is required")
		}
	}

	if dto.OAuth2TwitterRedirectURL == "" {
		return fmt.Errorf("oauth2_twitter_redirect_url is required")
	}

	return nil
}
