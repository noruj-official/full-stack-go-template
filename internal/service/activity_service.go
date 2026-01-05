// Package service provides business logic implementations.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/repository/postgres"
)

// ActivityService handles activity log operations.
type ActivityService interface {
	LogActivity(ctx context.Context, userID uuid.UUID, activityType domain.ActivityType, description string, ipAddress, userAgent *string) error
	GetUserActivities(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.ActivityLog, error)
}

type activityService struct {
	activityRepo *postgres.ActivityLogRepository
}

// NewActivityService creates a new activity service.
func NewActivityService(activityRepo *postgres.ActivityLogRepository) ActivityService {
	return &activityService{
		activityRepo: activityRepo,
	}
}

// LogActivity logs a user activity.
func (s *activityService) LogActivity(ctx context.Context, userID uuid.UUID, activityType domain.ActivityType, description string, ipAddress, userAgent *string) error {
	log := &domain.ActivityLog{
		UserID:       userID,
		ActivityType: activityType,
		Description:  description,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	if err := s.activityRepo.Create(ctx, log); err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// GetUserActivities retrieves user activities.
func (s *activityService) GetUserActivities(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.ActivityLog, error) {
	logs, err := s.activityRepo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activities: %w", err)
	}

	return logs, nil
}

// AuditService handles audit log operations.
type AuditService interface {
	LogAudit(ctx context.Context, adminID uuid.UUID, action domain.AuditAction, resourceType string, resourceID *uuid.UUID, oldValues, newValues map[string]interface{}, ipAddress *string) error
	GetAuditLogs(ctx context.Context, limit, offset int) ([]*domain.AuditLog, int, error)
	GetAdminAuditLogs(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.AuditLog, error)
}

type auditService struct {
	auditRepo *postgres.AuditLogRepository
}

// NewAuditService creates a new audit service.
func NewAuditService(auditRepo *postgres.AuditLogRepository) AuditService {
	return &auditService{
		auditRepo: auditRepo,
	}
}

// LogAudit logs an administrative action.
func (s *auditService) LogAudit(ctx context.Context, adminID uuid.UUID, action domain.AuditAction, resourceType string, resourceID *uuid.UUID, oldValues, newValues map[string]interface{}, ipAddress *string) error {
	log := &domain.AuditLog{
		AdminID:      adminID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValues:    oldValues,
		NewValues:    newValues,
		IPAddress:    ipAddress,
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		return fmt.Errorf("failed to log audit: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs with pagination.
func (s *auditService) GetAuditLogs(ctx context.Context, limit, offset int) ([]*domain.AuditLog, int, error) {
	logs, err := s.auditRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	count, err := s.auditRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return logs, count, nil
}

// GetAdminAuditLogs retrieves audit logs for a specific admin.
func (s *auditService) GetAdminAuditLogs(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	logs, err := s.auditRepo.ListByAdmin(ctx, adminID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin audit logs: %w", err)
	}

	return logs, nil
}
