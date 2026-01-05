// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
)

// SessionRepository implements the repository.SessionRepository interface for PostgreSQL.
type SessionRepository struct {
	db *DB
}

// NewSessionRepository creates a new PostgreSQL session repository.
func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new session into the database.
func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
		session.CreatedAt,
	)

	return err
}

// GetByID retrieves a session by its ID.
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = $1
	`

	session := &domain.Session{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return session, nil
}

// Delete removes a session by its ID.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// DeleteByUserID removes all sessions for a user.
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	return err
}

// DeleteExpired removes all expired sessions.
func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.Pool.Exec(ctx, query)
	return err
}
