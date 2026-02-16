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
	RateLimitDefaultPerMinute int    `json:"rate_limit_default_per_minute"`
	SystemName                string `json:"system_name"`
	MaintenanceMode           bool   `json:"maintenance_mode"`
}
