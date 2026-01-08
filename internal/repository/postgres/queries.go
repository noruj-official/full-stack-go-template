// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"embed"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// GetMigrations returns all migration SQL statements from the embedded schema file.
func GetMigrations() ([]string, error) {
	content, err := migrationsFS.ReadFile("migrations/schema.sql")
	if err != nil {
		return nil, err
	}

	// Split by semicolon to get individual statements
	statements := strings.Split(string(content), ";")

	var migrations []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// Skip pure comment blocks
		lines := strings.Split(stmt, "\n")
		hasSQL := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				hasSQL = true
				break
			}
		}
		if hasSQL {
			migrations = append(migrations, stmt)
		}
	}

	return migrations, nil
}
