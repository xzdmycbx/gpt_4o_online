package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
)

// ConversationRepository handles conversation data access
type ConversationRepository struct {
	db *sql.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create creates a new conversation
func (r *ConversationRepository) Create(ctx context.Context, conv *model.Conversation) error {
	query := `
		INSERT INTO conversations (id, user_id, title, model_id)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		conv.ID, conv.UserID, conv.Title, conv.ModelID,
	).Scan(&conv.CreatedAt, &conv.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	return nil
}

// GetByID retrieves a conversation by ID
func (r *ConversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Conversation, error) {
	query := `
		SELECT id, user_id, title, model_id, message_count, total_tokens,
			last_message_at, created_at, updated_at
		FROM conversations WHERE id = $1
	`

	conv := &model.Conversation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID, &conv.UserID, &conv.Title, &conv.ModelID, &conv.MessageCount, &conv.TotalTokens,
		&conv.LastMessageAt, &conv.CreatedAt, &conv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conv, nil
}

// ListByUser retrieves conversations for a user
func (r *ConversationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Conversation, error) {
	query := `
		SELECT id, user_id, title, model_id, message_count, total_tokens,
			last_message_at, created_at, updated_at
		FROM conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*model.Conversation
	for rows.Next() {
		conv := &model.Conversation{}
		err := rows.Scan(
			&conv.ID, &conv.UserID, &conv.Title, &conv.ModelID, &conv.MessageCount, &conv.TotalTokens,
			&conv.LastMessageAt, &conv.CreatedAt, &conv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// Update updates a conversation
func (r *ConversationRepository) Update(ctx context.Context, conv *model.Conversation) error {
	query := `
		UPDATE conversations SET
			title = $2,
			model_id = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, conv.ID, conv.Title, conv.ModelID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// Delete deletes a conversation
func (r *ConversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Count counts total conversations
func (r *ConversationRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM conversations`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MessageRepository handles message data access
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, role, content, input_tokens, output_tokens, total_tokens, model_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		msg.ID, msg.ConversationID, msg.Role, msg.Content,
		msg.InputTokens, msg.OutputTokens, msg.TotalTokens, msg.ModelID,
	).Scan(&msg.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, input_tokens, output_tokens, total_tokens, model_id, created_at
		FROM messages WHERE id = $1
	`

	msg := &model.Message{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
		&msg.InputTokens, &msg.OutputTokens, &msg.TotalTokens, &msg.ModelID, &msg.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msg, nil
}

// ListByConversation retrieves messages for a conversation
func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*model.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, input_tokens, output_tokens, total_tokens, model_id, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
			&msg.InputTokens, &msg.OutputTokens, &msg.TotalTokens, &msg.ModelID, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetRecentMessages retrieves recent messages for memory context
func (r *MessageRepository) GetRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*model.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, input_tokens, output_tokens, total_tokens, model_id, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
			&msg.InputTokens, &msg.OutputTokens, &msg.TotalTokens, &msg.ModelID, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Delete deletes a message
func (r *MessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Count counts total messages
func (r *MessageRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM messages`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountToday counts messages created today
func (r *MessageRepository) CountToday(ctx context.Context) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM messages
		WHERE created_at >= CURRENT_DATE
	`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountLastDays counts messages in the last N days
func (r *MessageRepository) CountLastDays(ctx context.Context, days int) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM messages
		WHERE created_at >= NOW() - $1 * INTERVAL '1 day'
	`
	err := r.db.QueryRowContext(ctx, query, days).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
