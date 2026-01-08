// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the PostgreSQL connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool.
func New(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// Health checks if the database connection is healthy.
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// RunMigrations executes database migrations from embedded SQL files.
// In a production app, consider using a proper migration tool like golang-migrate.
func (db *DB) RunMigrations(ctx context.Context) error {
	migrations, err := GetMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, migration := range migrations {
		if _, err := db.Pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}
