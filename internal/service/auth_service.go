// Package service implements the business logic layer.
package service

import (
	"context"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// authService implements the AuthService interface.
type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

// NewAuthService creates a new auth service.
func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// Register creates a new user account.
func (s *authService) Register(ctx context.Context, input *domain.RegisterInput) (*domain.User, *domain.Session, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}

	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, nil, domain.ErrConflict
	}

	// Hash password
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, nil, err
	}

	// Determine role (first user becomes super_admin)
	role := domain.RoleUser
	count, err := s.userRepo.Count(ctx)
	if err == nil && count == 0 {
		role = domain.RoleSuperAdmin
	}

	// Create user
	user := domain.NewUser(input.Email, input.Name, passwordHash, role)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Create session
	session := domain.NewSession(user.ID)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	return user, session, nil
}

// Login authenticates a user and creates a session.
func (s *authService) Login(ctx context.Context, input *domain.LoginInput) (*domain.User, *domain.Session, error) {
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

	// Verify password
	if !checkPassword(user.PasswordHash, input.Password) {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Create session
	session := domain.NewSession(user.ID)
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
