package analytics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/migrations"
)

type Entry struct {
	ID             int64
	FilePath       string
	NormalizedPath string
	Method         string
	StatusCode     int
	RequestSize    int64
	ResponseSize   int64
	DurationMs     int64
	ErrorMessage   string
	Timestamp      time.Time
	ProfileName    string
}

type Stats struct {
	FilePath       string
	NormalizedPath string
	Method         string
	TotalCalls     int
	SuccessCount   int
	ErrorCount     int
	NetworkErrors  int // DNS, connection timeout, etc (status code 0)
	AvgDurationMs  float64
	MinDurationMs  int64
	MaxDurationMs  int64
	TotalReqSize   int64
	TotalRespSize  int64
	StatusCodes    map[int]int
	LastCalled     time.Time
}

type Manager struct {
	db *sql.DB
}

func NewManager(dbPath string) (*Manager, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, config.DirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create analytics directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open analytics database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to analytics database: %w", err)
	}

	m := &Manager{db: db}
	if err := m.initSchema(); err != nil {
		return nil, err
	}

	// Run database migrations
	if err := migrations.Run(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return m, nil
}

func (m *Manager) initSchema() error {
	schema := `
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
	`

	_, err := m.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize analytics schema: %w", err)
	}

	return nil
}

func (m *Manager) Save(entry Entry) error {
	query := `
		INSERT INTO analytics (file_path, normalized_path, method, status_code, request_size, response_size, duration_ms, error_message, timestamp, profile_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Format timestamp for SQLite in local time (YYYY-MM-DD HH:MM:SS)
	timestampStr := entry.Timestamp.Local().Format("2006-01-02 15:04:05")

	_, err := m.db.Exec(query,
		entry.FilePath,
		entry.NormalizedPath,
		entry.Method,
		entry.StatusCode,
		entry.RequestSize,
		entry.ResponseSize,
		entry.DurationMs,
		entry.ErrorMessage,
		timestampStr,
		entry.ProfileName,
	)

	if err != nil {
		return fmt.Errorf("failed to save analytics entry: %w", err)
	}

	return nil
}

func (m *Manager) LoadForFile(filePath string, profileName string, limit int) ([]Entry, error) {
	query := `
		SELECT id, file_path, normalized_path, method, status_code, request_size, response_size, duration_ms, error_message, timestamp, COALESCE(profile_name, '')
		FROM analytics
		WHERE file_path = ? AND (profile_name = ? OR (profile_name IS NULL AND ? = ''))
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, filePath, profileName, profileName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load analytics for file: %w", err)
	}
	defer rows.Close()

	return m.scanEntries(rows)
}

func (m *Manager) LoadForNormalizedPath(normalizedPath string, profileName string, limit int) ([]Entry, error) {
	query := `
		SELECT id, file_path, normalized_path, method, status_code, request_size, response_size, duration_ms, error_message, timestamp, COALESCE(profile_name, '')
		FROM analytics
		WHERE normalized_path = ? AND (profile_name = ? OR (profile_name IS NULL AND ? = ''))
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, normalizedPath, profileName, profileName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load analytics for normalized path: %w", err)
	}
	defer rows.Close()

	return m.scanEntries(rows)
}

func (m *Manager) LoadAll(profileName string, limit int) ([]Entry, error) {
	query := `
		SELECT id, file_path, normalized_path, method, status_code, request_size, response_size, duration_ms, error_message, timestamp, COALESCE(profile_name, '')
		FROM analytics
		WHERE profile_name = ? OR (profile_name IS NULL AND ? = '')
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, profileName, profileName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load all analytics: %w", err)
	}
	defer rows.Close()

	return m.scanEntries(rows)
}

func (m *Manager) scanEntries(rows *sql.Rows) ([]Entry, error) {
	var entries []Entry

	for rows.Next() {
		var e Entry
		var timestamp string
		var errorMsg sql.NullString

		err := rows.Scan(
			&e.ID,
			&e.FilePath,
			&e.NormalizedPath,
			&e.Method,
			&e.StatusCode,
			&e.RequestSize,
			&e.ResponseSize,
			&e.DurationMs,
			&errorMsg,
			&timestamp,
			&e.ProfileName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analytics entry: %w", err)
		}

		if errorMsg.Valid {
			e.ErrorMessage = errorMsg.String
		}

		// Parse as local time (SQLite stores without timezone info)
		e.Timestamp, err = time.ParseInLocation("2006-01-02 15:04:05", timestamp, time.Local)
		if err != nil {
			// Try RFC3339 format as fallback
			e.Timestamp, err = time.Parse(time.RFC3339, timestamp)
			if err != nil {
				// If both fail, use current time to avoid zero value
				e.Timestamp = time.Now()
			}
		}

		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (m *Manager) GetStatsPerFile(profileName string) ([]Stats, error) {
	// Use a subquery with JSON aggregation to get status codes in a single query
	query := `
		WITH status_codes_agg AS (
			SELECT
				file_path,
				normalized_path,
				method,
				json_group_object(CAST(status_code AS TEXT), count) as status_codes_json
			FROM (
				SELECT
					file_path,
					normalized_path,
					method,
					status_code,
					COUNT(*) as count
				FROM analytics
				WHERE profile_name = ? OR (profile_name IS NULL AND ? = '')
				GROUP BY file_path, normalized_path, method, status_code
			)
			GROUP BY file_path, normalized_path, method
		)
		SELECT
			a.file_path,
			a.normalized_path,
			a.method,
			COUNT(*) as total_calls,
			SUM(CASE WHEN a.status_code >= 200 AND a.status_code < 300 THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN a.status_code >= 400 THEN 1 ELSE 0 END) as error_count,
			SUM(CASE WHEN a.status_code = 0 THEN 1 ELSE 0 END) as network_errors,
			AVG(a.duration_ms) as avg_duration,
			MIN(a.duration_ms) as min_duration,
			MAX(a.duration_ms) as max_duration,
			SUM(a.request_size) as total_req_size,
			SUM(a.response_size) as total_resp_size,
			MAX(a.timestamp) as last_called,
			COALESCE(s.status_codes_json, '{}') as status_codes_json
		FROM analytics a
		LEFT JOIN status_codes_agg s ON a.file_path = s.file_path AND a.normalized_path = s.normalized_path AND a.method = s.method
		WHERE a.profile_name = ? OR (a.profile_name IS NULL AND ? = '')
		GROUP BY a.file_path, a.normalized_path, a.method
		ORDER BY last_called DESC
	`

	rows, err := m.db.Query(query, profileName, profileName, profileName, profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats per file: %w", err)
	}
	defer rows.Close()

	var statsList []Stats
	for rows.Next() {
		var s Stats
		var lastCalled sql.NullString
		var statusCodesJSON string

		err := rows.Scan(
			&s.FilePath,
			&s.NormalizedPath,
			&s.Method,
			&s.TotalCalls,
			&s.SuccessCount,
			&s.ErrorCount,
			&s.NetworkErrors,
			&s.AvgDurationMs,
			&s.MinDurationMs,
			&s.MaxDurationMs,
			&s.TotalReqSize,
			&s.TotalRespSize,
			&lastCalled,
			&statusCodesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		// Handle NULL timestamp (shouldn't happen with MAX, but be safe)
		if lastCalled.Valid && lastCalled.String != "" {
			// Parse as local time (SQLite stores without timezone info)
			s.LastCalled, err = time.ParseInLocation("2006-01-02 15:04:05", lastCalled.String, time.Local)
			if err != nil {
				// Try RFC3339 format as fallback
				s.LastCalled, err = time.Parse(time.RFC3339, lastCalled.String)
				if err != nil {
					// If both fail, use current time to avoid zero value
					s.LastCalled = time.Now()
				}
			}
		} else {
			// No timestamp available, use current time
			s.LastCalled = time.Now()
		}

		// Parse status codes from JSON
		s.StatusCodes = make(map[int]int)
		if statusCodesJSON != "{}" {
			var statusCodesMap map[string]int
			if err := json.Unmarshal([]byte(statusCodesJSON), &statusCodesMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal status codes: %w", err)
			}
			// Convert string keys to int keys
			for codeStr, count := range statusCodesMap {
				var code int
				if _, err := fmt.Sscanf(codeStr, "%d", &code); err == nil {
					s.StatusCodes[code] = count
				}
			}
		}

		statsList = append(statsList, s)
	}

	return statsList, rows.Err()
}

func (m *Manager) GetStatsPerNormalizedPath(profileName string) ([]Stats, error) {
	// Use a subquery with JSON aggregation to get status codes in a single query
	query := `
		WITH status_codes_agg AS (
			SELECT
				normalized_path,
				method,
				json_group_object(CAST(status_code AS TEXT), count) as status_codes_json
			FROM (
				SELECT
					normalized_path,
					method,
					status_code,
					COUNT(*) as count
				FROM analytics
				WHERE profile_name = ? OR (profile_name IS NULL AND ? = '')
				GROUP BY normalized_path, method, status_code
			)
			GROUP BY normalized_path, method
		)
		SELECT
			a.normalized_path,
			a.method,
			COUNT(*) as total_calls,
			SUM(CASE WHEN a.status_code >= 200 AND a.status_code < 300 THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN a.status_code >= 400 THEN 1 ELSE 0 END) as error_count,
			SUM(CASE WHEN a.status_code = 0 THEN 1 ELSE 0 END) as network_errors,
			AVG(a.duration_ms) as avg_duration,
			MIN(a.duration_ms) as min_duration,
			MAX(a.duration_ms) as max_duration,
			SUM(a.request_size) as total_req_size,
			SUM(a.response_size) as total_resp_size,
			MAX(a.timestamp) as last_called,
			COALESCE(s.status_codes_json, '{}') as status_codes_json
		FROM analytics a
		LEFT JOIN status_codes_agg s ON a.normalized_path = s.normalized_path AND a.method = s.method
		WHERE a.profile_name = ? OR (a.profile_name IS NULL AND ? = '')
		GROUP BY a.normalized_path, a.method
		ORDER BY last_called DESC
	`

	rows, err := m.db.Query(query, profileName, profileName, profileName, profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats per normalized path: %w", err)
	}
	defer rows.Close()

	var statsList []Stats
	for rows.Next() {
		var s Stats
		var lastCalled sql.NullString
		var statusCodesJSON string

		err := rows.Scan(
			&s.NormalizedPath,
			&s.Method,
			&s.TotalCalls,
			&s.SuccessCount,
			&s.ErrorCount,
			&s.NetworkErrors,
			&s.AvgDurationMs,
			&s.MinDurationMs,
			&s.MaxDurationMs,
			&s.TotalReqSize,
			&s.TotalRespSize,
			&lastCalled,
			&statusCodesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		// Handle NULL timestamp (shouldn't happen with MAX, but be safe)
		if lastCalled.Valid && lastCalled.String != "" {
			// Parse as local time (SQLite stores without timezone info)
			s.LastCalled, err = time.ParseInLocation("2006-01-02 15:04:05", lastCalled.String, time.Local)
			if err != nil {
				// Try RFC3339 format as fallback
				s.LastCalled, err = time.Parse(time.RFC3339, lastCalled.String)
				if err != nil {
					// If both fail, use current time to avoid zero value
					s.LastCalled = time.Now()
				}
			}
		} else {
			// No timestamp available, use current time
			s.LastCalled = time.Now()
		}

		// Parse status codes from JSON
		s.StatusCodes = make(map[int]int)
		if statusCodesJSON != "{}" {
			var statusCodesMap map[string]int
			if err := json.Unmarshal([]byte(statusCodesJSON), &statusCodesMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal status codes: %w", err)
			}
			// Convert string keys to int keys
			for codeStr, count := range statusCodesMap {
				var code int
				if _, err := fmt.Sscanf(codeStr, "%d", &code); err == nil {
					s.StatusCodes[code] = count
				}
			}
		}

		statsList = append(statsList, s)
	}

	return statsList, rows.Err()
}

func (m *Manager) getStatusCodesForFile(filePath string, profileName string) (map[int]int, error) {
	query := `
		SELECT status_code, COUNT(*) as count
		FROM analytics
		WHERE file_path = ? AND (profile_name = ? OR (profile_name IS NULL AND ? = ''))
		GROUP BY status_code
	`

	rows, err := m.db.Query(query, filePath, profileName, profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get status codes: %w", err)
	}
	defer rows.Close()

	statusCodes := make(map[int]int)
	for rows.Next() {
		var code, count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, err
		}
		statusCodes[code] = count
	}

	return statusCodes, rows.Err()
}

func (m *Manager) getStatusCodesForNormalizedPath(normalizedPath string, profileName string) (map[int]int, error) {
	query := `
		SELECT status_code, COUNT(*) as count
		FROM analytics
		WHERE normalized_path = ? AND (profile_name = ? OR (profile_name IS NULL AND ? = ''))
		GROUP BY status_code
	`

	rows, err := m.db.Query(query, normalizedPath, profileName, profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get status codes: %w", err)
	}
	defer rows.Close()

	statusCodes := make(map[int]int)
	for rows.Next() {
		var code, count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, err
		}
		statusCodes[code] = count
	}

	return statusCodes, rows.Err()
}

func (m *Manager) Clear() error {
	_, err := m.db.Exec("DELETE FROM analytics")
	if err != nil {
		return fmt.Errorf("failed to clear analytics: %w", err)
	}
	return nil
}

func (m *Manager) ClearForFile(filePath string) error {
	_, err := m.db.Exec("DELETE FROM analytics WHERE file_path = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to clear analytics for file: %w", err)
	}
	return nil
}

func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
