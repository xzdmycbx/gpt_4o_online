package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ai-chat/backend/internal/model"
	"github.com/ai-chat/backend/internal/pkg/crypto"
	"github.com/ai-chat/backend/internal/pkg/email"
	"github.com/ai-chat/backend/internal/repository"
)

// EmailService handles email-related operations
type EmailService struct {
	sender           *email.Sender
	userRepo         *repository.UserRepository
	resetTokenRepo   *repository.PasswordResetTokenRepository
	baseURL          string
}

// NewEmailService creates a new email service
func NewEmailService(sender *email.Sender, userRepo *repository.UserRepository, resetTokenRepo *repository.PasswordResetTokenRepository, baseURL string) *EmailService {
	return &EmailService{
		sender:         sender,
		userRepo:       userRepo,
		resetTokenRepo: resetTokenRepo,
		baseURL:        baseURL,
	}
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email string) error {
	// Check if email sender is configured
	if s.sender == nil {
		return fmt.Errorf("email service is not configured")
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Delete any existing reset tokens for this user
	_ = s.resetTokenRepo.DeleteByUserID(ctx, user.ID)

	// Generate reset token (random string that user will receive via email)
	token, err := crypto.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token before storing (only hash is stored in database)
	tokenHash := crypto.HashToken(token)

	// Create password reset token in database with hashed token
	resetToken := &model.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash, // Store hash, not plaintext
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token valid for 1 hour
		CreatedAt: time.Now(),
	}

	if err := s.resetTokenRepo.Create(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// Send email with plaintext token (user needs this to reset password)
	// resetURL already contains the token parameter
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)
	return s.sender.SendPasswordResetEmail(*user.Email, user.Username, resetURL)
}

// VerifyPasswordResetToken verifies a password reset token
func (s *EmailService) VerifyPasswordResetToken(ctx context.Context, token string) (*model.User, *model.PasswordResetToken, error) {
	// Hash the provided token to look up in database
	tokenHash := crypto.HashToken(token)

	// Retrieve token from database by hash
	resetToken, err := s.resetTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid or expired token")
	}

	// Check if token is valid (not expired and not used)
	if !resetToken.IsValid() {
		if resetToken.IsExpired() {
			return nil, nil, fmt.Errorf("token has expired")
		}
		if resetToken.IsUsed() {
			return nil, nil, fmt.Errorf("token has already been used")
		}
		return nil, nil, fmt.Errorf("invalid token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	return user, resetToken, nil
}

// ResetPassword resets a user's password
func (s *EmailService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Verify token and get user
	user, resetToken, err := s.VerifyPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Hash new password
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.PasswordHash = &passwordHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.resetTokenRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// SendVerificationEmail sends an email verification email
func (s *EmailService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// Check if email sender is configured
	if s.sender == nil {
		return fmt.Errorf("email service is not configured")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.Email == nil {
		return fmt.Errorf("user has no email address")
	}

	// Generate verification token
	token, err := crypto.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// TODO: Store token in database

	// Send email
	verificationURL := fmt.Sprintf("%s/verify-email", s.baseURL)
	return s.sender.SendVerificationEmail(*user.Email, user.Username, token, verificationURL)
}

// VerifyEmail verifies an email address
func (s *EmailService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Verify token and mark email as verified
	return fmt.Errorf("email verification not implemented")
}
