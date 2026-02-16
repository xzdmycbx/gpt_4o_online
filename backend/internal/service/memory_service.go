package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/repository"
)

// MemoryService handles memory extraction and management
type MemoryService struct {
	memoryRepo     *repository.MemoryRepository
	messageRepo    *repository.MessageRepository
	aiProxyService *AIProxyService
	modelRepo      *repository.AIModelRepository
	enabled        bool
	defaultModel   string
}

// NewMemoryService creates a new memory service
func NewMemoryService(
	memoryRepo *repository.MemoryRepository,
	messageRepo *repository.MessageRepository,
	aiProxyService *AIProxyService,
	modelRepo *repository.AIModelRepository,
	enabled bool,
	defaultModel string,
) *MemoryService {
	return &MemoryService{
		memoryRepo:     memoryRepo,
		messageRepo:    messageRepo,
		aiProxyService: aiProxyService,
		modelRepo:      modelRepo,
		enabled:        enabled,
		defaultModel:   defaultModel,
	}
}

// ExtractMemoriesFromConversation extracts memories from recent conversation
func (s *MemoryService) ExtractMemoriesFromConversation(ctx context.Context, userID, conversationID uuid.UUID) error {
	if !s.enabled {
		return nil
	}

	// Get recent messages (last 10)
	messages, err := s.messageRepo.GetRecentMessages(ctx, conversationID, 10)
	if err != nil {
		return fmt.Errorf("failed to get recent messages: %w", err)
	}

	if len(messages) < 2 {
		return nil // Not enough context for memory extraction
	}

	// Build conversation context
	conversationText := s.buildConversationText(messages)

	// Extract memories using small model
	extractedMemories, err := s.extractMemoriesWithAI(ctx, conversationText)
	if err != nil {
		return fmt.Errorf("failed to extract memories: %w", err)
	}

	// Save extracted memories
	for _, mem := range extractedMemories {
		// Check for duplicates
		existingMemories, err := s.memoryRepo.GetRelevantMemories(ctx, userID, 100)
		if err != nil {
			continue
		}

		isDuplicate := false
		for _, existing := range existingMemories {
			if s.isSimilar(existing.Content, mem.Content) {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			continue
		}

		// Create new memory
		memory := &model.Memory{
			ID:                   uuid.New(),
			UserID:               userID,
			Content:              mem.Content,
			Category:             mem.Category,
			Importance:           mem.Importance,
			SourceConversationID: &conversationID,
		}

		if err := s.memoryRepo.Create(ctx, memory); err != nil {
			// Log error but continue with other memories
			continue
		}
	}

	return nil
}

// buildConversationText builds a text representation of the conversation
func (s *MemoryService) buildConversationText(messages []*model.Message) string {
	var builder strings.Builder

	for _, msg := range messages {
		role := msg.Role
		if role == "user" {
			role = "用户"
		} else if role == "assistant" {
			role = "助手"
		}
		builder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	return builder.String()
}

// extractMemoriesWithAI uses AI to extract memories from conversation
func (s *MemoryService) extractMemoriesWithAI(ctx context.Context, conversationText string) ([]model.MemoryExtractionResult, error) {
	// Construct prompt for memory extraction (under 100 tokens)
	prompt := fmt.Sprintf(`根据对话提取用户信息，生成简洁记忆（每条≤50字）。
分类：preference/fact/context
返回JSON：[{"content":"记忆","category":"类别","importance":1-10}]
对话：
%s`, conversationText)

	// Get memory extraction model
	var modelID uuid.UUID
	models, err := s.modelRepo.List(ctx, true)
	if err != nil || len(models) == 0 {
		return nil, fmt.Errorf("no active models available")
	}

	// Find the default memory model or use first available
	for _, m := range models {
		if m.ModelIdentifier == s.defaultModel || m.Name == s.defaultModel {
			modelID = m.ID
			break
		}
	}
	if modelID == uuid.Nil {
		modelID = models[0].ID // Use first available model
	}

	// Prepare chat request
	request := &model.ChatCompletionRequest{
		Model: s.defaultModel,
		Messages: []model.ChatMessage{
			{
				Role:    "system",
				Content: "你是一个记忆提取专家，从对话中提取重要的用户信息。只返回JSON格式的数组，不要其他内容。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: func() *float64 { t := 0.3; return &t }(),
		MaxTokens:   func() *int { t := 500; return &t }(),
	}

	// Send request to AI
	response, err := s.aiProxyService.SendChatCompletion(ctx, modelID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Parse JSON response
	content := response.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Extract JSON from markdown code blocks if present
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var memories []model.MemoryExtractionResult
	if err := json.Unmarshal([]byte(content), &memories); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return memories, nil
}

// isSimilar checks if two memory contents are similar (simple check)
func (s *MemoryService) isSimilar(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))

	// Simple similarity: if one contains the other or they're very similar
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}

	// Could implement more sophisticated similarity check (Levenshtein, etc.)
	return false
}

// GetUserMemories retrieves memories for a user
func (s *MemoryService) GetUserMemories(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Memory, error) {
	return s.memoryRepo.ListByUser(ctx, userID, limit, offset)
}

// GetRelevantMemories gets the most relevant memories for context
func (s *MemoryService) GetRelevantMemories(ctx context.Context, userID uuid.UUID, limit int) ([]*model.Memory, error) {
	memories, err := s.memoryRepo.GetRelevantMemories(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	// Increment usage count for retrieved memories
	for _, mem := range memories {
		_ = s.memoryRepo.IncrementUsage(ctx, mem.ID)
	}

	return memories, nil
}

// CreateMemory creates a new memory manually
func (s *MemoryService) CreateMemory(ctx context.Context, userID uuid.UUID, req *model.MemoryCreateRequest) (*model.Memory, error) {
	memory := &model.Memory{
		ID:         uuid.New(),
		UserID:     userID,
		Content:    req.Content,
		Category:   req.Category,
		Importance: req.Importance,
	}

	if err := s.memoryRepo.Create(ctx, memory); err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}

	return memory, nil
}

// UpdateMemory updates an existing memory
func (s *MemoryService) UpdateMemory(ctx context.Context, userID, memoryID uuid.UUID, req *model.MemoryUpdateRequest) (*model.Memory, error) {
	// Get existing memory
	memory, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("memory not found")
	}

	// Check ownership
	if memory.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Update fields
	if req.Content != nil {
		memory.Content = *req.Content
	}
	if req.Category != nil {
		memory.Category = *req.Category
	}
	if req.Importance != nil {
		memory.Importance = *req.Importance
	}

	if err := s.memoryRepo.Update(ctx, memory); err != nil {
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}

	return memory, nil
}

// DeleteMemory deletes a memory
func (s *MemoryService) DeleteMemory(ctx context.Context, userID, memoryID uuid.UUID) error {
	// Get existing memory
	memory, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("memory not found")
	}

	// Check ownership
	if memory.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.memoryRepo.Delete(ctx, memoryID)
}

// CleanupOldMemories removes old low-importance memories
func (s *MemoryService) CleanupOldMemories(ctx context.Context, userID uuid.UUID) error {
	// Delete memories with importance <= 3 that are older than 30 days and unused
	return s.memoryRepo.DeleteLowImportance(ctx, userID, 30, 3)
}

// BuildMemoryContext builds a context string from relevant memories
func (s *MemoryService) BuildMemoryContext(ctx context.Context, userID uuid.UUID) (string, error) {
	memories, err := s.GetRelevantMemories(ctx, userID, 10)
	if err != nil {
		return "", err
	}

	if len(memories) == 0 {
		return "", nil
	}

	var builder strings.Builder
	builder.WriteString("关于用户的记忆：\n")

	for _, mem := range memories {
		builder.WriteString(fmt.Sprintf("- %s\n", mem.Content))
	}

	return builder.String(), nil
}
