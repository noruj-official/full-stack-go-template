// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
	"github.com/shaik-noor/full-stack-go-template/web/templ/pages"
)

// ActivityHandler handles activity-related HTTP requests.
type ActivityHandler struct {
	*Handler
	activityService service.ActivityService
}

// NewActivityHandler creates a new activity handler.
func NewActivityHandler(base *Handler, activityService service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		Handler:         base,
		activityService: activityService,
	}
}

// UserActivity renders the user's activity log page.
func (h *ActivityHandler) UserActivity(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get user activities (limit to last 50)
	activities, err := h.activityService.GetUserActivities(r.Context(), user.ID, 50)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load activity logs")
		return
	}

	// Format activities for display
	formattedActivities := make([]pages.ActivityViewModel, 0, len(activities))
	for _, activity := range activities {
		var ipAddress string
		if activity.IPAddress != nil {
			ipAddress = *activity.IPAddress
		}
		formattedActivities = append(formattedActivities, pages.ActivityViewModel{
			Type:        string(activity.ActivityType),
			Description: activity.Description,
			IPAddress:   ipAddress,
			TimeAgo:     formatTimeAgo(activity.CreatedAt),
			FullTime:    activity.CreatedAt.Format("Jan 02, 2006 at 3:04 PM"),
		})
	}

	h.RenderTempl(w, r, pages.UserActivity("Activity Log", formattedActivities, user, h.GetTheme(r)))
}

// formatTimeAgo formats a time as a relative time string.
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return formatInt(minutes) + " minutes ago"
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return formatInt(hours) + " hours ago"
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 30 {
		return formatInt(days) + " days ago"
	}

	// For older dates, show the actual date
	return t.Format("Jan 02, 2006")
}

func formatInt(n int) string {
	// Use built-in conversion for correctness across all ranges
	return strconv.Itoa(n)
}
