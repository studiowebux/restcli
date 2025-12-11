package stresstest

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/studiowebux/restcli/internal/migrations"
)

// Manager handles stress test data persistence
type Manager struct {
	db *sql.DB
}

// NewManager creates a new stress test manager
func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	m := &Manager{db: db}

	// Run database migrations (includes schema initialization)
	if err := migrations.Run(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return m, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// SaveConfig saves or updates a stress test configuration
func (m *Manager) SaveConfig(config *Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	if config.ID == 0 {
		// Insert new config
		result, err := m.db.Exec(`
			INSERT INTO stress_test_configs
			(name, request_file, profile_name, concurrent_connections, total_requests, ramp_up_duration_sec, test_duration_sec)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, config.Name, config.RequestFile, config.ProfileName, config.ConcurrentConns, config.TotalRequests, config.RampUpDurationSec, config.TestDurationSec)
		if err != nil {
			return fmt.Errorf("failed to insert config: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		config.ID = id
	} else {
		// Update existing config
		_, err := m.db.Exec(`
			UPDATE stress_test_configs
			SET name = ?, request_file = ?, profile_name = ?, concurrent_connections = ?,
			    total_requests = ?, ramp_up_duration_sec = ?, test_duration_sec = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, config.Name, config.RequestFile, config.ProfileName, config.ConcurrentConns, config.TotalRequests, config.RampUpDurationSec, config.TestDurationSec, config.ID)
		if err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}
	}
	return nil
}

// GetConfig retrieves a config by ID
func (m *Manager) GetConfig(id int64) (*Config, error) {
	config := &Config{}
	err := m.db.QueryRow(`
		SELECT id, name, request_file, COALESCE(profile_name, ''), concurrent_connections, total_requests,
		       ramp_up_duration_sec, test_duration_sec, created_at, updated_at
		FROM stress_test_configs WHERE id = ?
	`, id).Scan(&config.ID, &config.Name, &config.RequestFile, &config.ProfileName,
		&config.ConcurrentConns, &config.TotalRequests, &config.RampUpDurationSec,
		&config.TestDurationSec, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetConfigByName retrieves a config by name and profile
func (m *Manager) GetConfigByName(name string, profileName string) (*Config, error) {
	config := &Config{}
	err := m.db.QueryRow(`
		SELECT id, name, request_file, COALESCE(profile_name, ''), concurrent_connections, total_requests,
		       ramp_up_duration_sec, test_duration_sec, created_at, updated_at
		FROM stress_test_configs WHERE name = ? AND (profile_name = ? OR profile_name IS NULL)
	`, name, profileName).Scan(&config.ID, &config.Name, &config.RequestFile, &config.ProfileName,
		&config.ConcurrentConns, &config.TotalRequests, &config.RampUpDurationSec,
		&config.TestDurationSec, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ListConfigs returns all saved configurations for the specified profile
func (m *Manager) ListConfigs(profileName string) ([]*Config, error) {
	rows, err := m.db.Query(`
		SELECT id, name, request_file, COALESCE(profile_name, ''), concurrent_connections, total_requests,
		       ramp_up_duration_sec, test_duration_sec, created_at, updated_at
		FROM stress_test_configs
		WHERE profile_name = ? OR (profile_name IS NULL AND ? = '')
		ORDER BY updated_at DESC
	`, profileName, profileName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*Config
	for rows.Next() {
		config := &Config{}
		err := rows.Scan(&config.ID, &config.Name, &config.RequestFile, &config.ProfileName,
			&config.ConcurrentConns, &config.TotalRequests, &config.RampUpDurationSec,
			&config.TestDurationSec, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// DeleteConfig deletes a configuration
func (m *Manager) DeleteConfig(id int64) error {
	_, err := m.db.Exec("DELETE FROM stress_test_configs WHERE id = ?", id)
	return err
}

// CreateRun creates a new stress test run record
func (m *Manager) CreateRun(run *Run) error {
	result, err := m.db.Exec(`
		INSERT INTO stress_test_runs
		(config_id, config_name, request_file, profile_name, started_at, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, run.ConfigID, run.ConfigName, run.RequestFile, run.ProfileName, run.StartedAt, run.Status)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	run.ID = id
	return nil
}

// UpdateRun updates a stress test run record
func (m *Manager) UpdateRun(run *Run) error {
	_, err := m.db.Exec(`
		UPDATE stress_test_runs
		SET completed_at = ?, status = ?, total_requests_sent = ?, total_requests_completed = ?,
		    total_errors = ?, total_validation_errors = ?, avg_duration_ms = ?, min_duration_ms = ?, max_duration_ms = ?,
		    p50_duration_ms = ?, p95_duration_ms = ?, p99_duration_ms = ?
		WHERE id = ?
	`, run.CompletedAt, run.Status, run.TotalRequestsSent, run.TotalRequestsCompleted,
		run.TotalErrors, run.TotalValidationErrors, run.AvgDurationMs, run.MinDurationMs, run.MaxDurationMs,
		run.P50DurationMs, run.P95DurationMs, run.P99DurationMs, run.ID)
	return err
}

// GetRun retrieves a run by ID
func (m *Manager) GetRun(id int64) (*Run, error) {
	run := &Run{}
	var configID sql.NullInt64
	var completedAt sql.NullTime

	err := m.db.QueryRow(`
		SELECT id, config_id, config_name, request_file, COALESCE(profile_name, ''), started_at, completed_at, status,
		       total_requests_sent, total_requests_completed, total_errors, COALESCE(total_validation_errors, 0),
		       COALESCE(avg_duration_ms, 0), COALESCE(min_duration_ms, 0), COALESCE(max_duration_ms, 0),
		       COALESCE(p50_duration_ms, 0), COALESCE(p95_duration_ms, 0), COALESCE(p99_duration_ms, 0)
		FROM stress_test_runs WHERE id = ?
	`, id).Scan(&run.ID, &configID, &run.ConfigName, &run.RequestFile, &run.ProfileName,
		&run.StartedAt, &completedAt, &run.Status, &run.TotalRequestsSent,
		&run.TotalRequestsCompleted, &run.TotalErrors, &run.TotalValidationErrors, &run.AvgDurationMs, &run.MinDurationMs,
		&run.MaxDurationMs, &run.P50DurationMs, &run.P95DurationMs, &run.P99DurationMs)
	if err != nil {
		return nil, err
	}

	if configID.Valid {
		run.ConfigID = &configID.Int64
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}

	return run, nil
}

// ListRuns returns all stress test runs for the specified profile
func (m *Manager) ListRuns(profileName string, limit int) ([]*Run, error) {
	query := `
		SELECT id, config_id, config_name, request_file, COALESCE(profile_name, ''), started_at, completed_at, status,
		       total_requests_sent, total_requests_completed, total_errors, COALESCE(total_validation_errors, 0),
		       COALESCE(avg_duration_ms, 0), COALESCE(min_duration_ms, 0), COALESCE(max_duration_ms, 0),
		       COALESCE(p50_duration_ms, 0), COALESCE(p95_duration_ms, 0), COALESCE(p99_duration_ms, 0)
		FROM stress_test_runs
		WHERE profile_name = ? OR (profile_name IS NULL AND ? = '')
		ORDER BY started_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := m.db.Query(query, profileName, profileName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*Run
	for rows.Next() {
		run := &Run{}
		var configID sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(&run.ID, &configID, &run.ConfigName, &run.RequestFile, &run.ProfileName,
			&run.StartedAt, &completedAt, &run.Status, &run.TotalRequestsSent,
			&run.TotalRequestsCompleted, &run.TotalErrors, &run.TotalValidationErrors, &run.AvgDurationMs, &run.MinDurationMs,
			&run.MaxDurationMs, &run.P50DurationMs, &run.P95DurationMs, &run.P99DurationMs)
		if err != nil {
			return nil, err
		}

		if configID.Valid {
			run.ConfigID = &configID.Int64
		}
		if completedAt.Valid {
			run.CompletedAt = &completedAt.Time
		}

		runs = append(runs, run)
	}
	return runs, nil
}

// DeleteRun deletes a stress test run and all its metrics
func (m *Manager) DeleteRun(id int64) error {
	_, err := m.db.Exec("DELETE FROM stress_test_runs WHERE id = ?", id)
	return err
}

// SaveMetric saves a single request metric
func (m *Manager) SaveMetric(metric *Metric) error {
	result, err := m.db.Exec(`
		INSERT INTO stress_test_metrics
		(run_id, timestamp, elapsed_ms, status_code, duration_ms, request_size, response_size, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, metric.RunID, metric.Timestamp, metric.ElapsedMs, metric.StatusCode, metric.DurationMs,
		metric.RequestSize, metric.ResponseSize, metric.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to save metric: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	metric.ID = id
	return nil
}

// SaveMetricsBatch saves multiple metrics in a single transaction
func (m *Manager) SaveMetricsBatch(metrics []*Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO stress_test_metrics
		(run_id, timestamp, elapsed_ms, status_code, duration_ms, request_size, response_size, error_message, validation_error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, metric := range metrics {
		_, err := stmt.Exec(metric.RunID, metric.Timestamp, metric.ElapsedMs, metric.StatusCode,
			metric.DurationMs, metric.RequestSize, metric.ResponseSize, metric.ErrorMessage, metric.ValidationError)
		if err != nil {
			return fmt.Errorf("failed to insert metric: %w", err)
		}
	}

	return tx.Commit()
}

// GetMetrics retrieves all metrics for a run
func (m *Manager) GetMetrics(runID int64) ([]*Metric, error) {
	rows, err := m.db.Query(`
		SELECT id, run_id, timestamp, elapsed_ms, status_code, duration_ms, request_size, response_size,
		       error_message, COALESCE(validation_error, '')
		FROM stress_test_metrics
		WHERE run_id = ?
		ORDER BY elapsed_ms
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		metric := &Metric{}
		var errorMsg sql.NullString
		var validationErr sql.NullString

		err := rows.Scan(&metric.ID, &metric.RunID, &metric.Timestamp, &metric.ElapsedMs,
			&metric.StatusCode, &metric.DurationMs, &metric.RequestSize, &metric.ResponseSize,
			&errorMsg, &validationErr)
		if err != nil {
			return nil, err
		}

		if errorMsg.Valid {
			metric.ErrorMessage = errorMsg.String
		}
		if validationErr.Valid {
			metric.ValidationError = validationErr.String
		}

		metrics = append(metrics, metric)
	}
	return metrics, nil
}
