package migrations

import (
	"database/sql"
	"fmt"
)

// Migration represents a single database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// AllMigrations contains all database migrations in order
var AllMigrations = []Migration{
	{
		Version: 1,
		Name:    "Add profile_name columns and indices",
		Up: `
			-- Create indices for filtering by profile (columns already exist in schema)
			CREATE INDEX IF NOT EXISTS idx_analytics_profile ON analytics(profile_name);
			CREATE INDEX IF NOT EXISTS idx_history_profile ON history(profile_name);
			CREATE INDEX IF NOT EXISTS idx_stress_configs_profile ON stress_test_configs(profile_name);
			CREATE INDEX IF NOT EXISTS idx_stress_runs_profile ON stress_test_runs(profile_name);
		`,
		Down: `
			-- Drop indices
			DROP INDEX IF EXISTS idx_analytics_profile;
			DROP INDEX IF EXISTS idx_history_profile;
			DROP INDEX IF EXISTS idx_stress_configs_profile;
			DROP INDEX IF EXISTS idx_stress_runs_profile;
		`,
	},
	{
		Version: 2,
		Name:    "Clean up legacy data without profile_name",
		Up: `
			-- Delete all entries where profile_name is NULL (legacy data from before profiles)
			DELETE FROM analytics WHERE profile_name IS NULL;
			DELETE FROM history WHERE profile_name IS NULL;
			DELETE FROM stress_test_configs WHERE profile_name IS NULL;
			DELETE FROM stress_test_runs WHERE profile_name IS NULL;
		`,
		Down: `
			-- Cannot restore deleted data
		`,
	},
	{
		Version: 3,
		Name:    "Add validation_error column to stress_test_metrics",
		Up: `
			-- validation_error column already exists in current schema
			-- This migration is kept for backward compatibility with older databases
		`,
		Down: `
			-- SQLite does not support DROP COLUMN easily
			-- Leaving column in place for backward compatibility
		`,
	},
	{
		Version: 4,
		Name:    "Add total_validation_errors column to stress_test_runs",
		Up: `
			-- total_validation_errors column already exists in current schema
			-- This migration is kept for backward compatibility with older databases
		`,
		Down: `
			-- SQLite does not support DROP COLUMN easily
			-- Leaving column in place for backward compatibility
		`,
	},
}

// Run executes all pending migrations on the database
func Run(db *sql.DB) error {
	// Create migrations tracking table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// Apply pending migrations
	for _, migration := range AllMigrations {
		if migration.Version <= currentVersion {
			continue
		}

		// Execute migration
		_, err := db.Exec(migration.Up)
		if err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		// Record migration
		_, err = db.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version,
			migration.Name,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// GetCurrentVersion returns the current database schema version
func GetCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow(`
		SELECT COALESCE(MAX(version), 0)
		FROM schema_migrations
	`).Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return version, nil
}
