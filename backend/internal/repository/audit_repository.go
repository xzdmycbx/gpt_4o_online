package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
)

// AuditLogRepository handles audit log data access
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO audit_logs (id, action, actor_id, username, resource_type, target_user_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
		log.ID, log.Action, log.ActorID, log.Username, log.ResourceType, log.TargetUserID, detailsJSON, log.IPAddress, log.UserAgent,
	).Scan(&log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// List retrieves audit logs with pagination
func (r *AuditLogRepository) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, error) {
	query := `
		SELECT
			al.id, al.action, al.actor_id,
			COALESCE(al.username, u.username, 'system') as username,
			COALESCE(al.resource_type, 'system') as resource_type,
			al.target_user_id, al.details, al.ip_address, al.user_agent, al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.actor_id = u.id
		ORDER BY al.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		log := &model.AuditLog{}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID, &log.Action, &log.ActorID, &log.Username, &log.ResourceType, &log.TargetUserID,
			&detailsJSON, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// ListByActor retrieves audit logs for a specific actor
func (r *AuditLogRepository) ListByActor(ctx context.Context, actorID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	query := `
		SELECT id, action, actor_id, target_user_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE actor_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, actorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs by actor: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		log := &model.AuditLog{}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID, &log.Action, &log.ActorID, &log.TargetUserID,
			&detailsJSON, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// TokenUsageRepository handles token usage data access
type TokenUsageRepository struct {
	db *sql.DB
}

// NewTokenUsageRepository creates a new token usage repository
func NewTokenUsageRepository(db *sql.DB) *TokenUsageRepository {
	return &TokenUsageRepository{db: db}
}

// Create creates or updates token usage for a period
func (r *TokenUsageRepository) Create(ctx context.Context, usage *model.TokenUsage) error {
	query := `
		INSERT INTO token_usage (
			id, user_id, model_id, input_tokens, output_tokens, total_tokens,
			estimated_cost, period_start, period_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, model_id, period_start) DO UPDATE SET
			input_tokens = token_usage.input_tokens + EXCLUDED.input_tokens,
			output_tokens = token_usage.output_tokens + EXCLUDED.output_tokens,
			total_tokens = token_usage.total_tokens + EXCLUDED.total_tokens,
			estimated_cost = token_usage.estimated_cost + EXCLUDED.estimated_cost
		RETURNING created_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		usage.ID, usage.UserID, usage.ModelID,
		usage.InputTokens, usage.OutputTokens, usage.TotalTokens,
		usage.EstimatedCost, usage.PeriodStart, usage.PeriodEnd,
	).Scan(&usage.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create/update token usage: %w", err)
	}

	return nil
}

// RecordUsage records token usage for a user
func (r *TokenUsageRepository) RecordUsage(ctx context.Context, userID uuid.UUID, modelID *uuid.UUID, inputTokens, outputTokens int, cost *float64) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 0, 1)

	usage := &model.TokenUsage{
		ID:            uuid.New(),
		UserID:        userID,
		ModelID:       modelID,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		EstimatedCost: cost,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
	}

	return r.Create(ctx, usage)
}

// GetLeaderboard retrieves the token usage leaderboard
func (r *TokenUsageRepository) GetLeaderboard(ctx context.Context, limit int) ([]*model.TokenLeaderboard, error) {
	query := `
		SELECT id, username, display_name, avatar_url,
			total_tokens, total_requests, input_tokens, output_tokens, estimated_cost, rank
		FROM token_leaderboard
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*model.TokenLeaderboard
	for rows.Next() {
		entry := &model.TokenLeaderboard{}
		err := rows.Scan(
			&entry.UserID, &entry.Username, &entry.DisplayName, &entry.AvatarURL,
			&entry.TotalTokens, &entry.TotalRequests, &entry.InputTokens, &entry.OutputTokens, &entry.EstimatedCost, &entry.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetUserStats retrieves token usage stats for a user
func (r *TokenUsageRepository) GetUserStats(ctx context.Context, userID uuid.UUID, periodStart, periodEnd time.Time) (*model.TokenUsage, error) {
	query := `
		SELECT
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(estimated_cost), 0) as estimated_cost
		FROM token_usage
		WHERE user_id = $1 AND period_start >= $2 AND period_end <= $3
	`

	usage := &model.TokenUsage{
		UserID:      userID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	err := r.db.QueryRowContext(ctx, query, userID, periodStart, periodEnd).Scan(
		&usage.InputTokens,
		&usage.OutputTokens,
		&usage.TotalTokens,
		&usage.EstimatedCost,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return usage, nil
}

// GetTotalTokens returns the sum of all tokens used
func (r *TokenUsageRepository) GetTotalTokens(ctx context.Context) (int64, error) {
	var total int64
	query := `SELECT COALESCE(SUM(total_tokens), 0) FROM token_usage`
	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
