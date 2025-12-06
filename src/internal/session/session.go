package session

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/types"
)

// Manager handles session and profile management
type Manager struct {
	session  *types.Session
	profiles []types.Profile
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		session: &types.Session{
			Variables: make(map[string]string),
		},
		profiles: []types.Profile{},
	}
}

// Load loads session and profiles from disk
func (m *Manager) Load() error {
	// Load session
	if err := m.LoadSession(); err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Load profiles
	if err := m.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	return nil
}

// LoadSession loads the session file
func (m *Manager) LoadSession() error {
	sessionPath := config.GetSessionFilePath()

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		// If file doesn't exist, use default session
		m.session = &types.Session{
			Variables: make(map[string]string),
		}
		enabled := true
		m.session.HistoryEnabled = &enabled
		return nil
	}

	var session types.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	if session.Variables == nil {
		session.Variables = make(map[string]string)
	}

	if session.HistoryEnabled == nil {
		enabled := true
		session.HistoryEnabled = &enabled
	}

	m.session = &session
	return nil
}

// SaveSession saves the session to disk
func (m *Manager) SaveSession() error {
	sessionPath := config.GetSessionFilePath()

	data, err := json.MarshalIndent(m.session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadProfiles loads the profiles file
func (m *Manager) LoadProfiles() error {
	profilesPath := config.GetProfilesFilePath()

	data, err := os.ReadFile(profilesPath)
	if err != nil {
		// If file doesn't exist, create default profile
		m.profiles = []types.Profile{
			{
				Name:      "Default",
				Workdir:   "requests",
				Headers:   make(map[string]string),
				Variables: make(map[string]types.VariableValue),
			},
		}
		return nil
	}

	var profiles []types.Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return fmt.Errorf("failed to parse profiles file: %w", err)
	}

	// Ensure all profiles have initialized maps and validate variables
	for i := range profiles {
		if profiles[i].Headers == nil {
			profiles[i].Headers = make(map[string]string)
		}
		if profiles[i].Variables == nil {
			profiles[i].Variables = make(map[string]types.VariableValue)
		}

		// Validate all variables in this profile
		for varName, varValue := range profiles[i].Variables {
			if err := varValue.Validate(varName); err != nil {
				fmt.Fprintf(os.Stderr, "warning: profile '%s': %v\n", profiles[i].Name, err)
			}
		}
	}

	m.profiles = profiles
	return nil
}

// SaveProfiles saves the profiles to disk
func (m *Manager) SaveProfiles() error {
	profilesPath := config.GetProfilesFilePath()

	data, err := json.MarshalIndent(m.profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(profilesPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profiles file: %w", err)
	}

	return nil
}

// GetSession returns the current session
func (m *Manager) GetSession() *types.Session {
	return m.session
}

// GetProfiles returns all profiles
func (m *Manager) GetProfiles() []types.Profile {
	return m.profiles
}

// GetActiveProfile returns the currently active profile
func (m *Manager) GetActiveProfile() *types.Profile {
	if m.session.ActiveProfile == "" {
		// Return first profile or create default
		if len(m.profiles) > 0 {
			return &m.profiles[0]
		}
		return &types.Profile{
			Name:      "Default",
			Workdir:   "requests",
			Headers:   make(map[string]string),
			Variables: make(map[string]types.VariableValue),
		}
	}

	// Find profile by name
	for i := range m.profiles {
		if m.profiles[i].Name == m.session.ActiveProfile {
			return &m.profiles[i]
		}
	}

	// If not found, return first profile
	if len(m.profiles) > 0 {
		return &m.profiles[0]
	}

	return &types.Profile{
		Name:      "Default",
		Workdir:   "requests",
		Headers:   make(map[string]string),
		Variables: make(map[string]types.VariableValue),
	}
}

// SetActiveProfile sets the active profile by name
func (m *Manager) SetActiveProfile(name string) error {
	// Check if profile exists
	found := false
	for _, profile := range m.profiles {
		if profile.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", name)
	}

	// Clear session variables when switching profiles
	// This prevents stale tokens/data from previous profile
	m.session.Variables = make(map[string]string)

	m.session.ActiveProfile = name
	return m.SaveSession()
}

// AddProfile adds a new profile
func (m *Manager) AddProfile(profile types.Profile) error {
	// Check for duplicate name
	for _, p := range m.profiles {
		if p.Name == profile.Name {
			return fmt.Errorf("profile already exists: %s", profile.Name)
		}
	}

	m.profiles = append(m.profiles, profile)
	return m.SaveProfiles()
}

// UpdateProfile updates an existing profile
func (m *Manager) UpdateProfile(name string, profile types.Profile) error {
	for i := range m.profiles {
		if m.profiles[i].Name == name {
			// Preserve the name if it wasn't changed
			if profile.Name == "" {
				profile.Name = name
			}
			m.profiles[i] = profile
			return m.SaveProfiles()
		}
	}

	return fmt.Errorf("profile not found: %s", name)
}

// DeleteProfile deletes a profile by name
func (m *Manager) DeleteProfile(name string) error {
	for i := range m.profiles {
		if m.profiles[i].Name == name {
			m.profiles = append(m.profiles[:i], m.profiles[i+1:]...)
			return m.SaveProfiles()
		}
	}

	return fmt.Errorf("profile not found: %s", name)
}

// SetSessionVariable sets a session variable
func (m *Manager) SetSessionVariable(name, value string) error {
	m.session.Variables[name] = value
	return m.SaveSession()
}

// GetSessionVariable gets a session variable
func (m *Manager) GetSessionVariable(name string) (string, bool) {
	value, ok := m.session.Variables[name]
	return value, ok
}

// DeleteSessionVariable deletes a session variable
func (m *Manager) DeleteSessionVariable(name string) error {
	delete(m.session.Variables, name)
	return m.SaveSession()
}

// IsHistoryEnabled returns whether history tracking is enabled
func (m *Manager) IsHistoryEnabled() bool {
	if m.session.HistoryEnabled == nil {
		return true
	}
	return *m.session.HistoryEnabled
}

// SetHistoryEnabled sets whether history tracking is enabled
func (m *Manager) SetHistoryEnabled(enabled bool) error {
	m.session.HistoryEnabled = &enabled
	return m.SaveSession()
}

// AddRecentFile adds a file to the MRU (Most Recently Used) list
// The file is added to the front of the list, and duplicates are removed
// The list is limited to maxRecentFiles (10) entries
func (m *Manager) AddRecentFile(filePath string) error {
	const maxRecentFiles = 10

	// Initialize if nil
	if m.session.RecentFiles == nil {
		m.session.RecentFiles = []string{}
	}

	// Remove duplicate if exists
	newRecent := []string{}
	for _, f := range m.session.RecentFiles {
		if f != filePath {
			newRecent = append(newRecent, f)
		}
	}

	// Add to front
	newRecent = append([]string{filePath}, newRecent...)

	// Limit to maxRecentFiles
	if len(newRecent) > maxRecentFiles {
		newRecent = newRecent[:maxRecentFiles]
	}

	m.session.RecentFiles = newRecent
	return m.SaveSession()
}

// GetRecentFiles returns the MRU file list
func (m *Manager) GetRecentFiles() []string {
	if m.session.RecentFiles == nil {
		return []string{}
	}
	return m.session.RecentFiles
}
