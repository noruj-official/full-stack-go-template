// Package main is the entry point for the Go Starter application.
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

	"github.com/go-starter/internal/config"
	"github.com/go-starter/internal/domain"
	"github.com/go-starter/internal/handler"
	"github.com/go-starter/internal/middleware"
	"github.com/go-starter/internal/repository/postgres"
	"github.com/go-starter/internal/service"
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

	// Initialize services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo)

	// Initialize handlers
	baseHandler := handler.NewHandler("web/templates")
	if err := baseHandler.LoadTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	homeHandler := handler.NewHomeHandler(baseHandler, db)
	userHandler := handler.NewUserHandler(baseHandler, userService)
	authHandler := handler.NewAuthHandler(baseHandler, authService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuth(authService)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	fileServer := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Public routes (no auth required)
	mux.HandleFunc("GET /", homeHandler.Index)
	mux.HandleFunc("GET /health", homeHandler.HealthCheck)

	// Auth routes
	mux.HandleFunc("GET /login", authHandler.LoginPage)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /signup", authHandler.SignupPage)
	mux.HandleFunc("POST /signup", authHandler.Signup)
	mux.HandleFunc("POST /logout", authHandler.Logout)

	// Protected routes (require authentication)
	mux.Handle("GET /dashboard", middleware.RequireAuth(http.HandlerFunc(homeHandler.Dashboard)))

	// Admin routes (require admin role)
	adminOnly := middleware.RequireRole(domain.RoleAdmin, domain.RoleSuperAdmin)
	mux.Handle("GET /users", adminOnly(http.HandlerFunc(userHandler.List)))
	mux.Handle("GET /users/create", adminOnly(http.HandlerFunc(userHandler.Create)))
	mux.Handle("POST /users/create", adminOnly(http.HandlerFunc(userHandler.Create)))
	mux.Handle("GET /users/{id}/edit", adminOnly(http.HandlerFunc(userHandler.Edit)))
	mux.Handle("POST /users/{id}/edit", adminOnly(http.HandlerFunc(userHandler.Edit)))
	mux.Handle("DELETE /users/{id}", middleware.RequireRole(domain.RoleSuperAdmin)(http.HandlerFunc(userHandler.Delete)))

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
