package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// FilePermissions is the default permission mode for regular files (read/write for owner, read for others)
	FilePermissions = 0644
	// DirPermissions is the default permission mode for directories (rwxr-xr-x)
	DirPermissions = 0755
)

var (
	// ConfigDir is the global configuration directory (~/.restcli)
	ConfigDir string

	// RequestsDir is the default requests directory
	RequestsDir string

	// HistoryDir is the history storage directory (legacy JSON files)
	HistoryDir string

	// DatabasePath is the SQLite database file for history and analytics
	DatabasePath string

	// SessionFile is the session state file
	SessionFile string

	// ProfilesFile is the profiles configuration file
	ProfilesFile string
)

// Initialize sets up the configuration directories and files
// It creates ~/.restcli/ if it doesn't exist
func Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Set global paths
	ConfigDir = filepath.Join(homeDir, ".restcli")
	RequestsDir = filepath.Join(ConfigDir, "requests")
	HistoryDir = filepath.Join(ConfigDir, "history")
	DatabasePath = filepath.Join(ConfigDir, "restcli.db")
	SessionFile = filepath.Join(ConfigDir, ".session.json")
	ProfilesFile = filepath.Join(ConfigDir, ".profiles.json")

	// Create directories if they don't exist
	dirs := []string{ConfigDir, RequestsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, DirPermissions); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create empty session file if it doesn't exist
	if _, err := os.Stat(SessionFile); os.IsNotExist(err) {
		defaultSession := []byte(`{"variables":{},"historyEnabled":true}`)
		if err := os.WriteFile(SessionFile, defaultSession, FilePermissions); err != nil {
			return fmt.Errorf("failed to create session file: %w", err)
		}
	}

	// Create empty profiles file if it doesn't exist
	if _, err := os.Stat(ProfilesFile); os.IsNotExist(err) {
		defaultProfiles := []byte(`[{"name":"Default","workdir":".restcli/requests","headers":{},"variables":{}}]`)
		if err := os.WriteFile(ProfilesFile, defaultProfiles, FilePermissions); err != nil {
			return fmt.Errorf("failed to create profiles file: %w", err)
		}
	}

	return nil
}

// GetWorkingDirectory returns the working directory for a profile
// Falls back to global requests directory if profile workdir is not set
func GetWorkingDirectory(profileWorkdir string) (string, error) {
	if profileWorkdir == "" {
		return RequestsDir, nil
	}

	// Expand tilde to home directory
	if strings.HasPrefix(profileWorkdir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		profileWorkdir = filepath.Join(homeDir, profileWorkdir[2:])
	}

	// If it's an absolute path, use it directly
	if filepath.IsAbs(profileWorkdir) {
		return profileWorkdir, nil
	}

	// Otherwise, it's relative to config directory
	workdir := filepath.Join(ConfigDir, profileWorkdir)

	// Ensure the directory exists
	if err := os.MkdirAll(workdir, DirPermissions); err != nil {
		return "", fmt.Errorf("failed to create working directory %s: %w", workdir, err)
	}

	return workdir, nil
}

// LocalConfigExists checks if there's a local .session.json or .profiles.json
func LocalConfigExists() bool {
	_, sessionErr := os.Stat(".session.json")
	_, profilesErr := os.Stat(".profiles.json")
	return sessionErr == nil || profilesErr == nil
}

// GetSessionFilePath returns the session file path (local or global)
func GetSessionFilePath() string {
	if _, err := os.Stat(".session.json"); err == nil {
		return ".session.json"
	}
	return SessionFile
}

// GetProfilesFilePath returns the profiles file path (local or global)
func GetProfilesFilePath() string {
	if _, err := os.Stat(".profiles.json"); err == nil {
		return ".profiles.json"
	}
	return ProfilesFile
}
