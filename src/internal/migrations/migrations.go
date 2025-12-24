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
	{
		Version: 5,
		Name:    "Add composite indexes for analytics query optimization",
		Up: `
			-- Composite index for profile filtering + timestamp ordering (GetStatsPerFile ORDER BY)
			CREATE INDEX IF NOT EXISTS idx_analytics_profile_timestamp ON analytics(profile_name, timestamp DESC);

			-- Composite index for GROUP BY operations (file_path, normalized_path, method)
			CREATE INDEX IF NOT EXISTS idx_analytics_grouping ON analytics(file_path, normalized_path, method);

			-- Covering index for the main query (includes commonly accessed columns)
			-- This helps with the WHERE + GROUP BY + aggregate functions
			CREATE INDEX IF NOT EXISTS idx_analytics_profile_grouping ON analytics(profile_name, file_path, normalized_path, method, status_code, duration_ms, timestamp);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_analytics_profile_timestamp;
			DROP INDEX IF EXISTS idx_analytics_grouping;
			DROP INDEX IF EXISTS idx_analytics_profile_grouping;
		`,
	},
}

// InitSchema creates all tables required across all modules
// This must be called before running migrations to ensure all tables exist
func InitSchema(db *sql.DB) error {
	schema := `
	-- Analytics table
	CREATE TABLE IF NOT EXISTS analytics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		normalized_path TEXT NOT NULL,
		method TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		request_size INTEGER NOT NULL DEFAULT 0,
		response_size INTEGER NOT NULL DEFAULT 0,
		duration_ms INTEGER NOT NULL,
		error_message TEXT,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		profile_name TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_file_path ON analytics(file_path);
	CREATE INDEX IF NOT EXISTS idx_normalized_path ON analytics(normalized_path);
	CREATE INDEX IF NOT EXISTS idx_method ON analytics(method);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON analytics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_status_code ON analytics(status_code);

	-- History table
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		request_file TEXT NOT NULL,
		request_name TEXT,
		method TEXT NOT NULL,
		url TEXT NOT NULL,
		headers TEXT NOT NULL,
		body TEXT,
		response_status INTEGER NOT NULL,
		response_status_text TEXT NOT NULL,
		response_headers TEXT NOT NULL,
		response_body TEXT NOT NULL,
		duration_ms INTEGER NOT NULL,
		request_size INTEGER,
		response_size INTEGER,
		error TEXT,
		profile_name TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_history_request_file ON history(request_file);
	CREATE INDEX IF NOT EXISTS idx_history_method ON history(method);
	CREATE INDEX IF NOT EXISTS idx_history_url ON history(url);

	-- Stress test tables
	CREATE TABLE IF NOT EXISTS stress_test_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		request_file TEXT NOT NULL,
		profile_name TEXT,
		concurrent_connections INTEGER NOT NULL DEFAULT 10,
		total_requests INTEGER NOT NULL DEFAULT 100,
		ramp_up_duration_sec INTEGER DEFAULT 0,
		test_duration_sec INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS stress_test_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_id INTEGER,
		config_name TEXT NOT NULL,
		request_file TEXT NOT NULL,
		profile_name TEXT,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		status TEXT NOT NULL,
		total_requests_sent INTEGER DEFAULT 0,
		total_requests_completed INTEGER DEFAULT 0,
		total_errors INTEGER DEFAULT 0,
		total_validation_errors INTEGER DEFAULT 0,
		avg_duration_ms REAL DEFAULT 0,
		min_duration_ms INTEGER DEFAULT 0,
		max_duration_ms INTEGER DEFAULT 0,
		p50_duration_ms INTEGER DEFAULT 0,
		p95_duration_ms INTEGER DEFAULT 0,
		p99_duration_ms INTEGER DEFAULT 0,
		FOREIGN KEY (config_id) REFERENCES stress_test_configs(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_stress_runs_started_at ON stress_test_runs(started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_stress_runs_config_id ON stress_test_runs(config_id);
	CREATE INDEX IF NOT EXISTS idx_stress_runs_status ON stress_test_runs(status);

	CREATE TABLE IF NOT EXISTS stress_test_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		elapsed_ms INTEGER NOT NULL,
		status_code INTEGER NOT NULL,
		duration_ms INTEGER NOT NULL,
		request_size INTEGER DEFAULT 0,
		response_size INTEGER DEFAULT 0,
		error_message TEXT,
		validation_error TEXT,
		FOREIGN KEY (run_id) REFERENCES stress_test_runs(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_stress_metrics_run_id ON stress_test_metrics(run_id);
	CREATE INDEX IF NOT EXISTS idx_stress_metrics_timestamp ON stress_test_metrics(run_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_stress_metrics_elapsed ON stress_test_metrics(run_id, elapsed_ms);

	-- JSONPath bookmarks table
	CREATE TABLE IF NOT EXISTS jsonpath_bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		expression TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_jsonpath_bookmarks_created_at ON jsonpath_bookmarks(created_at DESC);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// Run executes all pending migrations on the database
func Run(db *sql.DB) error {
	// Initialize schema first to ensure all tables exist
	if err := InitSchema(db); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

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
