package model

import (
	"time"

	"github.com/google/uuid"
)

// UserSettings represents user preferences and settings
type UserSettings struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`

	// UI Settings
	Theme    string `json:"theme" db:"theme"` // "dark", "light", "auto"
	FontSize string `json:"font_size" db:"font_size"` // "small", "medium", "large"
	Language string `json:"language" db:"language"`

	// Notification Settings
	NotificationsEnabled bool `json:"notifications_enabled" db:"notifications_enabled"`
	NotificationSound    bool `json:"notification_sound" db:"notification_sound"`

	// Chat Preferences
	DefaultModelID  *uuid.UUID `json:"default_model_id,omitempty" db:"default_model_id"`
	StreamResponse  bool       `json:"stream_response" db:"stream_response"`
	ShowTokenCount  bool       `json:"show_token_count" db:"show_token_count"`

	// Advanced Settings (JSONB)
	AdvancedSettings map[string]interface{} `json:"advanced_settings,omitempty" db:"advanced_settings"`

	// Sync metadata
	DeviceID     *string    `json:"device_id,omitempty" db:"device_id"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserSettingsUpdateRequest represents request to update settings
type UserSettingsUpdateRequest struct {
	Theme                *string                 `json:"theme" binding:"omitempty,oneof=dark light auto"`
	FontSize             *string                 `json:"font_size" binding:"omitempty,oneof=small medium large"`
	Language             *string                 `json:"language"`
	NotificationsEnabled *bool                   `json:"notifications_enabled"`
	NotificationSound    *bool                   `json:"notification_sound"`
	DefaultModelID       *uuid.UUID              `json:"default_model_id"`
	StreamResponse       *bool                   `json:"stream_response"`
	ShowTokenCount       *bool                   `json:"show_token_count"`
	AdvancedSettings     map[string]interface{}  `json:"advanced_settings"`
	DeviceID             *string                 `json:"device_id"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	ModelID       *uuid.UUID `json:"model_id,omitempty" db:"model_id"`
	InputTokens   int        `json:"input_tokens" db:"input_tokens"`
	OutputTokens  int        `json:"output_tokens" db:"output_tokens"`
	TotalTokens   int        `json:"total_tokens" db:"total_tokens"`
	EstimatedCost *float64   `json:"estimated_cost,omitempty" db:"estimated_cost"`
	PeriodStart   time.Time  `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time  `json:"period_end" db:"period_end"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// TokenLeaderboard represents leaderboard entry
type TokenLeaderboard struct {
	UserID        uuid.UUID `json:"user_id" db:"id"`
	Username      string    `json:"username" db:"username"`
	DisplayName   *string   `json:"display_name,omitempty" db:"display_name"`
	AvatarURL     *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	TotalTokens   int       `json:"total_tokens" db:"total_tokens"`
	TotalRequests int       `json:"total_requests" db:"total_requests"`
	InputTokens   int       `json:"input_tokens" db:"input_tokens"`
	OutputTokens  int       `json:"output_tokens" db:"output_tokens"`
	EstimatedCost *float64  `json:"estimated_cost,omitempty" db:"estimated_cost"`
	Rank          int       `json:"rank" db:"rank"`
}

// AuditAction represents audit log action types
type AuditAction string

const (
	AuditUserCreated       AuditAction = "user_created"
	AuditUserUpdated       AuditAction = "user_updated"
	AuditUserBanned        AuditAction = "user_banned"
	AuditUserUnbanned      AuditAction = "user_unbanned"
	AuditModelCreated      AuditAction = "model_created"
	AuditModelUpdated      AuditAction = "model_updated"
	AuditModelDeleted      AuditAction = "model_deleted"
	AuditSettingsUpdated   AuditAction = "settings_updated"
	AuditPermissionChanged AuditAction = "permission_changed"
	AuditPasswordReset     AuditAction = "password_reset"
	AuditPasswordChange    AuditAction = "password_change"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	Action       AuditAction            `json:"action" db:"action"`
	ActorID      *uuid.UUID             `json:"actor_id,omitempty" db:"actor_id"`
	Username     string                 `json:"username" db:"username"` // Actor username
	ResourceType string                 `json:"resource_type" db:"resource_type"` // Type of resource affected
	TargetUserID *uuid.UUID             `json:"target_user_id,omitempty" db:"target_user_id"`
	Details      map[string]interface{} `json:"details,omitempty" db:"details"`
	IPAddress    *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// EmailProvider represents email service provider
type EmailProvider string

const (
	EmailProviderSMTP   EmailProvider = "smtp"
	EmailProviderResend EmailProvider = "resend"
)

// EmailConfig represents email configuration
type EmailConfig struct {
	ID                    int           `json:"id" db:"id"`
	Provider              EmailProvider `json:"provider" db:"provider"`
	SMTPHost              *string       `json:"smtp_host,omitempty" db:"smtp_host"`
	SMTPPort              *int          `json:"smtp_port,omitempty" db:"smtp_port"`
	SMTPUser              *string       `json:"smtp_user,omitempty" db:"smtp_user"`
	SMTPPasswordEncrypted *string       `json:"-" db:"smtp_password_encrypted"`
	SMTPUseTLS            *bool         `json:"smtp_use_tls,omitempty" db:"smtp_use_tls"`
	ResendAPIKeyEncrypted *string       `json:"-" db:"resend_api_key_encrypted"`
	FromEmail             string        `json:"from_email" db:"from_email"`
	FromName              *string       `json:"from_name,omitempty" db:"from_name"`
	IsActive              bool          `json:"is_active" db:"is_active"`
	LastTestedAt          *time.Time    `json:"last_tested_at,omitempty" db:"last_tested_at"`
	TestResult            *string       `json:"test_result,omitempty" db:"test_result"`
	CreatedAt             time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at" db:"updated_at"`
}
