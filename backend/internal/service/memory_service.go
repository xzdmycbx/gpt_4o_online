package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/repository"
)

const memoryCharBudget = 1200

type memoryCacheEntry struct {
	context   string
	expiresAt time.Time
}

// MemoryService handles memory extraction and management
type MemoryService struct {
	memoryRepo     *repository.MemoryRepository
	messageRepo    *repository.MessageRepository
	aiProxyService *AIProxyService
	modelRepo      *repository.AIModelRepository
	enabled        bool
	defaultModel   string
	cache          sync.Map      // key: uuid.UUID string → *memoryCacheEntry
	cacheTTL       time.Duration // 5 minutes
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
		cacheTTL:       5 * time.Minute,
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

	// Invalidate cache for this user so next BuildMemoryContext reloads
	s.cache.Delete(userID.String())

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
	// Optimized prompt — minimal, targeted, anti-hallucination
	prompt := fmt.Sprintf(`对话内容：
%s

提取用户明确说明的信息（每条≤40字），输出JSON数组：
[{"content":"内容","category":"preference|fact|context","importance":1-10}]
无信息时返回空数组[]`, conversationText)

	// Get memory extraction model
	var modelID uuid.UUID
	var modelIdentifier string
	models, err := s.modelRepo.List(ctx, true)
	if err != nil || len(models) == 0 {
		return nil, fmt.Errorf("no active models available")
	}

	// Find the default memory model or use first available
	for _, m := range models {
		if m.ModelIdentifier == s.defaultModel || m.Name == s.defaultModel {
			modelID = m.ID
			modelIdentifier = m.ModelIdentifier
			break
		}
	}
	if modelID == uuid.Nil {
		modelID = models[0].ID // Use first available model
		modelIdentifier = models[0].ModelIdentifier
	}

	// Prepare chat request
	request := &model.ChatCompletionRequest{
		Model: modelIdentifier,
		Messages: []model.ChatMessage{
			{
				Role:    "system",
				Content: "只提取对话中用户明确说明的信息，不推断，不补充，不虚构。返回JSON数组。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: func() *float64 { t := 0.3; return &t }(),
		MaxTokens:   func() *int { t := 400; return &t }(),
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

// DeleteMemory deletes a memory and invalidates the user's cache
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

	if err := s.memoryRepo.Delete(ctx, memoryID); err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Delete(userID.String())
	return nil
}

// CleanupOldMemories removes old low-importance memories
func (s *MemoryService) CleanupOldMemories(ctx context.Context, userID uuid.UUID) error {
	// Delete memories with importance <= 3 that are older than 30 days and unused
	return s.memoryRepo.DeleteLowImportance(ctx, userID, 30, 3)
}

// buildBudgetedContext builds memory context within ~1200 character budget
func (s *MemoryService) buildBudgetedContext(memories []*model.Memory) string {
	var builder strings.Builder
	total := 0

	for _, mem := range memories {
		line := fmt.Sprintf("- [%s] %s\n", mem.Category, mem.Content)
		if total+len(line) > memoryCharBudget {
			break
		}
		builder.WriteString(line)
		total += len(line)
	}

	return builder.String()
}

// BuildMemoryContext builds a context string from relevant memories (with cache)
func (s *MemoryService) BuildMemoryContext(ctx context.Context, userID uuid.UUID) (string, error) {
	// Check cache
	cacheKey := userID.String()
	if v, ok := s.cache.Load(cacheKey); ok {
		entry := v.(*memoryCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.context, nil
		}
		s.cache.Delete(cacheKey)
	}

	// Fetch memories (limit 20, budget will trim)
	memories, err := s.GetRelevantMemories(ctx, userID, 20)
	if err != nil {
		return "", err
	}

	if len(memories) == 0 {
		return "", nil
	}

	body := s.buildBudgetedContext(memories)
	if body == "" {
		return "", nil
	}

	result := "关于用户的记忆：\n" + body

	// Store in cache
	s.cache.Store(cacheKey, &memoryCacheEntry{
		context:   result,
		expiresAt: time.Now().Add(s.cacheTTL),
	})

	return result, nil
}
