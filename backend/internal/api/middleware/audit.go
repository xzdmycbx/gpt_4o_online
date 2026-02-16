package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/repository"
)

// AuditMiddleware logs important actions to audit log
func AuditMiddleware(auditRepo *repository.AuditLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute request first
		c.Next()

		// Only log for certain methods and successful requests
		if c.Request.Method == "GET" || c.Writer.Status() >= 400 {
			return
		}

		// Get audit action from context if set by handler
		action, exists := c.Get("audit_action")
		if !exists {
			return
		}

		auditAction := action.(model.AuditAction)

		// Get actor ID
		var actorID *uuid.UUID
		if userIDStr, exists := c.Get("user_id"); exists {
			if userID, err := uuid.Parse(userIDStr.(string)); err == nil {
				actorID = &userID
			}
		}

		// Get target user ID if exists
		var targetUserID *uuid.UUID
		if targetIDStr, exists := c.Get("audit_target_user_id"); exists {
			if targetID, err := uuid.Parse(targetIDStr.(string)); err == nil {
				targetUserID = &targetID
			}
		}

		// Get details
		var details map[string]interface{}
		if d, exists := c.Get("audit_details"); exists {
			details = d.(map[string]interface{})
		}

		// Get username
		var username string
		if un, exists := c.Get("username"); exists {
			username = un.(string)
		}

		// Get resource type
		var resourceType string
		if rt, exists := c.Get("audit_resource_type"); exists {
			resourceType = rt.(string)
		}

		// Get client IP
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Create audit log
		log := &model.AuditLog{
			ID:           uuid.New(),
			Action:       auditAction,
			ActorID:      actorID,
			Username:     username,
			ResourceType: resourceType,
			TargetUserID: targetUserID,
			Details:      details,
			IPAddress:    &clientIP,
			UserAgent:    &userAgent,
			CreatedAt:    time.Now(),
		}

		// Save to database (async to not slow down request)
		// Use background context to avoid capturing gin.Context which can be reused
		go func(logCopy *model.AuditLog) {
			ctx := context.Background()
			_ = auditRepo.Create(ctx, logCopy)
		}(log)
	}
}
