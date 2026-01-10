// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/repository/postgres"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/admin"
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
	growthData := make([]admin.GrowthMetric, 7)
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

		growthData[6-i] = admin.GrowthMetric{
			Date:  startOfDay.Format("Mon"),
			Count: count,
		}
	}

	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	props := admin.AdminAnalyticsProps{
		User:          middleware.GetUserFromContext(r.Context()),
		TotalUsers:    totalUsers,
		RecentUsers:   recentCount,
		UserRoleStats: roleStats,
		GrowthData:    growthData,
		Theme:         theme,
		ThemeEnabled:  themeEnabled,
		OAuthEnabled:  oauthEnabled,
	}

	admin.AdminAnalytics(props).Render(r.Context(), w)
}

// SystemActivity renders the system-wide activity feed.
func (h *AnalyticsHandler) SystemActivity(w http.ResponseWriter, r *http.Request) {
	// Parse offset parameter for pagination
	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	const limit = 10 // TODO: Change back to 50 for production

	// Get activities from all users with pagination
	query := `
		SELECT a.id, a.user_id, u.name as user_name, a.activity_type, a.description, 
		       a.ip_address, a.created_at
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := h.db.Pool.Query(r.Context(), query, limit, offset)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load activity feed")
		return
	}
	defer rows.Close()

	var activities []admin.SystemActivityItem
	for rows.Next() {
		var id, userID string
		var userName, activityType, description string
		var ipAddress *string
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &userName, &activityType, &description, &ipAddress, &createdAt); err != nil {
			continue
		}

		activities = append(activities, admin.SystemActivityItem{
			UserName:    userName,
			Type:        activityType,
			Description: description,
			IPAddress:   ipAddress,
			TimeAgo:     formatTimeAgo(createdAt),
		})
	}

	// Check if there are more activities
	hasMore := len(activities) == limit

	theme, themeEnabled := h.GetTheme(r)

	// If this is a partial request (offset provided), render activity items and button
	if offset > 0 {
		// Render the new activity items
		err := admin.SystemActivityItems(activities, offset+len(activities), false).Render(r.Context(), w)
		if err != nil {
			h.Error(w, r, http.StatusInternalServerError, "Failed to render activities")
			return
		}
		// Render the load more button
		admin.LoadMoreButton(offset+len(activities), hasMore).Render(r.Context(), w)
		return
	}

	oauthEnabled := h.GetOAuthEnabled(r)
	// Otherwise, render the full page
	props := admin.SystemActivityProps{
		User:         middleware.GetUserFromContext(r.Context()),
		Activities:   activities,
		CurrentCount: len(activities),
		HasMore:      hasMore,
		Theme:        theme,
		ThemeEnabled: themeEnabled,
		OAuthEnabled: oauthEnabled,
	}

	admin.SystemActivity(props).Render(r.Context(), w)
}
