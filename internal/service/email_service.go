package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/templates/email"
)

type resendEmailService struct {
	apiKey    string
	fromEmail string
	appURL    string
}

// NewResendEmailService creates a new email service using Resend.
func NewResendEmailService(apiKey, fromEmail, appURL string) EmailService {
	return &resendEmailService{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		appURL:    appURL,
	}
}

// SendVerificationEmail sends a verification email to the user.
func (s *resendEmailService) SendVerificationEmail(ctx context.Context, emailAddr, name, token string) error {
	if s.apiKey == "" {
		fmt.Printf("[MOCK EMAIL] To: %s, Token: %s\n", emailAddr, token)
		return nil
	}

	url := "https://api.resend.com/emails"

	// Create verification link
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, token)

	htmlContent := email.GetVerificationEmailContent(name, verificationLink)

	payload := map[string]interface{}{
		"from":    s.fromEmail,
		"to":      []string{emailAddr},
		"subject": "Verify your email address",
		"html":    htmlContent,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("failed to send email via Resend")
	}

	return nil
}
