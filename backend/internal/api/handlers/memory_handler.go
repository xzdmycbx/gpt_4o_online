package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/service"
)

// MemoryHandler handles memory-related endpoints
type MemoryHandler struct {
	memoryService *service.MemoryService
}

// NewMemoryHandler creates a new memory handler
func NewMemoryHandler(memoryService *service.MemoryService) *MemoryHandler {
	return &MemoryHandler{
		memoryService: memoryService,
	}
}

// List retrieves user's memories
func (h *MemoryHandler) List(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	memories, err := h.memoryService.GetUserMemories(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"limit":    limit,
		"offset":   offset,
	})
}

// Create creates a new memory manually
func (h *MemoryHandler) Create(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	var req model.MemoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	memory, err := h.memoryService.CreateMemory(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, memory)
}

// Update updates an existing memory
func (h *MemoryHandler) Update(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	memoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	var req model.MemoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	memory, err := h.memoryService.UpdateMemory(c.Request.Context(), userID, memoryID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, memory)
}

// Delete deletes a memory
func (h *MemoryHandler) Delete(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == uuid.Nil {
		return
	}

	memoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	if err := h.memoryService.DeleteMemory(c.Request.Context(), userID, memoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory deleted"})
}

// getUserID extracts user ID from context
func (h *MemoryHandler) getUserID(c *gin.Context) uuid.UUID {
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
