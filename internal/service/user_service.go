// Package service implements the business logic layer.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// userService implements the UserService interface.
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with validation (admin action).
func (s *userService) CreateUser(ctx context.Context, input *domain.CreateUserInput) (*domain.User, error) {
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
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := domain.NewUser(input.Email, input.Name, string(passwordHash), input.Role)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID.
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// ListUsers retrieves all users with pagination.
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	users, err := s.userRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser updates an existing user.
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, input *domain.UpdateUserInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Role != nil {
		user.Role = *input.Role
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser removes a user.
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}
