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
	case ModeProfileEdit:
		return m.handleProfileEditKeys(msg)
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

	case "e":
		// Edit the selected profile
		if m.profileIndex < len(profiles) {
			profile := profiles[m.profileIndex]
			m.mode = ModeProfileEdit
			m.profileEditField = 0
			m.profileEditName = profile.Name
			m.profileEditWorkdir = profile.Workdir
			m.profileEditEditor = profile.Editor
			m.profileEditOutput = profile.Output
			m.profileEditNamePos = len(profile.Name)
			m.profileEditWorkdirPos = len(profile.Workdir)
			m.profileEditEditorPos = len(profile.Editor)
			m.profileEditOutputPos = len(profile.Output)
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

// handleProfileEditKeys handles profile editing
func (m *Model) handleProfileEditKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = ModeProfileSwitch

	case "tab":
		// Move to next field
		m.profileEditField = (m.profileEditField + 1) % 4

	case "shift+tab":
		// Move to previous field
		m.profileEditField--
		if m.profileEditField < 0 {
			m.profileEditField = 3
		}

	case "enter":
		// Save profile changes
		profiles := m.sessionMgr.GetProfiles()
		if m.profileIndex < len(profiles) {
			profile := &profiles[m.profileIndex]
			oldName := profile.Name

			profile.Name = m.profileEditName
			profile.Workdir = m.profileEditWorkdir
			profile.Editor = m.profileEditEditor
			profile.Output = m.profileEditOutput

			m.sessionMgr.SaveProfiles()

			// If active profile name changed, update it
			activeProfile := m.sessionMgr.GetActiveProfile()
			if activeProfile.Name == oldName {
				m.sessionMgr.SetActiveProfile(m.profileEditName)
			}

			m.mode = ModeNormal
			m.statusMsg = fmt.Sprintf("Profile saved: %s", m.profileEditName)

			// Reload files if workdir changed
			return m.refreshFiles()
		}

	default:
		// Handle text input for current field
		var input *string
		var cursorPos *int

		switch m.profileEditField {
		case 0:
			input = &m.profileEditName
			cursorPos = &m.profileEditNamePos
		case 1:
			input = &m.profileEditWorkdir
			cursorPos = &m.profileEditWorkdirPos
		case 2:
			input = &m.profileEditEditor
			cursorPos = &m.profileEditEditorPos
		case 3:
			input = &m.profileEditOutput
			cursorPos = &m.profileEditOutputPos
		}

		if input != nil && cursorPos != nil {
			if _, shouldContinue := handleTextInputWithCursor(input, cursorPos, msg); shouldContinue {
				return nil
			}
			// Append character
			if len(msg.String()) == 1 {
				*input = (*input)[:*cursorPos] + msg.String() + (*input)[*cursorPos:]
				*cursorPos++
			}
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
		content.WriteString("\n↑/↓ select [Enter] switch [e]dit [ESC] cancel")
		return m.renderModal("Profile", content.String(), 50, 15)

	} else if m.mode == ModeProfileEdit {
		content.WriteString("Edit Profile\n\n")

		// Modal width for field display (account for border, padding, label)
		fieldWidth := 50

		// Helper to add cursor with horizontal scrolling
		addFieldCursorScrolling := func(value string, cursorPos int, maxWidth int) string {
			if cursorPos > len(value) {
				cursorPos = len(value)
			}

			// If text fits, just add cursor
			if len(value) < maxWidth-1 {
				return value[:cursorPos] + "█" + value[cursorPos:]
			}

			// Calculate visible window that keeps cursor in view
			// Reserve 1 char for cursor
			visibleWidth := maxWidth - 1

			// Calculate start position to keep cursor visible
			start := 0
			if cursorPos > visibleWidth-5 {
				start = cursorPos - visibleWidth + 5
			}
			if start < 0 {
				start = 0
			}

			// Calculate end position
			end := start + visibleWidth
			if end > len(value) {
				end = len(value)
				start = end - visibleWidth
				if start < 0 {
					start = 0
				}
			}

			// Build visible portion with cursor
			visible := value[start:end]
			cursorInWindow := cursorPos - start

			// Add scroll indicators
			prefix := ""
			suffix := ""
			if start > 0 {
				prefix = "…"
				visible = visible[1:]
				cursorInWindow--
			}
			if end < len(value) {
				suffix = "…"
				visible = visible[:len(visible)-1]
			}

			// Adjust cursor position for prefix
			if cursorInWindow < 0 {
				cursorInWindow = 0
			}
			if cursorInWindow > len(visible) {
				cursorInWindow = len(visible)
			}

			return prefix + visible[:cursorInWindow] + "█" + visible[cursorInWindow:] + suffix
		}

		// Helper to truncate non-active fields
		truncateField := func(value string, maxWidth int) string {
			if len(value) <= maxWidth {
				return value
			}
			return value[:maxWidth-1] + "…"
		}

		// Name field
		nameLabel := "Name:    "
		if m.profileEditField == 0 {
			nameLabel = styleSelected.Render(nameLabel)
			content.WriteString(nameLabel + addFieldCursorScrolling(m.profileEditName, m.profileEditNamePos, fieldWidth) + "\n")
		} else {
			content.WriteString(nameLabel + truncateField(m.profileEditName, fieldWidth) + "\n")
		}

		// Workdir field
		workdirLabel := "Workdir: "
		if m.profileEditField == 1 {
			workdirLabel = styleSelected.Render(workdirLabel)
			content.WriteString(workdirLabel + addFieldCursorScrolling(m.profileEditWorkdir, m.profileEditWorkdirPos, fieldWidth) + "\n")
		} else {
			content.WriteString(workdirLabel + truncateField(m.profileEditWorkdir, fieldWidth) + "\n")
		}

		// Editor field
		editorLabel := "Editor:  "
		if m.profileEditField == 2 {
			editorLabel = styleSelected.Render(editorLabel)
			content.WriteString(editorLabel + addFieldCursorScrolling(m.profileEditEditor, m.profileEditEditorPos, fieldWidth) + "\n")
		} else {
			content.WriteString(editorLabel + truncateField(m.profileEditEditor, fieldWidth) + "\n")
		}

		// Output field
		outputLabel := "Output:  "
		if m.profileEditField == 3 {
			outputLabel = styleSelected.Render(outputLabel)
			content.WriteString(outputLabel + addFieldCursorScrolling(m.profileEditOutput, m.profileEditOutputPos, fieldWidth) + "\n")
		} else {
			content.WriteString(outputLabel + truncateField(m.profileEditOutput, fieldWidth) + "\n")
		}

		content.WriteString("\nOutput: json, yaml, or text")
		content.WriteString("\n\n[TAB] next [Enter] save [ESC] cancel")
		return m.renderModal("Profile", content.String(), 70, 18)

	} else {
		content.WriteString("Create Profile\n\n")
		content.WriteString("Name: " + addCursor(m.profileName) + "\n")
		content.WriteString("\n[Enter] create [ESC] cancel")
		return m.renderModal("Profile", content.String(), 50, 15)
	}
}
