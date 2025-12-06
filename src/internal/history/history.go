package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/types"
)

// Save saves a request/response pair to history
func Save(requestFile string, req *types.HttpRequest, result *types.RequestResult) error {
	// Create history entry
	entry := types.HistoryEntry{
		Timestamp:          time.Now().Format(time.RFC3339),
		RequestFile:        requestFile,
		RequestName:        req.Name,
		Method:             req.Method,
		URL:                req.URL,
		Headers:            req.Headers,
		Body:               req.Body,
		ResponseStatus:     result.Status,
		ResponseStatusText: result.StatusText,
		ResponseHeaders:    result.Headers,
		ResponseBody:       result.Body,
		Duration:           result.Duration,
		RequestSize:        result.RequestSize,
		ResponseSize:       result.ResponseSize,
		Error:              result.Error,
	}

	// Generate filename: {requestBaseName}_{timestamp}.json
	baseName := filepath.Base(requestFile)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", baseName, timestamp)

	historyPath := filepath.Join(config.HistoryDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history entry: %w", err)
	}

	// Write to file
	if err := os.WriteFile(historyPath, data, config.FilePermissions); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// Load loads all history entries
func Load() ([]types.HistoryEntry, error) {
	entries, err := os.ReadDir(config.HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var history []types.HistoryEntry

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(config.HistoryDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var histEntry types.HistoryEntry
		if err := json.Unmarshal(data, &histEntry); err != nil {
			continue // Skip files that can't be parsed
		}

		history = append(history, histEntry)
	}

	// Sort by timestamp (newest first)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp > history[j].Timestamp
	})

	return history, nil
}

// LoadForFile loads history entries for a specific request file
func LoadForFile(requestFile string) ([]types.HistoryEntry, error) {
	baseName := filepath.Base(requestFile)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	entries, err := os.ReadDir(config.HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var history []types.HistoryEntry

	// Pattern: {baseName}_*.json
	prefix := baseName + "_"

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
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

	// Sort by timestamp (newest first)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp > history[j].Timestamp
	})

	return history, nil
}

// Clear deletes all history entries
func Clear() error {
	entries, err := os.ReadDir(config.HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read history directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(config.HistoryDir, entry.Name())
		if err := os.Remove(filePath); err != nil {
			// Continue even if some files fail to delete
			continue
		}
	}

	return nil
}

// Delete deletes a specific history entry by filename
func Delete(filename string) error {
	filePath := filepath.Join(config.HistoryDir, filename)
	return os.Remove(filePath)
}

// GetCount returns the total number of history entries
func GetCount() (int, error) {
	entries, err := os.ReadDir(config.HistoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read history directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			count++
		}
	}

	return count, nil
}
