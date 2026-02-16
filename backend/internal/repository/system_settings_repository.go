package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ai-chat/backend/internal/model"
)

// SystemSettingsRepository handles system settings database operations
type SystemSettingsRepository struct {
	db *sql.DB
}

// NewSystemSettingsRepository creates a new system settings repository
func NewSystemSettingsRepository(db *sql.DB) *SystemSettingsRepository {
	return &SystemSettingsRepository{db: db}
}

// GetByKey retrieves a setting by key
func (r *SystemSettingsRepository) GetByKey(ctx context.Context, key string) (*model.SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, value_type, created_at, updated_at
		FROM system_settings
		WHERE setting_key = $1
	`

	var setting model.SystemSetting
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&setting.ID,
		&setting.SettingKey,
		&setting.SettingValue,
		&setting.Description,
		&setting.ValueType,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("setting not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	return &setting, nil
}

// GetAll retrieves all settings
func (r *SystemSettingsRepository) GetAll(ctx context.Context) ([]model.SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, value_type, created_at, updated_at
		FROM system_settings
		ORDER BY setting_key
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	var settings []model.SystemSetting
	for rows.Next() {
		var setting model.SystemSetting
		err := rows.Scan(
			&setting.ID,
			&setting.SettingKey,
			&setting.SettingValue,
			&setting.Description,
			&setting.ValueType,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

// Update updates a setting value
func (r *SystemSettingsRepository) Update(ctx context.Context, key, value string) error {
	query := `
		UPDATE system_settings
		SET setting_value = $2, updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = $1
	`

	result, err := r.db.ExecContext(ctx, query, key, value)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("setting not found: %s", key)
	}

	return nil
}

// UpdateMultiple updates multiple settings in a transaction
func (r *SystemSettingsRepository) UpdateMultiple(ctx context.Context, updates map[string]string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE system_settings
		SET setting_value = $2, updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = $1
	`

	for key, value := range updates {
		_, err := tx.ExecContext(ctx, query, key, value)
		if err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
