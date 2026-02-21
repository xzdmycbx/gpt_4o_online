package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/service"
)

// ChatHandler handles chat-related endpoints
type ChatHandler struct {
	chatService *service.ChatService
	upgrader    websocket.Upgrader
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow only same origin in production
				// You can customize this to check specific allowed origins
				origin := r.Header.Get("Origin")
				// Allow localhost for development
				if origin == "http://localhost:3000" || origin == "http://localhost:8080" {
					return true
				}
				// In production, check against your actual domain
				// return origin == "https://your-domain.com"
				return false
			},
		},
	}
}

// CreateConversation creates a new conversation
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	var req model.ConversationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	conv, err := h.chatService.CreateConversation(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

// ListConversations lists user's conversations
func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, err := h.chatService.ListConversations(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if conversations == nil {
		conversations = []*model.Conversation{}
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetConversation retrieves a specific conversation
func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conv, err := h.chatService.GetConversation(c.Request.Context(), userID, conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// UpdateConversation updates a conversation
func (h *ChatHandler) UpdateConversation(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req model.ConversationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	conv, err := h.chatService.UpdateConversation(c.Request.Context(), userID, conversationID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// DeleteConversation deletes a conversation
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	if err := h.chatService.DeleteConversation(c.Request.Context(), userID, conversationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

// GetMessages retrieves messages for a conversation
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatService.GetMessages(c.Request.Context(), userID, conversationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if messages == nil {
		messages = []*model.Message{}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// SendMessage sends a message and gets AI response
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req model.MessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userMsg, assistantMsg, err := h.chatService.SendMessage(c.Request.Context(), userID, conversationID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_message":      userMsg,
		"assistant_message": assistantMsg,
	})
}

// StreamChat handles WebSocket streaming chat
func (h *ChatHandler) StreamChat(c *gin.Context) {
	// Get authenticated user ID from context (set by auth middleware)
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

	// Read initial message with request
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return
	}

	var wsRequest struct {
		ConversationID string                       `json:"conversation_id"`
		Message        model.MessageCreateRequest   `json:"message"`
	}

	if err := json.Unmarshal(msg, &wsRequest); err != nil {
		conn.WriteJSON(gin.H{"error": "Invalid request"})
		return
	}

	conversationID, err := uuid.Parse(wsRequest.ConversationID)
	if err != nil {
		conn.WriteJSON(gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Verify user has access to this conversation
	conv, err := h.chatService.GetConversation(c.Request.Context(), userID, conversationID)
	if err != nil {
		conn.WriteJSON(gin.H{"error": "Unauthorized or conversation not found"})
		return
	}

	_ = conv // Conversation verified

	// This is a simplified implementation
	// In production, implement proper streaming with AI service

	// Send chunks
	response := "This is a streaming response placeholder."
	for i, char := range response {
		chunk := model.ChatCompletionStreamResponse{
			ID:      uuid.New().String(),
			Object:  "chat.completion.chunk",
			Created: 0,
			Model:   "gpt-4",
			Choices: []model.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: model.ChatMessageDelta{
						Content: string(char),
					},
					FinishReason: func() *string {
						if i == len(response)-1 {
							s := "stop"
							return &s
						}
						return nil
					}(),
				},
			},
		}

		if err := conn.WriteJSON(chunk); err != nil {
			return
		}
	}
}

// getUserID extracts user ID from context
func (h *ChatHandler) getUserID(c *gin.Context) uuid.UUID {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return uuid.Nil
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return uuid.Nil
	}

	return userID
}

// StreamSSE handles Server-Sent Events streaming (alternative to WebSocket)
func (h *ChatHandler) StreamSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Get user ID
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req model.MessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get streaming response
	responseChan, errorChan, err := h.chatService.GetStreamingResponse(c.Request.Context(), userID, conversationID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Stream responses
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	writer := bufio.NewWriter(w)

	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				return
			}

			data, _ := json.Marshal(chunk)
			writer.WriteString("data: " + string(data) + "\n\n")
			writer.Flush()
			flusher.Flush()

		case err := <-errorChan:
			if err != nil {
				writer.WriteString("event: error\ndata: " + err.Error() + "\n\n")
				writer.Flush()
				flusher.Flush()
			}
			return

		case <-c.Request.Context().Done():
			return
		}
	}
}
