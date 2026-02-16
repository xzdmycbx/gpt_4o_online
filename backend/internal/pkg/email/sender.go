package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"

	"gopkg.in/gomail.v2"
)

// Sender provides email sending functionality
type Sender struct {
	provider     string
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	fromEmail    string
	fromName     string
	resendAPIKey string
}

// NewSender creates a new email sender
func NewSender(provider, smtpHost string, smtpPort int, smtpUser, smtpPassword, fromEmail, fromName, resendAPIKey string) *Sender {
	return &Sender{
		provider:     provider,
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
		resendAPIKey: resendAPIKey,
	}
}

// SendEmail sends an email using the configured provider
func (s *Sender) SendEmail(to, subject, htmlBody, textBody string) error {
	if s.provider == "resend" {
		return s.sendWithResend(to, subject, htmlBody)
	}
	return s.sendWithSMTP(to, subject, htmlBody, textBody)
}

// sendWithSMTP sends email using SMTP
func (s *Sender) sendWithSMTP(to, subject, htmlBody, textBody string) error {
	m := gomail.NewMessage()

	if s.fromName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail))
	} else {
		m.SetHeader("From", s.fromEmail)
	}

	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	if htmlBody != "" {
		m.SetBody("text/html", htmlBody)
		if textBody != "" {
			m.AddAlternative("text/plain", textBody)
		}
	} else {
		m.SetBody("text/plain", textBody)
	}

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUser, s.smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

// sendWithResend sends email using Resend.com API
func (s *Sender) sendWithResend(to, subject, htmlBody string) error {
	type resendEmail struct {
		From    string `json:"from"`
		To      []string `json:"to"`
		Subject string `json:"subject"`
		HTML    string `json:"html"`
	}

	fromAddress := s.fromEmail
	if s.fromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	email := resendEmail{
		From:    fromAddress,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.resendAPIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email
// resetURL should already contain the token parameter
func (s *Sender) SendPasswordResetEmail(to, username, resetURL string) error {
	subject := "Reset Your Password - AI Chat"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2b5278;">Reset Your Password</h2>
        <p>Hi %s,</p>
        <p>We received a request to reset your password for your AI Chat account.</p>
        <p>Click the button below to reset your password:</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #2b5278; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Reset Password
            </a>
        </p>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">%s</p>
        <p>This link will expire in 1 hour.</p>
        <p>If you didn't request this password reset, you can safely ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">
        <p style="color: #666; font-size: 12px;">
            AI Chat - Intelligent Assistant<br>
            This is an automated message, please do not reply.
        </p>
    </div>
</body>
</html>
`, username, resetURL, resetURL)

	textBody := fmt.Sprintf(`
Reset Your Password

Hi %s,

We received a request to reset your password for your AI Chat account.

Click the link below to reset your password:
%s

This link will expire in 1 hour.

If you didn't request this password reset, you can safely ignore this email.

---
AI Chat - Intelligent Assistant
This is an automated message, please do not reply.
`, username, resetURL)

	return s.SendEmail(to, subject, htmlBody, textBody)
}

// SendVerificationEmail sends an email verification link
func (s *Sender) SendVerificationEmail(to, username, verificationToken, verificationURL string) error {
	subject := "Verify Your Email - AI Chat"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2b5278;">Welcome to AI Chat!</h2>
        <p>Hi %s,</p>
        <p>Thank you for signing up! Please verify your email address to complete your registration.</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s?token=%s" style="background-color: #2b5278; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Verify Email
            </a>
        </p>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">%s?token=%s</p>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">
        <p style="color: #666; font-size: 12px;">
            AI Chat - Intelligent Assistant<br>
            This is an automated message, please do not reply.
        </p>
    </div>
</body>
</html>
`, username, verificationURL, verificationToken, verificationURL, verificationToken)

	textBody := fmt.Sprintf(`
Welcome to AI Chat!

Hi %s,

Thank you for signing up! Please verify your email address to complete your registration.

Click the link below:
%s?token=%s

---
AI Chat - Intelligent Assistant
This is an automated message, please do not reply.
`, username, verificationURL, verificationToken)

	return s.SendEmail(to, subject, htmlBody, textBody)
}

// TestConnection tests the email configuration
func (s *Sender) TestConnection() error {
	if s.provider == "resend" {
		// For Resend, we can't really test without sending an email
		// Just validate that we have an API key
		if s.resendAPIKey == "" {
			return fmt.Errorf("resend API key is empty")
		}
		return nil
	}

	// Test SMTP connection
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return nil
}
