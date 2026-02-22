package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
)

// AIProviderRepository handles AI provider data access
type AIProviderRepository struct {
	db *sql.DB
}

// NewAIProviderRepository creates a new AI provider repository
func NewAIProviderRepository(db *sql.DB) *AIProviderRepository {
	return &AIProviderRepository{db: db}
}

// Create creates a new AI provider
func (r *AIProviderRepository) Create(ctx context.Context, provider *model.AIProvider) error {
	query := `
		INSERT INTO ai_providers (
			id, name, display_name, provider_type, api_endpoint, api_key_encrypted,
			is_active, description, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		provider.ID, provider.Name, provider.DisplayName, provider.ProviderType,
		provider.APIEndpoint, provider.APIKeyEncrypted,
		provider.IsActive, provider.Description, provider.CreatedBy,
	).Scan(&provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create AI provider: %w", err)
	}

	return nil
}

// GetByID retrieves an AI provider by ID
func (r *AIProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AIProvider, error) {
	query := `
		SELECT id, name, display_name, provider_type, api_endpoint, api_key_encrypted,
			is_active, description, created_by, created_at, updated_at
		FROM ai_providers WHERE id = $1
	`

	p := &model.AIProvider{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.DisplayName, &p.ProviderType, &p.APIEndpoint, &p.APIKeyEncrypted,
		&p.IsActive, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("AI provider not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	return p, nil
}

// List retrieves all AI providers with model counts
func (r *AIProviderRepository) List(ctx context.Context, activeOnly bool) ([]*model.AIProvider, error) {
	query := `
		SELECT p.id, p.name, p.display_name, p.provider_type, p.api_endpoint, p.api_key_encrypted,
			p.is_active, p.description, p.created_by, p.created_at, p.updated_at,
			COUNT(m.id) AS model_count
		FROM ai_providers p
		LEFT JOIN ai_models m ON m.provider_id = p.id
	`

	if activeOnly {
		query += " WHERE p.is_active = true"
	}

	query += " GROUP BY p.id ORDER BY p.display_name ASC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list AI providers: %w", err)
	}
	defer rows.Close()

	var providers []*model.AIProvider
	for rows.Next() {
		p := &model.AIProvider{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.DisplayName, &p.ProviderType, &p.APIEndpoint, &p.APIKeyEncrypted,
			&p.IsActive, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
			&p.ModelCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AI provider: %w", err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}

// Update updates an AI provider
func (r *AIProviderRepository) Update(ctx context.Context, provider *model.AIProvider) error {
	query := `
		UPDATE ai_providers SET
			display_name = $2,
			provider_type = $3,
			api_endpoint = $4,
			api_key_encrypted = $5,
			is_active = $6,
			description = $7,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx, query,
		provider.ID, provider.DisplayName, provider.ProviderType, provider.APIEndpoint,
		provider.APIKeyEncrypted, provider.IsActive, provider.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update AI provider: %w", err)
	}

	return nil
}

// Delete deletes an AI provider
func (r *AIProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ai_providers WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CountModels returns the number of models linked to a provider
func (r *AIProviderRepository) CountModels(ctx context.Context, providerID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM ai_models WHERE provider_id = $1", providerID,
	).Scan(&count)
	return count, err
}
