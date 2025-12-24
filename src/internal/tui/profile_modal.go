package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/keybinds"
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

	// Handle 'e' specially (edit - not in registry)
	if msg.String() == "e" {
		if m.profileIndex < len(profiles) {
			profile := profiles[m.profileIndex]
			m.mode = ModeProfileEdit
			m.profileEditState.SetField(0)
			m.profileEditState.LoadFromProfile(
				profile.Name,
				profile.Workdir,
				profile.Editor,
				profile.Output,
				profile.HistoryEnabled,
				profile.AnalyticsEnabled,
			)
		}
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextProfileList, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionNavigateUp:
		if m.profileIndex > 0 {
			m.profileIndex--
		}

	case keybinds.ActionNavigateDown:
		if m.profileIndex < len(profiles)-1 {
			m.profileIndex++
		}

	case keybinds.ActionProfileSwitch:
		if m.profileIndex < len(profiles) {
			selectedProfile := profiles[m.profileIndex]
			m.sessionMgr.SetActiveProfile(selectedProfile.Name)
			m.mode = ModeNormal
			m.statusMsg = fmt.Sprintf("Switched to profile: %s", selectedProfile.Name)

			// Reload files from new profile's workdir
			return m.refreshFiles()
		}

	case keybinds.ActionProfileDuplicate:
		if m.profileIndex < len(profiles) {
			m.mode = ModeProfileDuplicate
			m.profileName = ""
			m.profileNamePos = 0
			m.errorMsg = ""
		}

	case keybinds.ActionProfileDelete:
		if m.profileIndex < len(profiles) {
			m.mode = ModeProfileDeleteConfirm
			m.errorMsg = ""
		}

	case keybinds.ActionProfileCreate:
		m.mode = ModeProfileCreate
		m.profileName = ""
		m.profileNamePos = 0
		m.errorMsg = ""
	}

	return nil
}

// handleProfileCreateKeys handles profile creation
func (m *Model) handleProfileCreateKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextProfileEdit, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeNormal
			return nil

		case keybinds.ActionTextSubmit:
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
		}
	}

	// Handle common text input operations (paste, clear, backspace)
	if _, shouldContinue := handleTextInput(&m.profileName, msg); shouldContinue {
		return nil
	}
	// Append character to profile name
	if len(msg.String()) == 1 {
		m.profileName += msg.String()
	}

	return nil
}

// handleProfileDuplicateKeys handles profile duplication
func (m *Model) handleProfileDuplicateKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextProfileEdit, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeProfileSwitch
			return nil

		case keybinds.ActionTextSubmit:
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
		}
	}

	// Handle text input with cursor support
	if _, shouldContinue := handleTextInputWithCursor(&m.profileName, &m.profileNamePos, msg); shouldContinue {
		return nil
	}
	// Insert character at cursor position
	if len(msg.String()) == 1 {
		m.profileName = m.profileName[:m.profileNamePos] + msg.String() + m.profileName[m.profileNamePos:]
		m.profileNamePos++
	}

	return nil
}

// handleProfileDeleteConfirmKeys handles profile deletion confirmation
func (m *Model) handleProfileDeleteConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextConfirm, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCancel:
		m.mode = ModeProfileSwitch
		m.errorMsg = ""

	case keybinds.ActionConfirm:
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
	// Handle special keys not in registry
	switch msg.String() {
	case "tab":
		// Move to next field
		m.profileEditState.SetField((m.profileEditState.GetField() + 1) % 6)
		return nil

	case "shift+tab":
		// Move to previous field
		m.profileEditState.Navigate(-1, 6)
		return nil

	case " ":
		// Toggle history setting (field 4)
		if m.profileEditState.GetField() == 4 {
			if m.profileEditState.GetHistoryEnabled() == nil {
				// nil -> true
				enabled := true
				m.profileEditState.SetHistoryEnabled(&enabled)
			} else if *m.profileEditState.GetHistoryEnabled() {
				// true -> false
				disabled := false
				m.profileEditState.SetHistoryEnabled(&disabled)
			} else {
				// false -> nil (default)
				m.profileEditState.SetHistoryEnabled(nil)
			}
			return nil
		}
		// Toggle analytics setting (field 5)
		if m.profileEditState.GetField() == 5 {
			if m.profileEditState.GetAnalyticsEnabled() == nil {
				// nil -> true
				enabled := true
				m.profileEditState.SetAnalyticsEnabled(&enabled)
			} else if *m.profileEditState.GetAnalyticsEnabled() {
				// true -> false
				disabled := false
				m.profileEditState.SetAnalyticsEnabled(&disabled)
			} else {
				// false -> nil (default/false)
				m.profileEditState.SetAnalyticsEnabled(nil)
			}
			return nil
		}
		// For other fields, let space fall through to character input
	}

	action, ok := m.keybinds.Match(keybinds.ContextProfileEdit, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeProfileSwitch
			return nil

		case keybinds.ActionTextSubmit:
			// Save profile changes
			profiles := m.sessionMgr.GetProfiles()
			if m.profileIndex < len(profiles) {
				profile := &profiles[m.profileIndex]
				oldName := profile.Name

				profile.Name = m.profileEditState.GetName()
				profile.Workdir = m.profileEditState.GetWorkdir()
				profile.Editor = m.profileEditState.GetEditor()
				profile.Output = m.profileEditState.GetOutput()
				profile.HistoryEnabled = m.profileEditState.GetHistoryEnabled()
				profile.AnalyticsEnabled = m.profileEditState.GetAnalyticsEnabled()

				m.sessionMgr.SaveProfiles()

				// If active profile name changed, update it
				activeProfile := m.sessionMgr.GetActiveProfile()
				if activeProfile.Name == oldName {
					m.sessionMgr.SetActiveProfile(m.profileEditState.GetName())
				}

				m.mode = ModeNormal
				m.statusMsg = fmt.Sprintf("Profile saved: %s", m.profileEditState.GetName())

				// Reload files if workdir changed
				return m.refreshFiles()
			}
			return nil
		}
	}

	// Handle text input for current field
	switch m.profileEditState.GetField() {
	case 0:
		input := m.profileEditState.GetName()
		cursorPos := m.profileEditState.GetNamePos()
		if _, shouldContinue := handleTextInputWithCursor(&input, &cursorPos, msg); shouldContinue {
			m.profileEditState.SetName(input)
			m.profileEditState.SetNamePos(cursorPos)
			return nil
		}
		// Append character
		if len(msg.String()) == 1 {
			input = input[:cursorPos] + msg.String() + input[cursorPos:]
			cursorPos++
			m.profileEditState.SetName(input)
			m.profileEditState.SetNamePos(cursorPos)
		}
	case 1:
		input := m.profileEditState.GetWorkdir()
		cursorPos := m.profileEditState.GetWorkdirPos()
		if _, shouldContinue := handleTextInputWithCursor(&input, &cursorPos, msg); shouldContinue {
			m.profileEditState.SetWorkdir(input)
			m.profileEditState.SetWorkdirPos(cursorPos)
			return nil
		}
		// Append character
		if len(msg.String()) == 1 {
			input = input[:cursorPos] + msg.String() + input[cursorPos:]
			cursorPos++
			m.profileEditState.SetWorkdir(input)
			m.profileEditState.SetWorkdirPos(cursorPos)
		}
	case 2:
		input := m.profileEditState.GetEditor()
		cursorPos := m.profileEditState.GetEditorPos()
		if _, shouldContinue := handleTextInputWithCursor(&input, &cursorPos, msg); shouldContinue {
			m.profileEditState.SetEditor(input)
			m.profileEditState.SetEditorPos(cursorPos)
			return nil
		}
		// Append character
		if len(msg.String()) == 1 {
			input = input[:cursorPos] + msg.String() + input[cursorPos:]
			cursorPos++
			m.profileEditState.SetEditor(input)
			m.profileEditState.SetEditorPos(cursorPos)
		}
	case 3:
		input := m.profileEditState.GetOutput()
		cursorPos := m.profileEditState.GetOutputPos()
		if _, shouldContinue := handleTextInputWithCursor(&input, &cursorPos, msg); shouldContinue {
			m.profileEditState.SetOutput(input)
			m.profileEditState.SetOutputPos(cursorPos)
			return nil
		}
		// Append character
		if len(msg.String()) == 1 {
			input = input[:cursorPos] + msg.String() + input[cursorPos:]
			cursorPos++
			m.profileEditState.SetOutput(input)
			m.profileEditState.SetOutputPos(cursorPos)
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
		if m.profileEditState.GetField() == 0 {
			nameLabel = styleSelected.Render(nameLabel)
			content.WriteString(nameLabel + addFieldCursorScrolling(m.profileEditState.GetName(), m.profileEditState.GetNamePos(), fieldWidth) + "\n")
		} else {
			content.WriteString(nameLabel + truncateField(m.profileEditState.GetName(), fieldWidth) + "\n")
		}

		// Workdir field
		workdirLabel := "Workdir: "
		if m.profileEditState.GetField() == 1 {
			workdirLabel = styleSelected.Render(workdirLabel)
			content.WriteString(workdirLabel + addFieldCursorScrolling(m.profileEditState.GetWorkdir(), m.profileEditState.GetWorkdirPos(), fieldWidth) + "\n")
		} else {
			content.WriteString(workdirLabel + truncateField(m.profileEditState.GetWorkdir(), fieldWidth) + "\n")
		}

		// Editor field
		editorLabel := "Editor:  "
		if m.profileEditState.GetField() == 2 {
			editorLabel = styleSelected.Render(editorLabel)
			content.WriteString(editorLabel + addFieldCursorScrolling(m.profileEditState.GetEditor(), m.profileEditState.GetEditorPos(), fieldWidth) + "\n")
		} else {
			content.WriteString(editorLabel + truncateField(m.profileEditState.GetEditor(), fieldWidth) + "\n")
		}

		// Output field
		outputLabel := "Output:  "
		if m.profileEditState.GetField() == 3 {
			outputLabel = styleSelected.Render(outputLabel)
			content.WriteString(outputLabel + addFieldCursorScrolling(m.profileEditState.GetOutput(), m.profileEditState.GetOutputPos(), fieldWidth) + "\n")
		} else {
			content.WriteString(outputLabel + truncateField(m.profileEditState.GetOutput(), fieldWidth) + "\n")
		}

		// History field (toggle)
		historyLabel := "History:   "
		historyValue := "default"
		if m.profileEditState.GetHistoryEnabled() != nil {
			if *m.profileEditState.GetHistoryEnabled() {
				historyValue = "enabled"
			} else {
				historyValue = "disabled"
			}
		}
		if m.profileEditState.GetField() == 4 {
			historyLabel = styleSelected.Render(historyLabel)
			content.WriteString(historyLabel + styleSelected.Render(historyValue) + " [SPACE to toggle]\n")
		} else {
			content.WriteString(historyLabel + historyValue + "\n")
		}

		// Analytics field (toggle)
		analyticsLabel := "Analytics: "
		analyticsValue := "disabled"
		if m.profileEditState.GetAnalyticsEnabled() != nil {
			if *m.profileEditState.GetAnalyticsEnabled() {
				analyticsValue = "enabled"
			} else {
				analyticsValue = "disabled"
			}
		}
		if m.profileEditState.GetField() == 5 {
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
