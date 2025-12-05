package history

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
	"github.com/studiowebux/restcli/internal/types"
)

type Manager struct {
	db *sql.DB
}

func NewManager(dbPath string) (*Manager, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to history database: %w", err)
	}

	m := &Manager{db: db}
	if err := m.initSchema(); err != nil {
		return nil, err
	}

	// Run database migrations
	if err := migrations.Run(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Auto-migrate from JSON files if database is empty
	if err := m.migrateFromJSONIfNeeded(); err != nil {
		// Log error but don't fail - migration is best-effort
		_ = err
	}

	return m, nil
}

func (m *Manager) initSchema() error {
	schema := `
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
		error TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_history_request_file ON history(request_file);
	CREATE INDEX IF NOT EXISTS idx_history_method ON history(method);
	CREATE INDEX IF NOT EXISTS idx_history_url ON history(url);
	`

	_, err := m.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize history schema: %w", err)
	}

	return nil
}

func (m *Manager) Save(requestFile string, profileName string, req *types.HttpRequest, result *types.RequestResult) error {
	// Serialize headers to JSON
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	// Serialize response headers to JSON
	responseHeadersJSON, err := json.Marshal(result.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal response headers: %w", err)
	}

	query := `
		INSERT INTO history (
			timestamp, request_file, request_name, method, url, headers, body,
			response_status, response_status_text, response_headers, response_body,
			duration_ms, request_size, response_size, error, profile_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Format timestamp for SQLite in local time
	timestampStr := time.Now().Local().Format("2006-01-02 15:04:05")

	_, err = m.db.Exec(query,
		timestampStr,
		requestFile,
		req.Name,
		req.Method,
		req.URL,
		string(headersJSON),
		req.Body,
		result.Status,
		result.StatusText,
		string(responseHeadersJSON),
		result.Body,
		result.Duration,
		result.RequestSize,
		result.ResponseSize,
		result.Error,
		profileName,
	)

	if err != nil {
		return fmt.Errorf("failed to save history entry: %w", err)
	}

	return nil
}

func (m *Manager) Load(profileName string) ([]types.HistoryEntry, error) {
	query := `
		SELECT id, timestamp, request_file, request_name, method, url, headers, body,
		       response_status, response_status_text, response_headers, response_body,
		       duration_ms, request_size, response_size, error, COALESCE(profile_name, '')
		FROM history
		WHERE profile_name = ? OR (profile_name IS NULL AND ? = '')
		ORDER BY timestamp DESC
	`

	rows, err := m.db.Query(query, profileName, profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}
	defer rows.Close()

	return m.scanEntries(rows)
}

func (m *Manager) LoadForFile(requestFile string) ([]types.HistoryEntry, error) {
	// Extract base name for matching (same logic as JSON file implementation)
	baseName := filepath.Base(requestFile)
	baseName = filepath.Base(baseName) // Ensure we have just the filename

	query := `
		SELECT id, timestamp, request_file, request_name, method, url, headers, body,
		       response_status, response_status_text, response_headers, response_body,
		       duration_ms, request_size, response_size, error
		FROM history
		WHERE request_file LIKE ?
		ORDER BY timestamp DESC
	`

	// Match both exact path and basename patterns
	pattern := "%" + baseName + "%"

	rows, err := m.db.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to load history for file: %w", err)
	}
	defer rows.Close()

	return m.scanEntries(rows)
}

func (m *Manager) scanEntries(rows *sql.Rows) ([]types.HistoryEntry, error) {
	var entries []types.HistoryEntry

	for rows.Next() {
		var id int64
		var timestamp string
		var requestFile string
		var requestName sql.NullString
		var method string
		var url string
		var headersJSON string
		var body sql.NullString
		var responseStatus int
		var responseStatusText string
		var responseHeadersJSON string
		var responseBody string
		var durationMs int64
		var requestSize sql.NullInt64
		var responseSize sql.NullInt64
		var errorMsg sql.NullString
		var profileName string

		err := rows.Scan(
			&id,
			&timestamp,
			&requestFile,
			&requestName,
			&method,
			&url,
			&headersJSON,
			&body,
			&responseStatus,
			&responseStatusText,
			&responseHeadersJSON,
			&responseBody,
			&durationMs,
			&requestSize,
			&responseSize,
			&errorMsg,
			&profileName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history entry: %w", err)
		}

		// Deserialize headers
		var headers map[string]string
		if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
			headers = make(map[string]string)
		}

		// Deserialize response headers
		var responseHeaders map[string]string
		if err := json.Unmarshal([]byte(responseHeadersJSON), &responseHeaders); err != nil {
			responseHeaders = make(map[string]string)
		}

		// Parse timestamp as local time
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", timestamp, time.Local)
		if err != nil {
			// Try RFC3339 format as fallback
			parsedTime, err = time.Parse(time.RFC3339, timestamp)
			if err != nil {
				// If both fail, use current time
				parsedTime = time.Now()
			}
		}

		entry := types.HistoryEntry{
			Timestamp:          parsedTime.Format(time.RFC3339),
			RequestFile:        requestFile,
			RequestName:        requestName.String,
			Method:             method,
			URL:                url,
			Headers:            headers,
			Body:               body.String,
			ResponseStatus:     responseStatus,
			ResponseStatusText: responseStatusText,
			ResponseHeaders:    responseHeaders,
			ResponseBody:       responseBody,
			Duration:           durationMs,
			RequestSize:        int(requestSize.Int64),
			ResponseSize:       int(responseSize.Int64),
			Error:              errorMsg.String,
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (m *Manager) Clear() error {
	_, err := m.db.Exec("DELETE FROM history")
	if err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}
	return nil
}

func (m *Manager) Delete(id int64) error {
	_, err := m.db.Exec("DELETE FROM history WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete history entry: %w", err)
	}
	return nil
}

func (m *Manager) GetCount() (int, error) {
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get history count: %w", err)
	}
	return count, nil
}

func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// migrateFromJSONIfNeeded checks if database is empty and migrates from JSON files if needed
func (m *Manager) migrateFromJSONIfNeeded() error {
	// Check if database already has entries
	count, err := m.GetCount()
	if err != nil {
		return err
	}

	// Database not empty, skip migration
	if count > 0 {
		return nil
	}

	// Check if JSON history directory exists
	if _, err := os.Stat(config.HistoryDir); os.IsNotExist(err) {
		return nil // No JSON files to migrate
	}

	// Load all JSON entries using existing Load function
	entries, err := loadFromJSON()
	if err != nil {
		return fmt.Errorf("failed to load JSON history: %w", err)
	}

	if len(entries) == 0 {
		return nil // No entries to migrate
	}

	// Migrate each entry
	migrated := 0
	for _, entry := range entries {
		if err := m.insertHistoryEntry(entry); err != nil {
			// Continue on error, log it but don't fail entire migration
			continue
		}
		migrated++
	}

	// If migration was successful, rename history directory to backup
	if migrated > 0 {
		backupDir := config.HistoryDir + ".backup"
		if err := os.Rename(config.HistoryDir, backupDir); err != nil {
			// Don't fail if backup rename fails
			_ = err
		}
	}

	return nil
}

// loadFromJSON loads history entries from JSON files (for migration)
func loadFromJSON() ([]types.HistoryEntry, error) {
	entries, err := os.ReadDir(config.HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var history []types.HistoryEntry

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(config.HistoryDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var histEntry types.HistoryEntry
		if err := json.Unmarshal(data, &histEntry); err != nil {
			continue
		}

		history = append(history, histEntry)
	}

	return history, nil
}

// insertHistoryEntry inserts a HistoryEntry into the database (for migration)
func (m *Manager) insertHistoryEntry(entry types.HistoryEntry) error {
	// Serialize headers to JSON
	headersJSON, err := json.Marshal(entry.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	// Serialize response headers to JSON
	responseHeadersJSON, err := json.Marshal(entry.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("failed to marshal response headers: %w", err)
	}

	query := `
		INSERT INTO history (
			timestamp, request_file, request_name, method, url, headers, body,
			response_status, response_status_text, response_headers, response_body,
			duration_ms, request_size, response_size, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Parse RFC3339 timestamp and convert to SQLite format
	parsedTime, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		parsedTime = time.Now()
	}
	timestampStr := parsedTime.Local().Format("2006-01-02 15:04:05")

	_, err = m.db.Exec(query,
		timestampStr,
		entry.RequestFile,
		entry.RequestName,
		entry.Method,
		entry.URL,
		string(headersJSON),
		entry.Body,
		entry.ResponseStatus,
		entry.ResponseStatusText,
		string(responseHeadersJSON),
		entry.ResponseBody,
		entry.Duration,
		entry.RequestSize,
		entry.ResponseSize,
		entry.Error,
	)

	if err != nil {
		return fmt.Errorf("failed to insert history entry: %w", err)
	}

	return nil
}
