package model

import (
	"time"

	"github.com/google/uuid"
)

// AIProvider represents an AI API provider configuration
type AIProvider struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	DisplayName     string     `json:"display_name" db:"display_name"`
	ProviderType    string     `json:"provider_type" db:"provider_type"` // openai | anthropic | custom
	APIEndpoint     string     `json:"api_endpoint" db:"api_endpoint"`
	APIKeyEncrypted string     `json:"-" db:"api_key_encrypted"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	Description     string     `json:"description,omitempty" db:"description"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	// Computed
	ModelCount int `json:"model_count,omitempty" db:"-"`
}

// AIProviderCreateRequest represents request to create a provider
type AIProviderCreateRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	DisplayName  string `json:"display_name" binding:"required,min=1,max=100"`
	ProviderType string `json:"provider_type" binding:"required"`
	APIEndpoint  string `json:"api_endpoint" binding:"required"`
	APIKey       string `json:"api_key" binding:"required"`
	Description  string `json:"description"`
}

// AIProviderUpdateRequest represents request to update a provider
type AIProviderUpdateRequest struct {
	DisplayName  *string `json:"display_name" binding:"omitempty,min=1,max=100"`
	ProviderType *string `json:"provider_type"`
	APIEndpoint  *string `json:"api_endpoint"`
	APIKey       *string `json:"api_key"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
}
