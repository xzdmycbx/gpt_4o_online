package model

import (
	"time"

	"github.com/google/uuid"
)

// SystemSetting represents a system-wide configuration setting
type SystemSetting struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SettingKey   string    `json:"setting_key" db:"setting_key"`
	SettingValue string    `json:"setting_value" db:"setting_value"`
	Description  *string   `json:"description,omitempty" db:"description"`
	ValueType    string    `json:"value_type" db:"value_type"` // string, int, bool, json
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// SystemSettingsDTO for API responses
type SystemSettingsDTO struct {
	// 基础设置
	RateLimitDefaultPerMinute int    `json:"rate_limit_default_per_minute"`
	SystemName                string `json:"system_name"`
	MaintenanceMode           bool   `json:"maintenance_mode"`

	// OAuth2 配置
	OAuth2TwitterEnabled      bool   `json:"oauth2_twitter_enabled"`
	OAuth2TwitterClientID     string `json:"oauth2_twitter_client_id"`
	OAuth2TwitterClientSecret string `json:"oauth2_twitter_client_secret"`
	OAuth2TwitterRedirectURL  string `json:"oauth2_twitter_redirect_url"`

	// 邮件配置
	EmailEnabled      bool   `json:"email_enabled"`
	EmailProvider     string `json:"email_provider"`
	EmailSMTPHost     string `json:"email_smtp_host"`
	EmailSMTPPort     int    `json:"email_smtp_port"`
	EmailSMTPUser     string `json:"email_smtp_user"`
	EmailSMTPPassword string `json:"email_smtp_password"`
	EmailFrom         string `json:"email_from"`
	EmailFromName     string `json:"email_from_name"`
	EmailResendAPIKey string `json:"email_resend_api_key"`

	// AI 配置
	AIDefaultMemoryModel      string `json:"ai_default_memory_model"`
	AIMemoryExtractionEnabled bool   `json:"ai_memory_extraction_enabled"`
}

// MaskSensitiveData 掩码敏感信息，用于API返回
func (dto *SystemSettingsDTO) MaskSensitiveData() {
	// OAuth2 Client Secret - 保留最后4位
	if dto.OAuth2TwitterClientSecret != "" {
		if len(dto.OAuth2TwitterClientSecret) > 4 {
			dto.OAuth2TwitterClientSecret = "********" + dto.OAuth2TwitterClientSecret[len(dto.OAuth2TwitterClientSecret)-4:]
		} else {
			dto.OAuth2TwitterClientSecret = "********"
		}
	}

	// SMTP 密码 - 完全隐藏
	if dto.EmailSMTPPassword != "" {
		dto.EmailSMTPPassword = "********"
	}

	// Resend API Key - 保留最后4位
	if dto.EmailResendAPIKey != "" {
		if len(dto.EmailResendAPIKey) > 4 {
			dto.EmailResendAPIKey = "********" + dto.EmailResendAPIKey[len(dto.EmailResendAPIKey)-4:]
		} else {
			dto.EmailResendAPIKey = "********"
		}
	}
}
