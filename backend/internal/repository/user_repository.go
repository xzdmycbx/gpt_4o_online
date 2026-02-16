package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
)

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, display_name, avatar_url, role,
			oauth2_provider, oauth2_id, oauth2_access_token, oauth2_refresh_token, oauth2_token_expiry
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.AvatarURL, user.Role,
		user.OAuth2Provider, user.OAuth2ID, user.OAuth2AccessToken, user.OAuth2RefreshToken, user.OAuth2TokenExpiry,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, role,
			oauth2_provider, oauth2_id, oauth2_access_token, oauth2_refresh_token, oauth2_token_expiry,
			is_banned, ban_reason, banned_at, banned_by,
			rate_limit_exempt, custom_rate_limit,
			email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.Role,
		&user.OAuth2Provider, &user.OAuth2ID, &user.OAuth2AccessToken, &user.OAuth2RefreshToken, &user.OAuth2TokenExpiry,
		&user.IsBanned, &user.BanReason, &user.BannedAt, &user.BannedBy,
		&user.RateLimitExempt, &user.CustomRateLimit,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, role,
			oauth2_provider, oauth2_id, oauth2_access_token, oauth2_refresh_token, oauth2_token_expiry,
			is_banned, ban_reason, banned_at, banned_by,
			rate_limit_exempt, custom_rate_limit,
			email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE username = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.Role,
		&user.OAuth2Provider, &user.OAuth2ID, &user.OAuth2AccessToken, &user.OAuth2RefreshToken, &user.OAuth2TokenExpiry,
		&user.IsBanned, &user.BanReason, &user.BannedAt, &user.BannedBy,
		&user.RateLimitExempt, &user.CustomRateLimit,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, role,
			oauth2_provider, oauth2_id, oauth2_access_token, oauth2_refresh_token, oauth2_token_expiry,
			is_banned, ban_reason, banned_at, banned_by,
			rate_limit_exempt, custom_rate_limit,
			email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE email = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.Role,
		&user.OAuth2Provider, &user.OAuth2ID, &user.OAuth2AccessToken, &user.OAuth2RefreshToken, &user.OAuth2TokenExpiry,
		&user.IsBanned, &user.BanReason, &user.BannedAt, &user.BannedBy,
		&user.RateLimitExempt, &user.CustomRateLimit,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByOAuth2 retrieves a user by OAuth2 provider and ID
func (r *UserRepository) GetByOAuth2(ctx context.Context, provider, oauth2ID string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, role,
			oauth2_provider, oauth2_id, oauth2_access_token, oauth2_refresh_token, oauth2_token_expiry,
			is_banned, ban_reason, banned_at, banned_by,
			rate_limit_exempt, custom_rate_limit,
			email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE oauth2_provider = $1 AND oauth2_id = $2
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, provider, oauth2ID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.Role,
		&user.OAuth2Provider, &user.OAuth2ID, &user.OAuth2AccessToken, &user.OAuth2RefreshToken, &user.OAuth2TokenExpiry,
		&user.IsBanned, &user.BanReason, &user.BannedAt, &user.BannedBy,
		&user.RateLimitExempt, &user.CustomRateLimit,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET
			email = $2,
			password_hash = $3,
			display_name = $4,
			avatar_url = $5,
			role = $6,
			oauth2_access_token = $7,
			oauth2_refresh_token = $8,
			oauth2_token_expiry = $9,
			is_banned = $10,
			ban_reason = $11,
			banned_at = $12,
			banned_by = $13,
			rate_limit_exempt = $14,
			custom_rate_limit = $15,
			email_verified_at = $16,
			last_login_at = $17
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx, query,
		user.ID, user.Email, user.PasswordHash, user.DisplayName, user.AvatarURL, user.Role,
		user.OAuth2AccessToken, user.OAuth2RefreshToken, user.OAuth2TokenExpiry,
		user.IsBanned, user.BanReason, user.BannedAt, user.BannedBy,
		user.RateLimitExempt, user.CustomRateLimit,
		user.EmailVerifiedAt, user.LastLoginAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, role,
			oauth2_provider, oauth2_id,
			is_banned, ban_reason, banned_at, banned_by,
			rate_limit_exempt, custom_rate_limit,
			email_verified_at, last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.Role,
			&user.OAuth2Provider, &user.OAuth2ID,
			&user.IsBanned, &user.BanReason, &user.BannedAt, &user.BannedBy,
			&user.RateLimitExempt, &user.CustomRateLimit,
			&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CountActiveUsers counts users who have sent messages in the last N days
func (r *UserRepository) CountActiveUsers(ctx context.Context, days int) (int, error) {
	var count int
	query := `
		SELECT COUNT(DISTINCT c.user_id)
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE m.created_at >= NOW() - $1 * INTERVAL '1 day'
	`
	err := r.db.QueryRowContext(ctx, query, days).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
