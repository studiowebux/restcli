package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/types"
)

// handleProfileKeys handles keyboard input in profile modes
func (m *Model) handleProfileKeys(msg tea.KeyMsg) tea.Cmd {
	switch m.mode {
	case ModeProfileSwitch:
		return m.handleProfileSwitchKeys(msg)
	case ModeProfileCreate:
		return m.handleProfileCreateKeys(msg)
	}
	return nil
}

// handleProfileSwitchKeys handles profile switching
func (m *Model) handleProfileSwitchKeys(msg tea.KeyMsg) tea.Cmd {
	profiles := m.sessionMgr.GetProfiles()

	switch msg.String() {
	case "esc":
		m.mode = ModeNormal

	case "up", "k":
		if m.profileIndex > 0 {
			m.profileIndex--
		}

	case "down", "j":
		if m.profileIndex < len(profiles)-1 {
			m.profileIndex++
		}

	case "enter":
		if m.profileIndex < len(profiles) {
			selectedProfile := profiles[m.profileIndex]
			m.sessionMgr.SetActiveProfile(selectedProfile.Name)
			m.mode = ModeNormal
			m.statusMsg = fmt.Sprintf("Switched to profile: %s", selectedProfile.Name)

			// Reload files from new profile's workdir
			return m.refreshFiles()
		}
	}

	return nil
}

// handleProfileCreateKeys handles profile creation
func (m *Model) handleProfileCreateKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal

	case "enter":
		if m.profileName == "" {
			m.errorMsg = "Profile name cannot be empty"
			return nil
		}

		// Create new profile
		newProfile := types.Profile{
			Name:      m.profileName,
			Workdir:   "requests",
			Headers:   make(map[string]string),
			Variables: make(map[string]types.VariableValue),
		}

		if err := m.sessionMgr.AddProfile(newProfile); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to create profile: %v", err)
			return nil
		}

		// Switch to the new profile
		m.sessionMgr.SetActiveProfile(m.profileName)
		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Created and switched to profile: %s", m.profileName)

		// Reload files
		return m.refreshFiles()

	default:
		// Handle common text input operations (paste, clear, backspace)
		if _, shouldContinue := handleTextInput(&m.profileName, msg); shouldContinue {
			return nil
		}
		// Append character to profile name
		if len(msg.String()) == 1 {
			m.profileName += msg.String()
		}
	}

	return nil
}

// renderProfileModal renders the profile switcher/creator modal
func (m *Model) renderProfileModal() string {
	var content strings.Builder

	if m.mode == ModeProfileSwitch {
		content.WriteString("Switch Profile\n\n")
		profiles := m.sessionMgr.GetProfiles()
		activeProfile := m.sessionMgr.GetActiveProfile()

		for i, profile := range profiles {
			line := fmt.Sprintf("  %s", profile.Name)
			if profile.Name == activeProfile.Name {
				line += " (current)"
			}
			if i == m.profileIndex {
				line = styleSelected.Render(line)
			}
			content.WriteString(line + "\n")
		}
		content.WriteString("\nUse ↑/↓ to select, Enter to switch, ESC to cancel")
	} else {
		content.WriteString("Create Profile\n\n")
		content.WriteString("Name: " + addCursor(m.profileName) + "\n")
		content.WriteString("\nEnter profile name, then press Enter to create, ESC to cancel")
	}

	return m.renderModal("Profile", content.String(), 50, 15)
}
