package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/repository"
)

// ChatService handles chat-related business logic
type ChatService struct {
	convRepo         *repository.ConversationRepository
	msgRepo          *repository.MessageRepository
	modelRepo        *repository.AIModelRepository
	tokenUsageRepo   *repository.TokenUsageRepository
	aiProxyService   *AIProxyService
	memoryService    *MemoryService
}

// NewChatService creates a new chat service
func NewChatService(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	modelRepo *repository.AIModelRepository,
	tokenUsageRepo *repository.TokenUsageRepository,
	aiProxyService *AIProxyService,
	memoryService *MemoryService,
) *ChatService {
	return &ChatService{
		convRepo:       convRepo,
		msgRepo:        msgRepo,
		modelRepo:      modelRepo,
		tokenUsageRepo: tokenUsageRepo,
		aiProxyService: aiProxyService,
		memoryService:  memoryService,
	}
}

// CreateConversation creates a new conversation
func (s *ChatService) CreateConversation(ctx context.Context, userID uuid.UUID, req *model.ConversationCreateRequest) (*model.Conversation, error) {
	title := "New Conversation"
	if req.Title != "" {
		title = req.Title
	}

	conv := &model.Conversation{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		ModelID: req.ModelID,
	}

	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv, nil
}

// GetConversation retrieves a conversation by ID
func (s *ChatService) GetConversation(ctx context.Context, userID, conversationID uuid.UUID) (*model.Conversation, error) {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Check ownership
	if conv.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	return conv, nil
}

// ListConversations lists conversations for a user
func (s *ChatService) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Conversation, error) {
	return s.convRepo.ListByUser(ctx, userID, limit, offset)
}

// UpdateConversation updates a conversation
func (s *ChatService) UpdateConversation(ctx context.Context, userID, conversationID uuid.UUID, req *model.ConversationUpdateRequest) (*model.Conversation, error) {
	conv, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		conv.Title = *req.Title
	}

	if err := s.convRepo.Update(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return conv, nil
}

// DeleteConversation deletes a conversation and all its messages
func (s *ChatService) DeleteConversation(ctx context.Context, userID, conversationID uuid.UUID) error {
	conv, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return err
	}

	if conv.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.convRepo.Delete(ctx, conversationID)
}

// GetMessages retrieves messages for a conversation
func (s *ChatService) GetMessages(ctx context.Context, userID, conversationID uuid.UUID, limit, offset int) ([]*model.Message, error) {
	// Verify ownership
	_, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	return s.msgRepo.ListByConversation(ctx, conversationID, limit, offset)
}

// SendMessage sends a message and gets AI response
func (s *ChatService) SendMessage(ctx context.Context, userID, conversationID uuid.UUID, req *model.MessageCreateRequest) (*model.Message, *model.Message, error) {
	// Verify conversation ownership
	conv, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, nil, err
	}

	// Determine which model to use
	modelID := conv.ModelID
	if req.ModelID != nil {
		modelID = req.ModelID
	}

	if modelID == nil {
		// Get default model
		defaultModel, err := s.modelRepo.GetDefault(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("no model specified and no default model configured")
		}
		modelID = &defaultModel.ID
	}

	// Save user message
	userMsg := &model.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           "user",
		Content:        req.Content,
		ModelID:        modelID,
	}

	if err := s.msgRepo.Create(ctx, userMsg); err != nil {
		return nil, nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get conversation history
	messages, err := s.msgRepo.GetRecentMessages(ctx, conversationID, 20)
	if err != nil {
		return userMsg, nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Build chat messages for AI
	chatMessages := make([]model.ChatMessage, 0, len(messages))

	// Add memory context if available
	memoryContext, err := s.memoryService.BuildMemoryContext(ctx, userID)
	if err == nil && memoryContext != "" {
		chatMessages = append(chatMessages, model.ChatMessage{
			Role:    "system",
			Content: memoryContext,
		})
	}

	// Add conversation history
	for _, msg := range messages {
		chatMessages = append(chatMessages, model.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Prepare AI request
	aiModel, err := s.modelRepo.GetByID(ctx, *modelID)
	if err != nil {
		return userMsg, nil, fmt.Errorf("failed to get model: %w", err)
	}

	aiRequest := &model.ChatCompletionRequest{
		Model:    aiModel.ModelIdentifier,
		Messages: chatMessages,
	}

	// Send to AI
	aiResponse, err := s.aiProxyService.SendChatCompletion(ctx, *modelID, aiRequest)
	if err != nil {
		return userMsg, nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(aiResponse.Choices) == 0 {
		return userMsg, nil, fmt.Errorf("no response from AI")
	}

	// Save AI response
	assistantMsg := &model.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        aiResponse.Choices[0].Message.Content,
		InputTokens:    &aiResponse.Usage.PromptTokens,
		OutputTokens:   &aiResponse.Usage.CompletionTokens,
		TotalTokens:    &aiResponse.Usage.TotalTokens,
		ModelID:        modelID,
	}

	if err := s.msgRepo.Create(ctx, assistantMsg); err != nil {
		return userMsg, nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// Record token usage
	cost, _ := s.aiProxyService.EstimateCost(ctx, *modelID, aiResponse.Usage.PromptTokens, aiResponse.Usage.CompletionTokens)
	_ = s.tokenUsageRepo.RecordUsage(
		ctx,
		userID,
		modelID,
		aiResponse.Usage.PromptTokens,
		aiResponse.Usage.CompletionTokens,
		cost,
	)

	// Extract memories from conversation (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.memoryService.ExtractMemoriesFromConversation(ctx, userID, conversationID); err != nil {
			log.Printf("Memory extraction failed for conversation %s: %v", conversationID, err)
		}
	}()

	return userMsg, assistantMsg, nil
}

// GetStreamingResponse gets a streaming response from AI
// Returns a channel that emits response chunks
func (s *ChatService) GetStreamingResponse(ctx context.Context, userID, conversationID uuid.UUID, req *model.MessageCreateRequest) (<-chan *model.ChatCompletionStreamResponse, <-chan error, error) {
	// Verify conversation ownership
	conv, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, nil, err
	}

	// Determine which model to use
	modelID := conv.ModelID
	if req.ModelID != nil {
		modelID = req.ModelID
	}

	if modelID == nil {
		defaultModel, err := s.modelRepo.GetDefault(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("no model specified and no default model configured")
		}
		modelID = &defaultModel.ID
	}

	// Save user message
	userMsg := &model.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           "user",
		Content:        req.Content,
		ModelID:        modelID,
	}

	if err := s.msgRepo.Create(ctx, userMsg); err != nil {
		return nil, nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get conversation history
	messages, err := s.msgRepo.GetRecentMessages(ctx, conversationID, 20)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Build chat messages
	chatMessages := make([]model.ChatMessage, 0, len(messages))

	memoryContext, err := s.memoryService.BuildMemoryContext(ctx, userID)
	if err == nil && memoryContext != "" {
		chatMessages = append(chatMessages, model.ChatMessage{
			Role:    "system",
			Content: memoryContext,
		})
	}

	for _, msg := range messages {
		chatMessages = append(chatMessages, model.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Prepare AI request
	aiModel, err := s.modelRepo.GetByID(ctx, *modelID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get model: %w", err)
	}

	_ = &model.ChatCompletionRequest{
		Model:    aiModel.ModelIdentifier,
		Messages: chatMessages,
		Stream:   true,
	}

	// This would return channels for streaming
	// Implementation details depend on WebSocket/SSE setup in handlers
	responseChan := make(chan *model.ChatCompletionStreamResponse)
	errorChan := make(chan error, 1)

	// TODO: Implement actual streaming logic in Stage 4 with WebSocket handler

	return responseChan, errorChan, nil
}

// AvailableModel is a user-safe view of an AI model (no API keys or internal URLs)
type AvailableModel struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	Description      string `json:"description,omitempty"`
	IsDefault        bool   `json:"is_default"`
	SupportsStreaming bool   `json:"supports_streaming"`
	MaxTokens        int    `json:"max_tokens,omitempty"`
}

// ListAvailableModels returns active models with only user-safe fields
func (s *ChatService) ListAvailableModels(ctx context.Context) ([]AvailableModel, error) {
	models, err := s.modelRepo.List(ctx, true) // activeOnly=true
	if err != nil {
		return nil, err
	}
	result := make([]AvailableModel, 0, len(models))
	for _, m := range models {
		result = append(result, AvailableModel{
			ID:               m.ID.String(),
			Name:             m.Name,
			DisplayName:      m.DisplayName,
			Description:      m.Description,
			IsDefault:        m.IsDefault,
			SupportsStreaming: m.SupportsStreaming,
			MaxTokens:        m.MaxTokens,
		})
	}
	return result, nil
}
