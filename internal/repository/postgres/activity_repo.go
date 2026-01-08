// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

// ActivityLogRepository handles activity log data operations.
type ActivityLogRepository struct {
	db *DB
}

// NewActivityLogRepository creates a new activity log repository.
func NewActivityLogRepository(db *DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

// Create creates a new activity log entry.
func (r *ActivityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (user_id, activity_type, description, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		log.UserID,
		log.ActivityType,
		log.Description,
		log.IPAddress,
		log.UserAgent,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create activity log: %w", err)
	}

	return nil
}

// ListByUser retrieves activity logs for a specific user.
func (r *ActivityLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.ActivityLog, error) {
	query := `
		SELECT id, user_id, activity_type, description, ip_address, user_agent, created_at
		FROM activity_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query activity logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.ActivityLog
	for rows.Next() {
		log := &domain.ActivityLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.ActivityType,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// AuditLogRepository handles audit log data operations.
type AuditLogRepository struct {
	db *DB
}

// NewAuditLogRepository creates a new audit log repository.
func NewAuditLogRepository(db *DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry.
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	var oldValuesJSON, newValuesJSON []byte
	var err error

	if log.OldValues != nil {
		oldValuesJSON, err = json.Marshal(log.OldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old values: %w", err)
		}
	}

	if log.NewValues != nil {
		newValuesJSON, err = json.Marshal(log.NewValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new values: %w", err)
		}
	}

	query := `
		INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, old_values, new_values, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err = r.db.Pool.QueryRow(
		ctx,
		query,
		log.AdminID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		oldValuesJSON,
		newValuesJSON,
		log.IPAddress,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// List retrieves audit logs with pagination.
func (r *AuditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT a.id, a.admin_id, u.name as admin_name, a.action, a.resource_type, 
		       a.resource_id, a.old_values, a.new_values, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON a.admin_id = u.id
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ListByAdmin retrieves audit logs for a specific admin.
func (r *AuditLogRepository) ListByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT a.id, a.admin_id, u.name as admin_name, a.action, a.resource_type, 
		       a.resource_id, a.old_values, a.new_values, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON a.admin_id = u.id
		WHERE a.admin_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, adminID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// Count returns the total number of audit logs.
func (r *AuditLogRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM audit_logs`

	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// scanAuditLogs is a helper function to scan audit log rows.
func (r *AuditLogRepository) scanAuditLogs(rows pgx.Rows) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog

	for rows.Next() {
		log := &domain.AuditLog{}
		var oldValuesJSON, newValuesJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.AdminID,
			&log.AdminName,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&oldValuesJSON,
			&newValuesJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if oldValuesJSON != nil {
			if err := json.Unmarshal(oldValuesJSON, &log.OldValues); err != nil {
				return nil, fmt.Errorf("failed to unmarshal old values: %w", err)
			}
		}

		if newValuesJSON != nil {
			if err := json.Unmarshal(newValuesJSON, &log.NewValues); err != nil {
				return nil, fmt.Errorf("failed to unmarshal new values: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}
