package model

import (
	"time"

	"github.com/google/uuid"
)

// MemoryCategory represents memory classification
type MemoryCategory string

const (
	MemoryCategoryPreference MemoryCategory = "preference"
	MemoryCategoryFact       MemoryCategory = "fact"
	MemoryCategoryContext    MemoryCategory = "context"
)

// Memory represents extracted user memory
type Memory struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	UserID               uuid.UUID       `json:"user_id" db:"user_id"`
	Content              string          `json:"content" db:"content"`
	Category             MemoryCategory  `json:"category" db:"category"`
	Importance           int             `json:"importance" db:"importance"` // 1-10
	SourceConversationID *uuid.UUID      `json:"source_conversation_id,omitempty" db:"source_conversation_id"`
	SourceMessageID      *uuid.UUID      `json:"source_message_id,omitempty" db:"source_message_id"`
	TimesUsed            int             `json:"times_used" db:"times_used"`
	LastUsedAt           *time.Time      `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// MemoryCreateRequest represents request to create memory
type MemoryCreateRequest struct {
	Content    string         `json:"content" binding:"required,min=1,max=500"`
	Category   MemoryCategory `json:"category" binding:"required,oneof=preference fact context"`
	Importance int            `json:"importance" binding:"required,min=1,max=10"`
}

// MemoryUpdateRequest represents request to update memory
type MemoryUpdateRequest struct {
	Content    *string         `json:"content" binding:"omitempty,min=1,max=500"`
	Category   *MemoryCategory `json:"category" binding:"omitempty,oneof=preference fact context"`
	Importance *int            `json:"importance" binding:"omitempty,min=1,max=10"`
}

// MemoryExtractionResult represents extracted memories from conversation
type MemoryExtractionResult struct {
	Content    string         `json:"content"`
	Category   MemoryCategory `json:"category"`
	Importance int            `json:"importance"`
}
