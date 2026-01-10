// Package main is the entry point for the Full Stack Go Template application.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/noruj-official/full-stack-go-template/internal/config"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/handler"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/repository/postgres"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	"github.com/noruj-official/full-stack-go-template/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	db, err := postgres.New(ctx, cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Run migrations
	if err := db.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	passwordResetRepo := postgres.NewPasswordResetRepository(db)
	activityRepo := postgres.NewActivityLogRepository(db)
	auditRepo := postgres.NewAuditLogRepository(db)
	featureRepo := postgres.NewFeatureRepository(db)
	oauthRepo := postgres.NewOAuthRepository(db, cfg.Auth.Secret)

	// Initialize services
	emailService := service.NewResendEmailService(cfg.Email.ResendAPIKey, cfg.Email.ResendFromEmail, cfg.App.URL)
	userService := service.NewUserService(userRepo)
	featureService := service.NewFeatureService(featureRepo)
	authService := service.NewAuthService(userRepo, sessionRepo, passwordResetRepo, oauthRepo, emailService, featureService, cfg.App.URL, cfg.Auth.Secret)
	activityService := service.NewActivityService(activityRepo)
	auditService := service.NewAuditService(auditRepo)

	// SyncFeatures feature flags
	err = featureService.SyncFeatures(context.Background(), map[string]domain.FeatureConfig{
		domain.FeatureThemeManagement: {
			Description:    "Enables theme switching logic",
			DefaultEnabled: true,
		},
		domain.FeatureEmailAuth: {
			Description:    "Enables email-based sign-in and sign-up (Magic Link)",
			DefaultEnabled: true,
		},
		domain.FeatureEmailPasswordAuth: {
			Description:    "Enables standard email and password authentication",
			DefaultEnabled: true,
		},
		domain.FeatureEmailVerification: {
			Description:    "Enables email verification for new signups",
			DefaultEnabled: true,
		},
		domain.FeatureOAuth: {
			Description:    "Enables OAuth authentication with third-party providers",
			DefaultEnabled: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to sync feature flags: %w", err)
	}

	// Initialize storage service
	storageService, err := storage.NewService(cfg, db.Pool)
	if err != nil {
		return fmt.Errorf("failed to initialize storage service: %w", err)
	}

	// Initialize handlers
	baseHandler := handler.NewHandler(cfg.App.Name, cfg.App.Logo, featureService)

	homeHandler := handler.NewHomeHandler(baseHandler, db)
	userHandler := handler.NewUserHandler(baseHandler, userService, auditService)
	authHandler := handler.NewAuthHandler(baseHandler, authService, userService, activityService)
	activityHandler := handler.NewActivityHandler(baseHandler, activityService)
	profileHandler := handler.NewProfileHandler(baseHandler, userService, activityService, storageService)
	settingsHandler := handler.NewSettingsHandler(baseHandler, userService, activityService)
	analyticsHandler := handler.NewAnalyticsHandler(baseHandler, db)
	auditHandler := handler.NewAuditHandler(baseHandler, auditService, db)
	featureHandler := handler.NewFeatureHandler(baseHandler, featureService, auditService)
	adminOAuthHandler := handler.NewAdminOAuthHandler(baseHandler, oauthRepo, auditService, cfg.App.URL)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuth(authService)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	fileServer := http.FileServer(http.Dir("web/assets"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", fileServer))

	// Public routes (no auth required)
	mux.HandleFunc("GET /{$}", homeHandler.Index)
	mux.HandleFunc("GET /health", homeHandler.HealthCheck)

	// Rate limiter for auth routes (5 reqs/10s roughly, burst 5)
	authLimiter := middleware.RateLimitMiddleware(0.5, 5)

	// Auth routes
	mux.Handle("GET /signin", authLimiter(http.HandlerFunc(authHandler.SignInPage)))
	mux.Handle("POST /signin", authLimiter(http.HandlerFunc(authHandler.SignIn)))
	mux.Handle("GET /signup", authLimiter(http.HandlerFunc(authHandler.SignupPage)))
	mux.Handle("POST /signup", authLimiter(http.HandlerFunc(authHandler.Signup)))
	mux.HandleFunc("POST /logout", authHandler.Logout)
	mux.HandleFunc("GET /verify-email", authHandler.VerifyEmailPage)
	mux.Handle("GET /forgot-password", authLimiter(http.HandlerFunc(authHandler.ForgotPasswordPage)))
	mux.Handle("POST /forgot-password", authLimiter(http.HandlerFunc(authHandler.ForgotPassword)))
	mux.Handle("GET /reset-password", authLimiter(http.HandlerFunc(authHandler.ResetPasswordPage)))
	mux.Handle("POST /reset-password", authLimiter(http.HandlerFunc(authHandler.ResetPassword)))

	// Email Auth routes
	mux.Handle("POST /auth/email/request", authLimiter(http.HandlerFunc(authHandler.HandleEmailAuthRequest)))
	mux.Handle("GET /auth/email/verify", authLimiter(http.HandlerFunc(authHandler.HandleEmailAuthVerify)))
	mux.Handle("GET /auth/complete-profile", middleware.RequireAuth(http.HandlerFunc(authHandler.CompleteProfilePage)))
	mux.Handle("POST /auth/complete-profile", middleware.RequireAuth(http.HandlerFunc(authHandler.CompleteProfile)))

	// OAuth routes
	mux.Handle("GET /auth/{provider}", authLimiter(http.HandlerFunc(authHandler.HandleOAuthLogin)))
	mux.Handle("GET /auth/{provider}/callback", authLimiter(http.HandlerFunc(authHandler.HandleOAuthCallback)))

	// Backwards compatible redirect from /login to /signin
	mux.HandleFunc("GET /login", authHandler.LoginRedirect)

	// Role-based dashboard routes
	mux.Handle("GET /u/dashboard", middleware.RequireAuth(http.HandlerFunc(homeHandler.UserDashboard)))
	mux.Handle("GET /a/dashboard", middleware.RequireRole(domain.RoleAdmin, domain.RoleSuperAdmin)(http.HandlerFunc(homeHandler.AdminDashboard)))
	mux.Handle("GET /s/dashboard", middleware.RequireRole(domain.RoleSuperAdmin)(http.HandlerFunc(homeHandler.SuperAdminDashboard)))

	// Backwards compatible redirect from /dashboard to role-appropriate dashboard
	mux.Handle("GET /dashboard", middleware.RequireAuth(http.HandlerFunc(homeHandler.DashboardRedirect)))

	// User routes (authenticated users)
	userOnly := middleware.RequireAuth
	mux.Handle("GET /components/sidebar", userOnly(http.HandlerFunc(homeHandler.Sidebar)))
	mux.Handle("GET /u/activity", userOnly(http.HandlerFunc(activityHandler.UserActivity)))
	mux.Handle("GET /u/profile", userOnly(http.HandlerFunc(profileHandler.ProfilePage)))
	mux.Handle("POST /u/profile", userOnly(http.HandlerFunc(profileHandler.UpdateProfile)))
	mux.Handle("POST /u/profile/image", userOnly(http.HandlerFunc(profileHandler.UploadProfileImage)))
	mux.Handle("GET /u/profile/image", userOnly(http.HandlerFunc(profileHandler.GetMyProfileImage)))
	mux.Handle("GET /u/settings", userOnly(http.HandlerFunc(settingsHandler.Settings)))
	mux.Handle("POST /u/settings", userOnly(http.HandlerFunc(settingsHandler.Settings)))
	mux.Handle("POST /u/settings/password", userOnly(http.HandlerFunc(settingsHandler.UpdatePassword)))
	mux.Handle("POST /u/signout-all", userOnly(http.HandlerFunc(authHandler.SignOutAllDevices)))

	// API routes for retrieving user profile images (accessible to authenticated users)
	mux.Handle("GET /api/users/{id}/image", userOnly(http.HandlerFunc(profileHandler.GetUserProfileImage)))

	// Admin routes (require admin role)
	adminOnly := middleware.RequireRole(domain.RoleAdmin, domain.RoleSuperAdmin)
	mux.Handle("GET /a/users", adminOnly(http.HandlerFunc(userHandler.List)))
	mux.Handle("GET /a/users/create", adminOnly(http.HandlerFunc(userHandler.Create)))
	mux.Handle("POST /a/users/create", adminOnly(http.HandlerFunc(userHandler.Create)))
	mux.Handle("GET /a/users/{id}/edit", adminOnly(http.HandlerFunc(userHandler.Edit)))
	mux.Handle("POST /a/users/{id}/edit", adminOnly(http.HandlerFunc(userHandler.Edit)))
	mux.Handle("POST /a/users/{id}/status", adminOnly(http.HandlerFunc(userHandler.UpdateStatus)))
	mux.Handle("DELETE /a/users/{id}", middleware.RequireRole(domain.RoleSuperAdmin)(http.HandlerFunc(userHandler.Delete)))

	// Feature Flags Admin
	mux.Handle("GET /a/features", adminOnly(http.HandlerFunc(featureHandler.List)))
	mux.Handle("POST /a/features/toggle", adminOnly(http.HandlerFunc(featureHandler.Toggle)))

	// OAuth Settings
	mux.Handle("GET /a/oauth", adminOnly(http.HandlerFunc(adminOAuthHandler.List)))
	mux.Handle("POST /a/oauth/{provider}", adminOnly(http.HandlerFunc(adminOAuthHandler.Update)))

	// Admin analytics and activity routes
	mux.Handle("GET /a/analytics", adminOnly(http.HandlerFunc(analyticsHandler.AdminAnalytics)))
	mux.Handle("GET /a/activity", adminOnly(http.HandlerFunc(analyticsHandler.SystemActivity)))

	// Super Admin routes (require super admin role)
	superAdminOnly := middleware.RequireRole(domain.RoleSuperAdmin)
	mux.Handle("GET /s/audit", superAdminOnly(http.HandlerFunc(auditHandler.AuditLogs)))
	mux.Handle("GET /s/system", superAdminOnly(http.HandlerFunc(auditHandler.SystemHealth)))

	// Catch-all for 404s (must be added last if using patterns that might overlap, but "/" is most general)
	mux.HandleFunc("/", homeHandler.NotFound)

	// Apply middleware stack
	var h http.Handler = mux
	h = authMiddleware.Handler(h) // Auth middleware (loads user into context)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.CORS(h)

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on http://%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server stopped")
	return nil
}
