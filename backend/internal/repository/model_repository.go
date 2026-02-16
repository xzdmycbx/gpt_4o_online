package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
)

// AIModelRepository handles AI model data access
type AIModelRepository struct {
	db *sql.DB
}

// NewAIModelRepository creates a new AI model repository
func NewAIModelRepository(db *sql.DB) *AIModelRepository {
	return &AIModelRepository{db: db}
}

// Create creates a new AI model
func (r *AIModelRepository) Create(ctx context.Context, aiModel *model.AIModel) error {
	query := `
		INSERT INTO ai_models (
			id, name, display_name, provider, api_endpoint, api_key_encrypted, model_identifier,
			supports_streaming, supports_functions, max_tokens,
			input_price_per_1k, output_price_per_1k,
			is_active, is_default, description, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		aiModel.ID, aiModel.Name, aiModel.DisplayName, aiModel.Provider, aiModel.APIEndpoint,
		aiModel.APIKeyEncrypted, aiModel.ModelIdentifier,
		aiModel.SupportsStreaming, aiModel.SupportsFunctions, aiModel.MaxTokens,
		aiModel.InputPricePer1k, aiModel.OutputPricePer1k,
		aiModel.IsActive, aiModel.IsDefault, aiModel.Description, aiModel.CreatedBy,
	).Scan(&aiModel.CreatedAt, &aiModel.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create AI model: %w", err)
	}

	return nil
}

// GetByID retrieves an AI model by ID
func (r *AIModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AIModel, error) {
	query := `
		SELECT id, name, display_name, provider, api_endpoint, api_key_encrypted, model_identifier,
			supports_streaming, supports_functions, max_tokens,
			input_price_per_1k, output_price_per_1k,
			is_active, is_default, description, created_by, created_at, updated_at
		FROM ai_models WHERE id = $1
	`

	aiModel := &model.AIModel{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&aiModel.ID, &aiModel.Name, &aiModel.DisplayName, &aiModel.Provider, &aiModel.APIEndpoint,
		&aiModel.APIKeyEncrypted, &aiModel.ModelIdentifier,
		&aiModel.SupportsStreaming, &aiModel.SupportsFunctions, &aiModel.MaxTokens,
		&aiModel.InputPricePer1k, &aiModel.OutputPricePer1k,
		&aiModel.IsActive, &aiModel.IsDefault, &aiModel.Description, &aiModel.CreatedBy,
		&aiModel.CreatedAt, &aiModel.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("AI model not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get AI model: %w", err)
	}

	return aiModel, nil
}

// GetDefault retrieves the default AI model
func (r *AIModelRepository) GetDefault(ctx context.Context) (*model.AIModel, error) {
	query := `
		SELECT id, name, display_name, provider, api_endpoint, api_key_encrypted, model_identifier,
			supports_streaming, supports_functions, max_tokens,
			input_price_per_1k, output_price_per_1k,
			is_active, is_default, description, created_by, created_at, updated_at
		FROM ai_models WHERE is_default = true AND is_active = true LIMIT 1
	`

	aiModel := &model.AIModel{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&aiModel.ID, &aiModel.Name, &aiModel.DisplayName, &aiModel.Provider, &aiModel.APIEndpoint,
		&aiModel.APIKeyEncrypted, &aiModel.ModelIdentifier,
		&aiModel.SupportsStreaming, &aiModel.SupportsFunctions, &aiModel.MaxTokens,
		&aiModel.InputPricePer1k, &aiModel.OutputPricePer1k,
		&aiModel.IsActive, &aiModel.IsDefault, &aiModel.Description, &aiModel.CreatedBy,
		&aiModel.CreatedAt, &aiModel.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no default AI model configured")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default AI model: %w", err)
	}

	return aiModel, nil
}

// List retrieves all AI models
func (r *AIModelRepository) List(ctx context.Context, activeOnly bool) ([]*model.AIModel, error) {
	query := `
		SELECT id, name, display_name, provider, api_endpoint, api_key_encrypted, model_identifier,
			supports_streaming, supports_functions, max_tokens,
			input_price_per_1k, output_price_per_1k,
			is_active, is_default, description, created_by, created_at, updated_at
		FROM ai_models
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY is_default DESC, display_name ASC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list AI models: %w", err)
	}
	defer rows.Close()

	var models []*model.AIModel
	for rows.Next() {
		aiModel := &model.AIModel{}
		err := rows.Scan(
			&aiModel.ID, &aiModel.Name, &aiModel.DisplayName, &aiModel.Provider, &aiModel.APIEndpoint,
			&aiModel.APIKeyEncrypted, &aiModel.ModelIdentifier,
			&aiModel.SupportsStreaming, &aiModel.SupportsFunctions, &aiModel.MaxTokens,
			&aiModel.InputPricePer1k, &aiModel.OutputPricePer1k,
			&aiModel.IsActive, &aiModel.IsDefault, &aiModel.Description, &aiModel.CreatedBy,
			&aiModel.CreatedAt, &aiModel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AI model: %w", err)
		}
		models = append(models, aiModel)
	}

	return models, nil
}

// Update updates an AI model
func (r *AIModelRepository) Update(ctx context.Context, aiModel *model.AIModel) error {
	query := `
		UPDATE ai_models SET
			display_name = $2,
			api_endpoint = $3,
			api_key_encrypted = $4,
			supports_streaming = $5,
			supports_functions = $6,
			max_tokens = $7,
			input_price_per_1k = $8,
			output_price_per_1k = $9,
			is_active = $10,
			description = $11
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx, query,
		aiModel.ID, aiModel.DisplayName, aiModel.APIEndpoint, aiModel.APIKeyEncrypted,
		aiModel.SupportsStreaming, aiModel.SupportsFunctions, aiModel.MaxTokens,
		aiModel.InputPricePer1k, aiModel.OutputPricePer1k,
		aiModel.IsActive, aiModel.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update AI model: %w", err)
	}

	return nil
}

// SetDefault sets a model as the default (unsets others)
func (r *AIModelRepository) SetDefault(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset all defaults
	_, err = tx.ExecContext(ctx, "UPDATE ai_models SET is_default = false")
	if err != nil {
		return fmt.Errorf("failed to unset defaults: %w", err)
	}

	// Set new default
	_, err = tx.ExecContext(ctx, "UPDATE ai_models SET is_default = true WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return tx.Commit()
}

// Delete deletes an AI model
func (r *AIModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ai_models WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UserSettingsRepository handles user settings data access
type UserSettingsRepository struct {
	db *sql.DB
}

// NewUserSettingsRepository creates a new user settings repository
func NewUserSettingsRepository(db *sql.DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

// GetByUserID retrieves user settings by user ID
func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.UserSettings, error) {
	query := `
		SELECT id, user_id, theme, font_size, language,
			notifications_enabled, notification_sound,
			default_model_id, stream_response, show_token_count,
			advanced_settings, device_id, last_synced_at,
			created_at, updated_at
		FROM user_settings WHERE user_id = $1
	`

	settings := &model.UserSettings{}
	var advancedJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.ID, &settings.UserID, &settings.Theme, &settings.FontSize, &settings.Language,
		&settings.NotificationsEnabled, &settings.NotificationSound,
		&settings.DefaultModelID, &settings.StreamResponse, &settings.ShowTokenCount,
		&advancedJSON, &settings.DeviceID, &settings.LastSyncedAt,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Unmarshal advanced settings
	if len(advancedJSON) > 0 {
		if err := json.Unmarshal(advancedJSON, &settings.AdvancedSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal advanced settings: %w", err)
		}
	}

	return settings, nil
}

// Upsert creates or updates user settings
func (r *UserSettingsRepository) Upsert(ctx context.Context, settings *model.UserSettings) error {
	advancedJSON, err := json.Marshal(settings.AdvancedSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal advanced settings: %w", err)
	}

	query := `
		INSERT INTO user_settings (
			id, user_id, theme, font_size, language,
			notifications_enabled, notification_sound,
			default_model_id, stream_response, show_token_count,
			advanced_settings, device_id, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			theme = EXCLUDED.theme,
			font_size = EXCLUDED.font_size,
			language = EXCLUDED.language,
			notifications_enabled = EXCLUDED.notifications_enabled,
			notification_sound = EXCLUDED.notification_sound,
			default_model_id = EXCLUDED.default_model_id,
			stream_response = EXCLUDED.stream_response,
			show_token_count = EXCLUDED.show_token_count,
			advanced_settings = EXCLUDED.advanced_settings,
			device_id = EXCLUDED.device_id,
			last_synced_at = NOW()
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
		settings.ID, settings.UserID, settings.Theme, settings.FontSize, settings.Language,
		settings.NotificationsEnabled, settings.NotificationSound,
		settings.DefaultModelID, settings.StreamResponse, settings.ShowTokenCount,
		advancedJSON, settings.DeviceID,
	).Scan(&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert user settings: %w", err)
	}

	return nil
}

// Delete deletes user settings
func (r *UserSettingsRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_settings WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
