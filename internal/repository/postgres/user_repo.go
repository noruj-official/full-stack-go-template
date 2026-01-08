// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
)

// UserRepository implements the repository.UserRepository interface for PostgreSQL.
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, role, status, created_at, updated_at, email_verified, verification_token, verification_token_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
		user.EmailVerified,
		user.VerificationToken,
		user.VerificationTokenExpiresAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}

	return nil
}

// GetByID retrieves a user by their unique identifier.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, status, created_at, updated_at, email_verified, verification_token, verification_token_expires_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.VerificationTokenExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, status, created_at, updated_at, email_verified, verification_token, verification_token_expires_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.VerificationTokenExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByVerificationToken retrieves a user by their verification token.
func (r *UserRepository) GetByVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, status, created_at, updated_at, email_verified, verification_token, verification_token_expires_at
		FROM users
		WHERE verification_token = $1
	`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.VerificationTokenExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

// List retrieves all users with pagination.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, status, created_at, updated_at, email_verified, verification_token, verification_token_expires_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.EmailVerified,
			&user.VerificationToken,
			&user.VerificationTokenExpiresAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Update modifies an existing user in the database.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = $2, name = $3, password_hash = $4, role = $5, status = $6, updated_at = $7, email_verified = $8, verification_token = $9, verification_token_expires_at = $10
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.UpdatedAt,
		user.EmailVerified,
		user.VerificationToken,
		user.VerificationTokenExpiresAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete removes a user from the database.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique")
}
