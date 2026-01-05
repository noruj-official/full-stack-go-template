// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/repository/postgres"
)

// AnalyticsHandler handles analytics-related HTTP requests.
type AnalyticsHandler struct {
	*Handler
	db *postgres.DB
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(base *Handler, db *postgres.DB) *AnalyticsHandler {
	return &AnalyticsHandler{
		Handler: base,
		db:      db,
	}
}

// AdminAnalytics renders the admin analytics dashboard.
func (h *AnalyticsHandler) AdminAnalytics(w http.ResponseWriter, r *http.Request) {
	userRepo := postgres.NewUserRepository(h.db)

	// Get user statistics
	totalUsers, err := userRepo.Count(r.Context())
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load statistics")
		return
	}

	// Get all users for role breakdown
	allUsers, err := userRepo.List(r.Context(), 10000, 0)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load user data")
		return
	}

	// Calculate role statistics
	roleStats := map[string]int{
		"user":        0,
		"admin":       0,
		"super_admin": 0,
	}

	for _, user := range allUsers {
		roleStats[string(user.Role)]++
	}

	// Get recent users (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentCount := 0
	for _, user := range allUsers {
		if user.CreatedAt.After(thirtyDaysAgo) {
			recentCount++
		}
	}

	// Get user growth data (last 7 days)
	growthData := make([]map[string]interface{}, 7)
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		count := 0
		for _, user := range allUsers {
			if user.CreatedAt.After(startOfDay) && user.CreatedAt.Before(endOfDay) {
				count++
			}
		}

		growthData[6-i] = map[string]interface{}{
			"date":  startOfDay.Format("Mon"),
			"count": count,
		}
	}

	data := map[string]any{
		"Title":         "Analytics",
		"TotalUsers":    totalUsers,
		"RecentUsers":   recentCount,
		"UserRoleStats": roleStats,
		"GrowthData":    growthData,
		"ShowSidebar":   true,
	}

	h.RenderWithUser(w, r, "admin_analytics.html", data)
}

// SystemActivity renders the system-wide activity feed.
func (h *AnalyticsHandler) SystemActivity(w http.ResponseWriter, r *http.Request) {
	// Get recent activities from all users
	query := `
		SELECT a.id, a.user_id, u.name as user_name, a.activity_type, a.description, 
		       a.ip_address, a.created_at
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC
		LIMIT 50
	`

	rows, err := h.db.Pool.Query(r.Context(), query)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load activity feed")
		return
	}
	defer rows.Close()

	var activities []map[string]interface{}
	for rows.Next() {
		var id, userID string
		var userName, activityType, description string
		var ipAddress *string
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &userName, &activityType, &description, &ipAddress, &createdAt); err != nil {
			continue
		}

		activities = append(activities, map[string]interface{}{
			"UserName":    userName,
			"Type":        activityType,
			"Description": description,
			"IPAddress":   ipAddress,
			"TimeAgo":     formatTimeAgo(createdAt),
		})
	}

	data := map[string]any{
		"Title":       "System Activity",
		"Activities":  activities,
		"ShowSidebar": true,
	}

	h.RenderWithUser(w, r, "system_activity.html", data)
}
