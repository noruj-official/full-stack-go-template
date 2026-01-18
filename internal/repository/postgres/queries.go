// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"embed"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// GetMigrations returns all migration SQL statements from the embedded schema files.
func GetMigrations() ([]string, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []string

	// Basic sort is provided by ReadDir (lexical), which works for 001, 002 prefix.
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, err
		}

		// Don't split by semicolon. pgx.Exec handles multiple statements (or we rely on it for DO blocks).
		// We return the whole file as one migration step.
		strContent := string(content)
		if strings.TrimSpace(strContent) != "" {
			migrations = append(migrations, strContent)
		}
	}

	return migrations, nil
}
