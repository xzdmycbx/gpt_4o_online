package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents user role types
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleUser       UserRole = "user"
)

// User represents a user account
type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       *string    `json:"email,omitempty" db:"email"`
	PasswordHash *string   `json:"-" db:"password_hash"`
	DisplayName *string    `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Role        UserRole   `json:"role" db:"role"`

	// OAuth2 fields
	OAuth2Provider     *string    `json:"oauth2_provider,omitempty" db:"oauth2_provider"`
	OAuth2ID           *string    `json:"oauth2_id,omitempty" db:"oauth2_id"`
	OAuth2AccessToken  *string    `json:"-" db:"oauth2_access_token"`
	OAuth2RefreshToken *string    `json:"-" db:"oauth2_refresh_token"`
	OAuth2TokenExpiry  *time.Time `json:"-" db:"oauth2_token_expiry"`

	// Status fields
	IsBanned       bool       `json:"is_banned" db:"is_banned"`
	BanReason      *string    `json:"ban_reason,omitempty" db:"ban_reason"`
	BannedAt       *time.Time `json:"banned_at,omitempty" db:"banned_at"`
	BannedBy       *uuid.UUID `json:"banned_by,omitempty" db:"banned_by"`

	// Rate limiting
	RateLimitExempt  bool  `json:"rate_limit_exempt" db:"rate_limit_exempt"`
	CustomRateLimit  *int  `json:"custom_rate_limit,omitempty" db:"custom_rate_limit"`

	// Timestamps
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// IsSuperAdmin checks if user is a super admin
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsAdmin checks if user is an admin (including super admin)
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// CanManageUser checks if this user can manage another user
func (u *User) CanManageUser(target *User) bool {
	// Super admin can manage anyone
	if u.IsSuperAdmin() {
		return true
	}

	// Regular admin cannot manage other admins or super admins
	if u.IsAdmin() && !u.IsSuperAdmin() {
		return target.Role == RoleUser
	}

	return false
}

// UserCreateRequest represents request to create a new user
type UserCreateRequest struct {
	Username    string  `json:"username" binding:"required,min=3,max=50"`
	Email       *string `json:"email" binding:"omitempty,email"`
	Password    string  `json:"password" binding:"required,min=8"`
	DisplayName *string `json:"display_name" binding:"omitempty,max=100"`
}

// UserUpdateRequest represents request to update user
type UserUpdateRequest struct {
	Email       *string `json:"email" binding:"omitempty,email"`
	DisplayName *string `json:"display_name" binding:"omitempty,max=100"`
	AvatarURL   *string `json:"avatar_url" binding:"omitempty,url"`
}

// UserLoginRequest represents login request
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents user data for API response
type UserResponse struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	Email            *string    `json:"email,omitempty"`
	DisplayName      *string    `json:"display_name,omitempty"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	Role             UserRole   `json:"role"`
	IsBanned         bool       `json:"is_banned"`
	RateLimitExempt  bool       `json:"rate_limit_exempt"`
	CustomRateLimit  *int       `json:"custom_rate_limit,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		DisplayName:     u.DisplayName,
		AvatarURL:       u.AvatarURL,
		Role:            u.Role,
		IsBanned:        u.IsBanned,
		RateLimitExempt: u.RateLimitExempt,
		CustomRateLimit: u.CustomRateLimit,
		CreatedAt:       u.CreatedAt,
	}
}
