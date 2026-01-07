// Package service implements the business logic layer.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// authService implements the AuthService interface.
type authService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	emailService EmailService
}

// NewAuthService creates a new auth service.
func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, emailService EmailService) AuthService {
	return &authService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		emailService: emailService,
	}
}

// Register creates a new user account.
func (s *authService) Register(ctx context.Context, input *domain.RegisterInput, ip, userAgent string) (*domain.User, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrConflict
	}

	// Hash password
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Determine role (first user becomes super_admin)
	role := domain.RoleUser
	count, err := s.userRepo.Count(ctx)
	if err == nil && count == 0 {
		role = domain.RoleSuperAdmin
	}

	// Create user
	user := domain.NewUser(input.Email, input.Name, passwordHash, role)

	// Generate verification token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	user.VerificationToken = &token
	expiresAt := time.Now().Add(24 * time.Hour)
	user.VerificationTokenExpiresAt = &expiresAt

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send verification email
	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.emailService.SendVerificationEmail(sendCtx, user.Email, user.Name, *user.VerificationToken); err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}()

	return user, nil
}

// Login authenticates a user and creates a session.
func (s *authService) Login(ctx context.Context, input *domain.LoginInput, ip, userAgent string) (*domain.User, *domain.Session, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, nil, domain.ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Check if email is verified
	if !user.EmailVerified {
		// Generate new verification token
		token, err := generateToken()
		if err != nil {
			return nil, nil, err
		}
		user.VerificationToken = &token
		expiresAt := time.Now().Add(24 * time.Hour)
		user.VerificationTokenExpiresAt = &expiresAt

		// Update user with new token
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, nil, err
		}

		// Send verification email
		// Use a goroutine so we don't block the login response
		go func() {
			sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.emailService.SendVerificationEmail(sendCtx, user.Email, user.Name, token); err != nil {
				// Log error (in a real app, use a logger)
				fmt.Printf("Failed to send verification email: %v\n", err)
			}
		}()

		return nil, nil, domain.ErrEmailNotVerified
	}

	// Create session
	session := domain.NewSession(user.ID, ip, userAgent)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	return user, session, nil
}

// Logout destroys a user session.
func (s *authService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// ValidateSession checks if a session is valid and returns the user.
func (s *authService) ValidateSession(ctx context.Context, sessionID string) (*domain.User, error) {
	if sessionID == "" {
		return nil, domain.ErrUnauthorized
	}

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	// Check if expired
	if session.IsExpired() {
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, domain.ErrSessionExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetCurrentUser retrieves the authenticated user from session.
func (s *authService) GetCurrentUser(ctx context.Context, sessionID string) (*domain.User, error) {
	return s.ValidateSession(ctx, sessionID)
}

// VerifyEmail verifies a user's email address using a token.
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// Simple lookup (in a real app, define GetByVerificationToken)
	// For this starter, we might need to add GetByVerificationToken to UserRepository or search
	// Since we don't have GetByVerificationToken yet, I'll add it to UserRepository first.
	// Wait, I should have planned that. Let me look at UserRepository again.

	// Assuming I add GetByVerificationToken to UserRepo
	user, err := s.userRepo.GetByVerificationToken(ctx, token)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return domain.ErrInvalidToken
		}
		return err
	}

	if user.EmailVerified {
		return nil // Already verified
	}

	if user.VerificationTokenExpiresAt != nil && time.Now().After(*user.VerificationTokenExpiresAt) {
		return domain.ErrTokenExpired
	}

	// Update user
	user.EmailVerified = true
	user.VerificationToken = nil
	user.VerificationTokenExpiresAt = nil

	return s.userRepo.Update(ctx, user)
}

// hashPassword creates a bcrypt hash of the password.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword compares a password with a hash.
func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken creates a random token string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
