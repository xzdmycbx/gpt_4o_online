package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/crypto"
	"github.com/ai-chat/backend/internal/pkg/jwt"
	"github.com/ai-chat/backend/internal/pkg/oauth2"
	"github.com/ai-chat/backend/internal/repository"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo       *repository.UserRepository
	jwtManager     *jwt.Manager
	oauth2Twitter  *oauth2.TwitterOAuth2Client
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	jwtManager *jwt.Manager,
	oauth2Twitter *oauth2.TwitterOAuth2Client,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		jwtManager:    jwtManager,
		oauth2Twitter: oauth2Twitter,
	}
}

// Login authenticates a user with username/password
func (s *AuthService) Login(ctx context.Context, req *model.UserLoginRequest) (string, *model.User, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is banned
	if user.IsBanned {
		return "", nil, fmt.Errorf("account is banned: %s", *user.BanReason)
	}

	// Verify password
	if user.PasswordHash == nil {
		return "", nil, fmt.Errorf("password not set for this account")
	}

	if !crypto.CheckPasswordHash(req.Password, *user.PasswordHash) {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *model.UserCreateRequest) (string, *model.User, error) {
	// Check if username already exists
	existingUser, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return "", nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if req.Email != nil {
		existingUser, _ = s.userRepo.GetByEmail(ctx, *req.Email)
		if existingUser != nil {
			return "", nil, fmt.Errorf("email already exists")
		}
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: &passwordHash,
		DisplayName:  req.DisplayName,
		Role:         model.RoleUser,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// GenerateTwitterOAuth2URL generates Twitter OAuth2 authorization URL
func (s *AuthService) GenerateTwitterOAuth2URL(state string) (string, string, error) {
	if s.oauth2Twitter == nil {
		return "", "", fmt.Errorf("Twitter OAuth2 is not configured")
	}
	return s.oauth2Twitter.GenerateAuthURL(state)
}

// HandleTwitterOAuth2Callback handles Twitter OAuth2 callback
func (s *AuthService) HandleTwitterOAuth2Callback(ctx context.Context, code, codeVerifier string) (string, *model.User, error) {
	if s.oauth2Twitter == nil {
		return "", nil, fmt.Errorf("Twitter OAuth2 is not configured")
	}

	// Exchange code for token
	token, err := s.oauth2Twitter.ExchangeCode(ctx, code, codeVerifier)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Twitter
	userInfo, err := s.oauth2Twitter.GetUserInfo(ctx, token)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if user already exists with this OAuth2 ID
	user, err := s.userRepo.GetByOAuth2(ctx, "twitter", userInfo.ID)
	if err != nil {
		// User doesn't exist, create new user
		user = &model.User{
			ID:                 uuid.New(),
			Username:           userInfo.Username,
			DisplayName:        &userInfo.Name,
			AvatarURL:          &userInfo.ProfileImageURL,
			Role:               model.RoleUser,
			OAuth2Provider:     func() *string { p := "twitter"; return &p }(),
			OAuth2ID:           &userInfo.ID,
			OAuth2AccessToken:  &token.AccessToken,
			OAuth2RefreshToken: &token.RefreshToken,
			OAuth2TokenExpiry:  &token.Expiry,
		}

		// Check if username already exists and append random suffix if needed
		if existingUser, _ := s.userRepo.GetByUsername(ctx, user.Username); existingUser != nil {
			user.Username = fmt.Sprintf("%s_%s", user.Username, uuid.New().String()[:8])
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists, update OAuth2 tokens and profile info
		user.OAuth2AccessToken = &token.AccessToken
		user.OAuth2RefreshToken = &token.RefreshToken
		user.OAuth2TokenExpiry = &token.Expiry
		user.DisplayName = &userInfo.Name
		user.AvatarURL = &userInfo.ProfileImageURL

		if err := s.userRepo.Update(ctx, user); err != nil {
			return "", nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Check if user is banned
	if user.IsBanned {
		return "", nil, fmt.Errorf("account is banned: %s", *user.BanReason)
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate JWT token
	jwtToken, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return jwtToken, user, nil
}

// RefreshToken refreshes an expired JWT token
func (s *AuthService) RefreshToken(ctx context.Context, oldToken string) (string, error) {
	return s.jwtManager.RefreshToken(oldToken)
}

// GetCurrentUser retrieves the current authenticated user
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if user has a password (not OAuth2 only)
	if user.PasswordHash == nil {
		return fmt.Errorf("password change not available for OAuth2 accounts")
	}

	// Verify current password
	if !crypto.CheckPasswordHash(currentPassword, *user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password")
	}

	// Update password
	user.PasswordHash = &hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password")
	}

	return nil
}
