package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters for each IP address.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new rate limiter that allows events up to rate r and permits bursts of at most b tokens.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Start a background goroutine to cleanup old entries every 5 minutes
	go i.cleanupLoop()

	return i
}

// AddIP creates a new rate limiter for an IP if one doesn't exist, or returns the existing one.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// cleanupLoop periodically removes unused limiters to prevent memory leaks.
// Note: This is a simple implementation. For production, track last access time per IP.
func (i *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		// In a real implementation we would check last access time.
		// For now, we'll just clear the map if it gets too large to prevent leaks in this starter
		i.mu.Lock()
		if len(i.ips) > 10000 {
			i.ips = make(map[string]*rate.Limiter)
		}
		i.mu.Unlock()
	}
}

// RateLimitMiddleware creates a middleware that limits requests by IP address.
func RateLimitMiddleware(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := NewIPRateLimiter(rate.Limit(requestsPerSecond), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIPAddress(r)
			if ip == "" {
				ip = r.RemoteAddr
			}

			if !limiter.GetLimiter(ip).Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper to get IP (duplicate of logic in auth handler, maybe move to utils later)
func getIPAddress(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
