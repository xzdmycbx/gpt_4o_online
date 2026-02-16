package model

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Title         string     `json:"title" db:"title"`
	ModelID       *uuid.UUID `json:"model_id,omitempty" db:"model_id"`
	MessageCount  int        `json:"message_count" db:"message_count"`
	TotalTokens   int        `json:"total_tokens" db:"total_tokens"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty" db:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// ConversationCreateRequest represents request to create a conversation
type ConversationCreateRequest struct {
	Title   string     `json:"title" binding:"omitempty,max=255"`
	ModelID *uuid.UUID `json:"model_id"`
}

// ConversationUpdateRequest represents request to update a conversation
type ConversationUpdateRequest struct {
	Title *string `json:"title" binding:"omitempty,max=255"`
}

// Message represents a chat message
type Message struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ConversationID uuid.UUID  `json:"conversation_id" db:"conversation_id"`
	Role           string     `json:"role" db:"role"` // "user", "assistant", "system"
	Content        string     `json:"content" db:"content"`
	InputTokens    *int       `json:"input_tokens,omitempty" db:"input_tokens"`
	OutputTokens   *int       `json:"output_tokens,omitempty" db:"output_tokens"`
	TotalTokens    *int       `json:"total_tokens,omitempty" db:"total_tokens"`
	ModelID        *uuid.UUID `json:"model_id,omitempty" db:"model_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// MessageCreateRequest represents request to create a message
type MessageCreateRequest struct {
	Content string `json:"content" binding:"required,min=1"`
	ModelID *uuid.UUID `json:"model_id"`
}

// ChatCompletionRequest represents OpenAI-compatible chat request
type ChatCompletionRequest struct {
	Model       string                 `json:"model"`
	Messages    []ChatMessage          `json:"messages"`
	Temperature *float64               `json:"temperature,omitempty"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream"`
	Stop        []string               `json:"stop,omitempty"`
	TopP        *float64               `json:"top_p,omitempty"`
	N           *int                   `json:"n,omitempty"`
	User        string                 `json:"user,omitempty"`
}

// ChatMessage represents a chat message in OpenAI format
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents OpenAI-compatible chat response
type ChatCompletionResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []ChatCompletionChoice   `json:"choices"`
	Usage   ChatCompletionUsage      `json:"usage"`
}

// ChatCompletionChoice represents a choice in chat response
type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatCompletionUsage represents token usage
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionStreamResponse represents streaming response chunk
type ChatCompletionStreamResponse struct {
	ID      string                        `json:"id"`
	Object  string                        `json:"object"`
	Created int64                         `json:"created"`
	Model   string                        `json:"model"`
	Choices []ChatCompletionStreamChoice  `json:"choices"`
}

// ChatCompletionStreamChoice represents a choice in streaming response
type ChatCompletionStreamChoice struct {
	Index        int                  `json:"index"`
	Delta        ChatMessageDelta     `json:"delta"`
	FinishReason *string              `json:"finish_reason"`
}

// ChatMessageDelta represents incremental message content
type ChatMessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}
