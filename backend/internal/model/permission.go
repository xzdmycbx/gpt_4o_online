package model

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents system permissions
type Permission string

const (
	PermManageUsers        Permission = "manage_users"
	PermManageAdmins       Permission = "manage_admins"
	PermManageModels       Permission = "manage_models"
	PermManageSettings     Permission = "manage_settings"
	PermViewAuditLogs      Permission = "view_audit_logs"
	PermViewStatistics     Permission = "view_statistics"
	PermViewConversations  Permission = "view_conversations"
	PermViewMemories       Permission = "view_memories"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[UserRole][]Permission{
	RoleSuperAdmin: {
		PermManageUsers,
		PermManageAdmins,
		PermManageModels,
		PermManageSettings,
		PermViewAuditLogs,
		PermViewStatistics,
		PermViewConversations,
		PermViewMemories,
	},
	RoleAdmin: {
		PermManageUsers,
		PermManageModels,
		PermManageSettings,
		PermViewAuditLogs,
		PermViewStatistics,
		// Note: NO PermViewConversations or PermViewMemories
	},
	RoleUser: {},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role UserRole, perm Permission) bool {
	perms, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// AIModel represents an AI model configuration
type AIModel struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	DisplayName      string    `json:"display_name" db:"display_name"`
	Provider         string    `json:"provider" db:"provider"`
	APIEndpoint      string    `json:"api_endpoint" db:"api_endpoint"`
	APIKeyEncrypted  string    `json:"-" db:"api_key_encrypted"`
	ModelIdentifier  string    `json:"model_identifier" db:"model_identifier"`

	// Capabilities
	SupportsStreaming  bool `json:"supports_streaming" db:"supports_streaming"`
	SupportsFunctions  bool `json:"supports_functions" db:"supports_functions"`
	MaxTokens          int  `json:"max_tokens" db:"max_tokens"`

	// Pricing
	InputPricePer1k  *float64 `json:"input_price_per_1k,omitempty" db:"input_price_per_1k"`
	OutputPricePer1k *float64 `json:"output_price_per_1k,omitempty" db:"output_price_per_1k"`

	// Status
	IsActive  bool `json:"is_active" db:"is_active"`
	IsDefault bool `json:"is_default" db:"is_default"`

	// Metadata
	Description string     `json:"description,omitempty" db:"description"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// AIModelCreateRequest represents request to create an AI model
type AIModelCreateRequest struct {
	Name             string   `json:"name" binding:"required,min=1,max=100"`
	DisplayName      string   `json:"display_name" binding:"required,min=1,max=100"`
	Provider         string   `json:"provider" binding:"required"`
	APIEndpoint      string   `json:"api_endpoint" binding:"required,url"`
	APIKey           string   `json:"api_key" binding:"required"`
	ModelIdentifier  string   `json:"model_identifier" binding:"required"`
	SupportsStreaming bool    `json:"supports_streaming"`
	SupportsFunctions bool    `json:"supports_functions"`
	MaxTokens        int     `json:"max_tokens" binding:"required,min=1"`
	InputPricePer1k  *float64 `json:"input_price_per_1k"`
	OutputPricePer1k *float64 `json:"output_price_per_1k"`
	Description      string  `json:"description"`
}

// AIModelUpdateRequest represents request to update an AI model
type AIModelUpdateRequest struct {
	DisplayName      *string  `json:"display_name" binding:"omitempty,min=1,max=100"`
	APIEndpoint      *string  `json:"api_endpoint" binding:"omitempty,url"`
	APIKey           *string  `json:"api_key"`
	SupportsStreaming *bool   `json:"supports_streaming"`
	SupportsFunctions *bool   `json:"supports_functions"`
	MaxTokens        *int     `json:"max_tokens" binding:"omitempty,min=1"`
	InputPricePer1k  *float64 `json:"input_price_per_1k"`
	OutputPricePer1k *float64 `json:"output_price_per_1k"`
	Description      *string  `json:"description"`
	IsActive         *bool    `json:"is_active"`
}
