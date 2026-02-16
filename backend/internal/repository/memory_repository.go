package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
)

// MemoryRepository handles memory data access
type MemoryRepository struct {
	db *sql.DB
}

// NewMemoryRepository creates a new memory repository
func NewMemoryRepository(db *sql.DB) *MemoryRepository {
	return &MemoryRepository{db: db}
}

// Create creates a new memory
func (r *MemoryRepository) Create(ctx context.Context, memory *model.Memory) error {
	query := `
		INSERT INTO memories (id, user_id, content, category, importance, source_conversation_id, source_message_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		memory.ID, memory.UserID, memory.Content, memory.Category, memory.Importance,
		memory.SourceConversationID, memory.SourceMessageID,
	).Scan(&memory.CreatedAt, &memory.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}

	return nil
}

// GetByID retrieves a memory by ID
func (r *MemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Memory, error) {
	query := `
		SELECT id, user_id, content, category, importance,
			source_conversation_id, source_message_id,
			times_used, last_used_at, created_at, updated_at
		FROM memories WHERE id = $1
	`

	memory := &model.Memory{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&memory.ID, &memory.UserID, &memory.Content, &memory.Category, &memory.Importance,
		&memory.SourceConversationID, &memory.SourceMessageID,
		&memory.TimesUsed, &memory.LastUsedAt, &memory.CreatedAt, &memory.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("memory not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	return memory, nil
}

// ListByUser retrieves memories for a user
func (r *MemoryRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Memory, error) {
	query := `
		SELECT id, user_id, content, category, importance,
			source_conversation_id, source_message_id,
			times_used, last_used_at, created_at, updated_at
		FROM memories
		WHERE user_id = $1
		ORDER BY importance DESC, last_used_at DESC NULLS LAST
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	var memories []*model.Memory
	for rows.Next() {
		memory := &model.Memory{}
		err := rows.Scan(
			&memory.ID, &memory.UserID, &memory.Content, &memory.Category, &memory.Importance,
			&memory.SourceConversationID, &memory.SourceMessageID,
			&memory.TimesUsed, &memory.LastUsedAt, &memory.CreatedAt, &memory.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		memories = append(memories, memory)
	}

	return memories, nil
}

// GetRelevantMemories retrieves most relevant memories for a user
func (r *MemoryRepository) GetRelevantMemories(ctx context.Context, userID uuid.UUID, limit int) ([]*model.Memory, error) {
	query := `
		SELECT id, user_id, content, category, importance,
			source_conversation_id, source_message_id,
			times_used, last_used_at, created_at, updated_at
		FROM memories
		WHERE user_id = $1
		ORDER BY importance DESC, times_used DESC, last_used_at DESC NULLS LAST
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get relevant memories: %w", err)
	}
	defer rows.Close()

	var memories []*model.Memory
	for rows.Next() {
		memory := &model.Memory{}
		err := rows.Scan(
			&memory.ID, &memory.UserID, &memory.Content, &memory.Category, &memory.Importance,
			&memory.SourceConversationID, &memory.SourceMessageID,
			&memory.TimesUsed, &memory.LastUsedAt, &memory.CreatedAt, &memory.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		memories = append(memories, memory)
	}

	return memories, nil
}

// Update updates a memory
func (r *MemoryRepository) Update(ctx context.Context, memory *model.Memory) error {
	query := `
		UPDATE memories SET
			content = $2,
			category = $3,
			importance = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, memory.ID, memory.Content, memory.Category, memory.Importance)
	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	return nil
}

// IncrementUsage increments the usage count and updates last used time
func (r *MemoryRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE memories SET
			times_used = times_used + 1,
			last_used_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Delete deletes a memory
func (r *MemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM memories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteLowImportance deletes old low-importance memories for cleanup
func (r *MemoryRepository) DeleteLowImportance(ctx context.Context, userID uuid.UUID, maxAge int, maxImportance int) error {
	query := `
		DELETE FROM memories
		WHERE user_id = $1
			AND importance <= $2
			AND created_at < NOW() - INTERVAL '1 day' * $3
			AND times_used = 0
	`

	_, err := r.db.ExecContext(ctx, query, userID, maxImportance, maxAge)
	return err
}
