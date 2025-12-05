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
	case ModeProfileDuplicate:
		return m.handleProfileDuplicateKeys(msg)
	case ModeProfileDeleteConfirm:
		return m.handleProfileDeleteConfirmKeys(msg)
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
			m.profileEditHistoryEnabled = profile.HistoryEnabled
			m.profileEditAnalyticsEnabled = profile.AnalyticsEnabled
			m.profileEditNamePos = len(profile.Name)
			m.profileEditWorkdirPos = len(profile.Workdir)
			m.profileEditEditorPos = len(profile.Editor)
			m.profileEditOutputPos = len(profile.Output)
		}

	case "d":
		// Duplicate the selected profile
		if m.profileIndex < len(profiles) {
			m.mode = ModeProfileDuplicate
			m.profileName = ""
			m.profileNamePos = 0
			m.errorMsg = ""
		}

	case "D":
		// Delete the selected profile (show confirmation)
		if m.profileIndex < len(profiles) {
			m.mode = ModeProfileDeleteConfirm
			m.errorMsg = ""
		}

	case "n":
		// Create new profile
		m.mode = ModeProfileCreate
		m.profileName = ""
		m.profileNamePos = 0
		m.errorMsg = ""
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
			Workdir:   ".restcli/requests",
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

// handleProfileDuplicateKeys handles profile duplication
func (m *Model) handleProfileDuplicateKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = ModeProfileSwitch

	case "enter":
		if m.profileName == "" {
			m.errorMsg = "Profile name cannot be empty"
			return nil
		}

		// Get the source profile to duplicate
		profiles := m.sessionMgr.GetProfiles()
		if m.profileIndex >= len(profiles) {
			m.errorMsg = "Invalid profile selection"
			return nil
		}

		sourceProfile := profiles[m.profileIndex]

		// Create new profile with all settings from source
		newProfile := types.Profile{
			Name:              m.profileName,
			Workdir:           sourceProfile.Workdir,
			Editor:            sourceProfile.Editor,
			Output:            sourceProfile.Output,
			HistoryEnabled:    sourceProfile.HistoryEnabled,
			AnalyticsEnabled:  sourceProfile.AnalyticsEnabled,
			Headers:           make(map[string]string),
			Variables:         make(map[string]types.VariableValue),
		}

		// Deep copy headers
		for k, v := range sourceProfile.Headers {
			newProfile.Headers[k] = v
		}

		// Deep copy variables
		for k, v := range sourceProfile.Variables {
			newProfile.Variables[k] = v
		}

		// Copy OAuth settings if present
		if sourceProfile.OAuth != nil {
			oauthCopy := *sourceProfile.OAuth
			newProfile.OAuth = &oauthCopy
		}

		if err := m.sessionMgr.AddProfile(newProfile); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to duplicate profile: %v", err)
			return nil
		}

		// Switch to the new profile
		m.sessionMgr.SetActiveProfile(m.profileName)
		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Duplicated profile '%s' as '%s'", sourceProfile.Name, m.profileName)

		// Reload files
		return m.refreshFiles()

	default:
		// Handle text input with cursor support
		if _, shouldContinue := handleTextInputWithCursor(&m.profileName, &m.profileNamePos, msg); shouldContinue {
			return nil
		}
		// Insert character at cursor position
		if len(msg.String()) == 1 {
			m.profileName = m.profileName[:m.profileNamePos] + msg.String() + m.profileName[m.profileNamePos:]
			m.profileNamePos++
		}
	}

	return nil
}

// handleProfileDeleteConfirmKeys handles profile deletion confirmation
func (m *Model) handleProfileDeleteConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n", "N":
		m.mode = ModeProfileSwitch
		m.errorMsg = ""

	case "y", "Y":
		profiles := m.sessionMgr.GetProfiles()
		if m.profileIndex >= len(profiles) {
			m.errorMsg = "Invalid profile selection"
			m.mode = ModeProfileSwitch
			return nil
		}

		profileToDelete := profiles[m.profileIndex]
		activeProfile := m.sessionMgr.GetActiveProfile()

		// Prevent deleting the last profile
		if len(profiles) <= 1 {
			m.errorMsg = "Cannot delete the last profile"
			m.mode = ModeProfileSwitch
			return nil
		}

		// Prevent deleting the active profile
		if profileToDelete.Name == activeProfile.Name {
			m.errorMsg = "Cannot delete active profile. Switch to another profile first."
			m.mode = ModeProfileSwitch
			return nil
		}

		// Delete the profile
		if err := m.sessionMgr.DeleteProfile(profileToDelete.Name); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to delete profile: %v", err)
			m.mode = ModeProfileSwitch
			return nil
		}

		// Adjust profileIndex if needed
		if m.profileIndex >= len(profiles)-1 {
			m.profileIndex = len(profiles) - 2
		}

		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Deleted profile: %s", profileToDelete.Name)
		return nil
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
		m.profileEditField = (m.profileEditField + 1) % 6

	case "shift+tab":
		// Move to previous field
		m.profileEditField--
		if m.profileEditField < 0 {
			m.profileEditField = 5
		}

	case " ":
		// Toggle history setting (field 4)
		if m.profileEditField == 4 {
			if m.profileEditHistoryEnabled == nil {
				// nil -> true
				enabled := true
				m.profileEditHistoryEnabled = &enabled
			} else if *m.profileEditHistoryEnabled {
				// true -> false
				disabled := false
				m.profileEditHistoryEnabled = &disabled
			} else {
				// false -> nil (default)
				m.profileEditHistoryEnabled = nil
			}
		}
		// Toggle analytics setting (field 5)
		if m.profileEditField == 5 {
			if m.profileEditAnalyticsEnabled == nil {
				// nil -> true
				enabled := true
				m.profileEditAnalyticsEnabled = &enabled
			} else if *m.profileEditAnalyticsEnabled {
				// true -> false
				disabled := false
				m.profileEditAnalyticsEnabled = &disabled
			} else {
				// false -> nil (default/false)
				m.profileEditAnalyticsEnabled = nil
			}
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
			profile.HistoryEnabled = m.profileEditHistoryEnabled
			profile.AnalyticsEnabled = m.profileEditAnalyticsEnabled

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
		footer := "↑/↓ select [Enter] switch [e]dit [d]uplicate [D]elete | [n]ew [ESC] cancel"
		// Use auto-scroll to keep selected profile visible
		return m.renderModalWithFooterAndScroll("Switch Profile", content.String(), footer, 70, 15, m.profileIndex)

	} else if m.mode == ModeProfileEdit {

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

		// History field (toggle)
		historyLabel := "History:   "
		historyValue := "default"
		if m.profileEditHistoryEnabled != nil {
			if *m.profileEditHistoryEnabled {
				historyValue = "enabled"
			} else {
				historyValue = "disabled"
			}
		}
		if m.profileEditField == 4 {
			historyLabel = styleSelected.Render(historyLabel)
			content.WriteString(historyLabel + styleSelected.Render(historyValue) + " [SPACE to toggle]\n")
		} else {
			content.WriteString(historyLabel + historyValue + "\n")
		}

		// Analytics field (toggle)
		analyticsLabel := "Analytics: "
		analyticsValue := "disabled"
		if m.profileEditAnalyticsEnabled != nil {
			if *m.profileEditAnalyticsEnabled {
				analyticsValue = "enabled"
			} else {
				analyticsValue = "disabled"
			}
		}
		if m.profileEditField == 5 {
			analyticsLabel = styleSelected.Render(analyticsLabel)
			content.WriteString(analyticsLabel + styleSelected.Render(analyticsValue) + " [SPACE to toggle]\n")
		} else {
			content.WriteString(analyticsLabel + analyticsValue + "\n")
		}

		content.WriteString("\nOutput: json, yaml, or text")
		content.WriteString("\nHistory: default (uses global), enabled, or disabled")
		content.WriteString("\nAnalytics: disabled (default), or enabled")
		footer := "[TAB] next [SPACE] toggle [Enter] save [ESC] cancel"
		return m.renderModalWithFooter("Edit Profile", content.String(), footer, 70, 22)

	} else if m.mode == ModeProfileDuplicate {
		profiles := m.sessionMgr.GetProfiles()
		if m.profileIndex < len(profiles) {
			sourceProfile := profiles[m.profileIndex]
			content.WriteString(fmt.Sprintf("Duplicating: %s\n\n", styleSuccess.Render(sourceProfile.Name)))

			// Show cursor in input
			inputWithCursor := m.profileName[:m.profileNamePos] + "█" + m.profileName[m.profileNamePos:]
			content.WriteString("New name: " + inputWithCursor + "\n")

			if m.errorMsg != "" {
				content.WriteString("\n" + styleError.Render(m.errorMsg))
			}

			content.WriteString("\n\nAll settings will be copied:")
			content.WriteString("\n• Workdir, editor, output")
			content.WriteString("\n• Headers and variables")
			content.WriteString("\n• OAuth configuration")
			footer := "[Enter] duplicate [ESC] cancel"
			return m.renderModalWithFooter("Duplicate Profile", content.String(), footer, 60, 18)
		}
		// Fallback if invalid selection
		content.WriteString("Invalid profile selection")
		return m.renderModal("Profile", content.String(), 50, 10)

	} else if m.mode == ModeProfileDeleteConfirm {
		profiles := m.sessionMgr.GetProfiles()
		if m.profileIndex < len(profiles) {
			profileToDelete := profiles[m.profileIndex]
			content.WriteString(fmt.Sprintf("Delete profile: %s?\n\n", styleError.Render(profileToDelete.Name)))
			content.WriteString("This action cannot be undone.\n")
			content.WriteString("All settings will be permanently removed.")

			if m.errorMsg != "" {
				content.WriteString("\n\n" + styleError.Render(m.errorMsg))
			}

			footer := "[y]es [n]o / ESC"
			return m.renderModalWithFooter("Delete Profile", content.String(), footer, 60, 14)
		}
		// Fallback if invalid selection
		content.WriteString("Invalid profile selection")
		return m.renderModal("Error", content.String(), 50, 10)

	} else {
		content.WriteString("Name: " + addCursor(m.profileName) + "\n")
		footer := "[Enter] create [ESC] cancel"
		return m.renderModalWithFooter("Create Profile", content.String(), footer, 50, 12)
	}
}
