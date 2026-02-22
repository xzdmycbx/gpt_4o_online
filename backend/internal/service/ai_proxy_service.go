package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/crypto"
	"github.com/ai-chat/backend/internal/repository"
)

// AIProxyService handles AI API interactions
type AIProxyService struct {
	modelRepo    *repository.AIModelRepository
	providerRepo *repository.AIProviderRepository
	encryptionKey string
	httpClient   *http.Client
}

// NewAIProxyService creates a new AI proxy service
func NewAIProxyService(modelRepo *repository.AIModelRepository, providerRepo *repository.AIProviderRepository, encryptionKey string) *AIProxyService {
	return &AIProxyService{
		modelRepo:    modelRepo,
		providerRepo: providerRepo,
		encryptionKey: encryptionKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// resolveCredentials returns the API endpoint and decrypted key for a model,
// falling back to the linked provider if the model itself has no credentials.
func (s *AIProxyService) resolveCredentials(ctx context.Context, aiModel *model.AIModel) (endpoint, apiKey string, err error) {
	if aiModel.ProviderID != nil && (aiModel.APIEndpoint == "" || aiModel.APIKeyEncrypted == "") {
		// Resolve from provider
		provider, pErr := s.providerRepo.GetByID(ctx, *aiModel.ProviderID)
		if pErr != nil {
			return "", "", fmt.Errorf("failed to get provider: %w", pErr)
		}
		decrypted, dErr := crypto.Decrypt(provider.APIKeyEncrypted, s.encryptionKey)
		if dErr != nil {
			return "", "", fmt.Errorf("failed to decrypt provider API key: %w", dErr)
		}
		return provider.APIEndpoint, decrypted, nil
	}

	// Use model's own credentials
	decrypted, dErr := crypto.Decrypt(aiModel.APIKeyEncrypted, s.encryptionKey)
	if dErr != nil {
		return "", "", fmt.Errorf("failed to decrypt API key: %w", dErr)
	}
	return aiModel.APIEndpoint, decrypted, nil
}

// SendChatCompletion sends a chat completion request to an AI model
func (s *AIProxyService) SendChatCompletion(ctx context.Context, modelID uuid.UUID, request *model.ChatCompletionRequest) (*model.ChatCompletionResponse, error) {
	// Get model configuration
	aiModel, err := s.modelRepo.GetByID(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if !aiModel.IsActive {
		return nil, fmt.Errorf("model is not active")
	}

	endpoint, apiKey, err := s.resolveCredentials(ctx, aiModel)
	if err != nil {
		return nil, err
	}

	// Prepare request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response model.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// SendStreamingChatCompletion sends a streaming chat completion request
func (s *AIProxyService) SendStreamingChatCompletion(ctx context.Context, modelID uuid.UUID, request *model.ChatCompletionRequest) (io.ReadCloser, error) {
	// Get model configuration
	aiModel, err := s.modelRepo.GetByID(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if !aiModel.IsActive {
		return nil, fmt.Errorf("model is not active")
	}

	if !aiModel.SupportsStreaming {
		return nil, fmt.Errorf("model does not support streaming")
	}

	endpoint, apiKey, err := s.resolveCredentials(ctx, aiModel)
	if err != nil {
		return nil, err
	}

	// Ensure streaming is enabled
	request.Stream = true

	// Prepare request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

// EstimateCost estimates the cost of a completion based on token usage
func (s *AIProxyService) EstimateCost(ctx context.Context, modelID uuid.UUID, inputTokens, outputTokens int) (*float64, error) {
	aiModel, err := s.modelRepo.GetByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	if aiModel.InputPricePer1k == nil || aiModel.OutputPricePer1k == nil {
		return nil, nil // No pricing information
	}

	inputCost := float64(inputTokens) / 1000.0 * *aiModel.InputPricePer1k
	outputCost := float64(outputTokens) / 1000.0 * *aiModel.OutputPricePer1k
	totalCost := inputCost + outputCost

	return &totalCost, nil
}

// CountTokens estimates token count for a text (simple approximation)
// In production, use a proper tokenizer like tiktoken
func (s *AIProxyService) CountTokens(text string) int {
	// Simple approximation: ~4 characters per token
	return len(text) / 4
}

// CountMessagesTokens estimates total tokens for a list of messages
func (s *AIProxyService) CountMessagesTokens(messages []model.ChatMessage) int {
	total := 0
	for _, msg := range messages {
		total += s.CountTokens(msg.Content)
		total += 4 // overhead per message
	}
	return total
}
