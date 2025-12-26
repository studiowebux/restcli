package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/keybinds"
	"github.com/studiowebux/restcli/internal/types"
)

// renderVariableEditor renders the variable editor in its various modes
func (m *Model) renderVariableEditor() string {
	profile := m.sessionMgr.GetActiveProfile()

	var content strings.Builder

	var footer string

	switch m.mode {
	case ModeVariableList:
		content.WriteString("Profile Variables:\n")

		if len(profile.Variables) == 0 {
			content.WriteString("  (none)\n")
		} else {
			// Convert to sorted slice for consistent ordering
			type varEntry struct {
				name  string
				value types.VariableValue
			}
			vars := make([]varEntry, 0, len(profile.Variables))
			for name, value := range profile.Variables {
				vars = append(vars, varEntry{name, value})
			}
			// Sort by name
			for i := 0; i < len(vars); i++ {
				for j := i + 1; j < len(vars); j++ {
					if vars[i].name > vars[j].name {
						vars[i], vars[j] = vars[j], vars[i]
					}
				}
			}

			for i, v := range vars {
				line := fmt.Sprintf("  %s = %s", v.name, truncate(v.value.GetValue(), 50))
				if v.value.IsMultiValue() {
					line += " [multi]"
				}
				if v.value.Interactive {
					line += " [interactive]"
				}
				if i == m.varEditIndex {
					line = styleSelected.Render(line)
				}
				content.WriteString(line + "\n")
			}
		}

		footer = "[a]dd [e]dit [d]elete [o]ptions [m]anage [i]nteractive [ESC]"

	case ModeVariableAdd:
		content.WriteString("Add Variable\n\n")
		nameField := m.varEditName
		valueField := m.varEditValue
		if m.varEditCursor == 0 {
			// Insert cursor at position in name field
			nameField = m.varEditName[:m.varEditNamePos] + "█" + m.varEditName[m.varEditNamePos:]
		} else {
			// Insert cursor at position in value field
			valueField = m.varEditValue[:m.varEditValuePos] + "█" + m.varEditValue[m.varEditValuePos:]
		}
		content.WriteString("Name:  " + nameField + "\n")
		content.WriteString("Value: " + valueField + "\n")
		footer = "[TAB] switch fields [Enter] save [ESC] cancel"

	case ModeVariableEdit:
		content.WriteString("Edit Variable\n\n")
		nameField := m.varEditName
		valueField := m.varEditValue
		if m.varEditCursor == 0 {
			// Insert cursor at position in name field
			nameField = m.varEditName[:m.varEditNamePos] + "█" + m.varEditName[m.varEditNamePos:]
		} else {
			// Insert cursor at position in value field
			valueField = m.varEditValue[:m.varEditValuePos] + "█" + m.varEditValue[m.varEditValuePos:]
		}
		content.WriteString("Name:  " + nameField + "\n")
		content.WriteString("Value: " + valueField + "\n")
		footer = "[TAB] switch fields [Enter] save [ESC] cancel"

	case ModeVariableDelete:
		content.WriteString("Delete Variable\n\n")
		content.WriteString(fmt.Sprintf("Are you sure you want to delete '%s'?", m.varEditName))
		footer = "[y]es [n]o"

	case ModeVariableManage:
		content.WriteString("Manage Multi-Value Variable\n\n")
		varValue, exists := profile.Variables[m.varEditName]
		if !exists || !varValue.IsMultiValue() {
			content.WriteString("Not a multi-value variable")
			footer = "[ESC] back"
		} else {
			content.WriteString(fmt.Sprintf("Variable: %s\n\n", m.varEditName))
			content.WriteString("Options:\n")
			for i, opt := range varValue.MultiValue.Options {
				line := fmt.Sprintf("  %s", opt)
				if i == varValue.MultiValue.Active {
					line += " [ACTIVE]"
				}
				// Show aliases for this option
				if varValue.MultiValue.Aliases != nil {
					var aliases []string
					for alias, idx := range varValue.MultiValue.Aliases {
						if idx == i {
							aliases = append(aliases, alias)
						}
					}
					if len(aliases) > 0 {
						line += fmt.Sprintf(" (%s)", strings.Join(aliases, ", "))
					}
				}
				if i == m.varOptionIndex {
					line = styleSelected.Render(line)
				}
				content.WriteString(line + "\n")
			}
			// Show edit input if editing
			if m.varEditCursor == 1 {
				editField := m.varEditValue[:m.varEditValuePos] + "█" + m.varEditValue[m.varEditValuePos:]
				content.WriteString(fmt.Sprintf("\nValue: %s", editField))
				footer = "[Enter] save [ESC] cancel"
			} else {
				footer = "↑/↓ [s]et [e]dit [a]dd [d]el [l]alias [L]del alias [ESC]"
			}
		}

	case ModeVariableAlias:
		content.WriteString("Set Alias\n\n")
		varValue, exists := profile.Variables[m.varEditName]
		if exists && varValue.IsMultiValue() && m.varAliasTargetIdx < len(varValue.MultiValue.Options) {
			content.WriteString(fmt.Sprintf("Option: %s\n\n", truncate(varValue.MultiValue.Options[m.varAliasTargetIdx], 50)))
		}
		aliasField := m.varAliasInput + "█"
		content.WriteString("Alias: " + aliasField + "\n")
		content.WriteString("\nEnter a short name for this option (e.g., 'u1', 'dev')")
		footer = "[Enter] save [ESC] cancel"

	case ModeVariableOptions:
		content.WriteString("Create Multi-Value Variable\n\n")
		nameField := m.varEditName
		valueField := m.varEditValue
		if m.varEditCursor == 0 {
			// Insert cursor at position in name field
			nameField = m.varEditName[:m.varEditNamePos] + "█" + m.varEditName[m.varEditNamePos:]
		} else {
			// Insert cursor at position in value field
			valueField = m.varEditValue[:m.varEditValuePos] + "█" + m.varEditValue[m.varEditValuePos:]
		}
		content.WriteString("Name:  " + nameField + "\n\n")

		// Wrap options field to prevent overflow
		// Modal width is typically 70-80, minus padding (4) = 66-76
		// "Options: " prefix takes ~30 chars, leaving ~40-45 chars for content
		wrappedValue := wrapText(valueField, 55)
		content.WriteString("Options (comma-separated):\n" + wrappedValue + "\n")
		footer = "[TAB] switch fields [Enter] save [ESC] cancel"
	}

	// Calculate selected line for auto-scroll
	selectedLine := -1
	switch m.mode {
	case ModeVariableList:
		// Line 0 is "Profile Variables:", selected item is at line 1 + index
		selectedLine = 1 + m.varEditIndex
	case ModeVariableManage:
		// Only scroll when not editing (editing input is at bottom)
		if m.varEditCursor != 1 {
			// Content lines:
			// 0: "Manage Multi-Value Variable"
			// 1: (empty)
			// 2: "Variable: x"
			// 3: (empty)
			// 4: "Options:"
			// 5+: options
			selectedLine = 5 + m.varOptionIndex
			// Clamp to reasonable value
			if selectedLine < 0 {
				selectedLine = 0
			}
		}
	}

	return m.renderModalWithFooterAndScroll("Variables", content.String(), footer, 70, 25, selectedLine)
}

// getSortedVariableNames returns sorted variable names
func getSortedVariableNames(vars map[string]types.VariableValue) []string {
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	// Simple bubble sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}

// handleVariableEditorKeys handles keyboard input in variable editor modes
func (m *Model) handleVariableEditorKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	switch m.mode {
	case ModeVariableList:
		sortedNames := getSortedVariableNames(profile.Variables)

		// Handle special keys not in registry
		switch msg.String() {
		case "o":
			m.mode = ModeVariableOptions
			m.varEditName = ""
			m.varEditValue = ""
			m.varEditCursor = 0
			m.varEditNamePos = 0
			m.varEditValuePos = 0
			return nil
		case "i":
			if len(sortedNames) > 0 && m.varEditIndex < len(sortedNames) {
				name := sortedNames[m.varEditIndex]
				varValue := profile.Variables[name]
				varValue.Interactive = !varValue.Interactive
				profile.Variables[name] = varValue
				m.sessionMgr.SaveProfiles()
				if varValue.Interactive {
					m.statusMsg = fmt.Sprintf("Variable '%s' is now interactive", name)
				} else {
					m.statusMsg = fmt.Sprintf("Variable '%s' is no longer interactive", name)
				}
			}
			return nil
		}

		action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextVariableList, msg.String())
		if partial {
			return nil
		}

		if !ok {
			m.gPressed = false
			return nil
		}

		switch action {
		case keybinds.ActionCloseModal:
			m.mode = ModeNormal

		case keybinds.ActionNavigateUp:
			if m.varEditIndex > 0 {
				m.varEditIndex--
			}

		case keybinds.ActionNavigateDown:
			if m.varEditIndex < len(sortedNames)-1 {
				m.varEditIndex++
			}

		case keybinds.ActionPageUp:
			pageSize := m.modalView.Height
			if pageSize < 1 {
				pageSize = 10
			}
			m.varEditIndex -= pageSize
			if m.varEditIndex < 0 {
				m.varEditIndex = 0
			}

		case keybinds.ActionPageDown:
			pageSize := m.modalView.Height
			if pageSize < 1 {
				pageSize = 10
			}
			m.varEditIndex += pageSize
			if m.varEditIndex >= len(sortedNames) {
				m.varEditIndex = len(sortedNames) - 1
			}
			if m.varEditIndex < 0 {
				m.varEditIndex = 0
			}

		case keybinds.ActionGoToTop:
			m.varEditIndex = 0

		case keybinds.ActionGoToBottom:
			if len(sortedNames) > 0 {
				m.varEditIndex = len(sortedNames) - 1
			}

		case keybinds.ActionVarAdd:
			m.mode = ModeVariableAdd
			m.varEditName = ""
			m.varEditValue = ""
			m.varEditCursor = 0
			m.varEditNamePos = 0
			m.varEditValuePos = 0

		case keybinds.ActionVarEdit:
			if len(sortedNames) > 0 && m.varEditIndex < len(sortedNames) {
				name := sortedNames[m.varEditIndex]
				value := profile.Variables[name]
				m.mode = ModeVariableEdit
				m.varEditName = name
				if value.IsMultiValue() {
					m.varEditValue = "[multi-value - use 'm' to manage options]"
					m.varEditCursor = 0
					m.varEditNamePos = len(name)
					m.varEditValuePos = 0
				} else {
					m.varEditValue = value.GetValue()
					m.varEditCursor = 0
					m.varEditNamePos = len(name)
					m.varEditValuePos = len(value.GetValue())
				}
			}

		case keybinds.ActionVarDelete:
			if len(sortedNames) > 0 && m.varEditIndex < len(sortedNames) {
				m.mode = ModeVariableDelete
				m.varEditName = sortedNames[m.varEditIndex]
			}

		case keybinds.ActionVarManage:
			if len(sortedNames) > 0 && m.varEditIndex < len(sortedNames) {
				name := sortedNames[m.varEditIndex]
				value := profile.Variables[name]
				if value.IsMultiValue() {
					m.mode = ModeVariableManage
					m.varEditName = name
					m.varOptionIndex = 0
					m.varEditCursor = 0
					m.modalView.SetYOffset(0)
				} else {
					m.statusMsg = "Not a multi-value variable"
				}
			}
		}

		m.gPressed = false

	case ModeVariableAdd, ModeVariableEdit:
		return m.handleVariableInputKeys(msg)

	case ModeVariableDelete:
		action, ok := m.keybinds.Match(keybinds.ContextConfirm, msg.String())
		if !ok {
			return nil
		}

		switch action {
		case keybinds.ActionConfirm:
			delete(profile.Variables, m.varEditName)
			m.sessionMgr.SaveProfiles()
			m.mode = ModeVariableList
			m.statusMsg = fmt.Sprintf("Deleted variable: %s", m.varEditName)

		case keybinds.ActionCancel:
			m.mode = ModeVariableList
		}

	case ModeVariableManage:
		return m.handleVariableManageKeys(msg)

	case ModeVariableAlias:
		return m.handleVariableAliasKeys(msg)

	case ModeVariableOptions:
		return m.handleVariableOptionsKeys(msg)
	}

	return nil
}

// handleVariableInputKeys handles text input for add/edit variable
func (m *Model) handleVariableInputKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	// Handle tab specially (field switching)
	if msg.String() == "tab" {
		if m.varEditValue != "[multi-value - use 'm' to manage options]" {
			m.varEditCursor = (m.varEditCursor + 1) % 2
		}
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextVariableEdit, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeVariableList
			return nil

		case keybinds.ActionTextSubmit:
			if m.varEditName == "" {
				m.errorMsg = "Variable name cannot be empty"
				return nil
			}

			if profile.Variables == nil {
				profile.Variables = make(map[string]types.VariableValue)
			}

			if m.varEditValue == "[multi-value - use 'm' to manage options]" {
				sortedNames := getSortedVariableNames(profile.Variables)
				if m.varEditIndex < len(sortedNames) {
					oldName := sortedNames[m.varEditIndex]
					if oldName != m.varEditName {
						profile.Variables[m.varEditName] = profile.Variables[oldName]
						delete(profile.Variables, oldName)
					}
				}
			} else {
				var varValue types.VariableValue
				if m.mode == ModeVariableEdit {
					if existingVar, exists := profile.Variables[m.varEditName]; exists {
						varValue.Interactive = existingVar.Interactive
					}
				}
				varValue.SetValue(m.varEditValue)
				profile.Variables[m.varEditName] = varValue
			}

			m.sessionMgr.SaveProfiles()
			m.mode = ModeVariableList
			m.statusMsg = fmt.Sprintf("Saved variable: %s", m.varEditName)
			return nil

		case keybinds.ActionNavigateUp:
			if m.varEditCursor == 0 {
				m.varEditCursor = 1
			} else {
				m.varEditCursor = 0
			}
			return nil
		}
	}

	// Handle text input with cursor support
	if m.varEditCursor == 0 {
		if _, shouldContinue := handleTextInputWithCursor(&m.varEditName, &m.varEditNamePos, msg); shouldContinue {
			return nil
		}
		if len(msg.String()) == 1 {
			m.varEditName = m.varEditName[:m.varEditNamePos] + msg.String() + m.varEditName[m.varEditNamePos:]
			m.varEditNamePos++
		}
	} else if m.varEditValue != "[multi-value - use 'm' to manage options]" {
		if _, shouldContinue := handleTextInputWithCursor(&m.varEditValue, &m.varEditValuePos, msg); shouldContinue {
			return nil
		}
		if len(msg.String()) == 1 {
			m.varEditValue = m.varEditValue[:m.varEditValuePos] + msg.String() + m.varEditValue[m.varEditValuePos:]
			m.varEditValuePos++
		}
	}

	return nil
}

// handleVariableManageKeys handles multi-value variable management
func (m *Model) handleVariableManageKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()
	varValue, exists := profile.Variables[m.varEditName]
	if !exists || !varValue.IsMultiValue() {
		m.mode = ModeVariableList
		return nil
	}

	// If in edit mode, handle text input first
	if m.varEditCursor == 1 {
		switch msg.String() {
		case "esc":
			// Cancel edit
			m.varEditValue = ""
			m.varEditCursor = 0
			m.modalView.SetYOffset(0) // Reset scroll
		case "enter":
			// Save edited option
			if m.varEditValue != "" {
				if m.varOptionIndex < len(varValue.MultiValue.Options) {
					// Update existing option
					varValue.MultiValue.Options[m.varOptionIndex] = m.varEditValue
				} else {
					// Add new option
					varValue.MultiValue.Options = append(varValue.MultiValue.Options, m.varEditValue)
					// Point to the newly added option
					m.varOptionIndex = len(varValue.MultiValue.Options) - 1
				}
				profile.Variables[m.varEditName] = varValue
				m.sessionMgr.SaveProfiles()
				m.statusMsg = "Option saved"
				m.varEditValue = ""
				m.varEditCursor = 0
				m.modalView.SetYOffset(0) // Reset scroll
			}
		default:
			// Handle text input
			if _, shouldContinue := handleTextInputWithCursor(&m.varEditValue, &m.varEditValuePos, msg); shouldContinue {
				return nil
			}
			// Insert character at cursor position
			if len(msg.String()) == 1 {
				m.varEditValue = m.varEditValue[:m.varEditValuePos] + msg.String() + m.varEditValue[m.varEditValuePos:]
				m.varEditValuePos++
			}
		}
		return nil
	}

	// Normal mode (not editing) - handle special keys before registry
	switch msg.String() {
	case "s":
		// Set as active
		varValue.MultiValue.Active = m.varOptionIndex
		profile.Variables[m.varEditName] = varValue
		m.sessionMgr.SaveProfiles()
		m.statusMsg = "Active option updated"
		return nil

	case "d":
		// Delete option
		if len(varValue.MultiValue.Options) > 1 {
			options := varValue.MultiValue.Options
			varValue.MultiValue.Options = append(options[:m.varOptionIndex], options[m.varOptionIndex+1:]...)
			if varValue.MultiValue.Active >= len(varValue.MultiValue.Options) {
				varValue.MultiValue.Active = len(varValue.MultiValue.Options) - 1
			}
			profile.Variables[m.varEditName] = varValue
			m.sessionMgr.SaveProfiles()
			if m.varOptionIndex >= len(varValue.MultiValue.Options) {
				m.varOptionIndex = len(varValue.MultiValue.Options) - 1
			}
			m.statusMsg = "Option deleted"
		}
		return nil

	case "e":
		// Edit current option
		if m.varOptionIndex < len(varValue.MultiValue.Options) {
			m.varEditValue = varValue.MultiValue.Options[m.varOptionIndex]
			m.varEditValuePos = len(m.varEditValue)
			m.varEditCursor = 1 // Focus on value field
		}
		return nil

	case "a":
		// Add option - set index past end to indicate new item
		m.varOptionIndex = len(varValue.MultiValue.Options)
		m.varEditValue = ""
		m.varEditValuePos = 0
		m.varEditCursor = 1 // Focus on value field
		return nil

	case "l":
		// Set alias for current option
		if len(varValue.MultiValue.Options) == 0 {
			m.statusMsg = "No options to alias"
		} else if m.varOptionIndex < len(varValue.MultiValue.Options) {
			m.mode = ModeVariableAlias
			m.varAliasTargetIdx = m.varOptionIndex
			m.varAliasInput = ""
		} else {
			m.statusMsg = "Invalid option index"
		}
		return nil

	case "L":
		// Delete all aliases from current option
		if varValue.MultiValue.Aliases != nil {
			// Find and delete aliases pointing to this option
			deleted := false
			for alias, idx := range varValue.MultiValue.Aliases {
				if idx == m.varOptionIndex {
					delete(varValue.MultiValue.Aliases, alias)
					deleted = true
				}
			}
			if deleted {
				profile.Variables[m.varEditName] = varValue
				m.sessionMgr.SaveProfiles()
				m.statusMsg = "Aliases deleted"
			} else {
				m.statusMsg = "No aliases to delete"
			}
		} else {
			m.statusMsg = "No aliases to delete"
		}
		return nil
	}

	// Use registry for navigation keys
	action, ok := m.keybinds.Match(keybinds.ContextVariableList, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeVariableList

	case keybinds.ActionNavigateUp:
		if m.varOptionIndex > 0 {
			m.varOptionIndex--
		}

	case keybinds.ActionNavigateDown:
		if m.varOptionIndex < len(varValue.MultiValue.Options)-1 {
			m.varOptionIndex++
		}

	case keybinds.ActionPageUp:
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.varOptionIndex -= pageSize
		if m.varOptionIndex < 0 {
			m.varOptionIndex = 0
		}

	case keybinds.ActionPageDown:
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.varOptionIndex += pageSize
		if m.varOptionIndex >= len(varValue.MultiValue.Options) {
			m.varOptionIndex = len(varValue.MultiValue.Options) - 1
		}
		if m.varOptionIndex < 0 {
			m.varOptionIndex = 0
		}

	case keybinds.ActionGoToTop:
		m.varOptionIndex = 0

	case keybinds.ActionGoToBottom:
		if len(varValue.MultiValue.Options) > 0 {
			m.varOptionIndex = len(varValue.MultiValue.Options) - 1
		}
	}

	return nil
}

// handleVariableOptionsKeys handles creating multi-value variables
func (m *Model) handleVariableOptionsKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	// Handle tab specially (field switching)
	if msg.String() == "tab" {
		m.varEditCursor = (m.varEditCursor + 1) % 2
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextVariableEdit, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeVariableList
			return nil

		case keybinds.ActionTextSubmit:
			if m.varEditName == "" {
				m.errorMsg = "Variable name cannot be empty"
				return nil
			}
			if m.varEditValue == "" {
				m.errorMsg = "Options cannot be empty"
				return nil
			}

			// Parse comma-separated options
			options := strings.Split(m.varEditValue, ",")
			for i := range options {
				options[i] = strings.TrimSpace(options[i])
			}

			// Filter out empty options
			var cleanOptions []string
			for _, opt := range options {
				if opt != "" {
					cleanOptions = append(cleanOptions, opt)
				}
			}

			if len(cleanOptions) == 0 {
				m.errorMsg = "No valid options provided"
				return nil
			}

			// Create multi-value variable
			if profile.Variables == nil {
				profile.Variables = make(map[string]types.VariableValue)
			}

			varValue := types.VariableValue{
				MultiValue: &types.MultiValueVariable{
					Options: cleanOptions,
					Active:  0,
				},
			}
			profile.Variables[m.varEditName] = varValue

			m.sessionMgr.SaveProfiles()
			m.mode = ModeVariableList
			m.statusMsg = fmt.Sprintf("Created multi-value variable: %s", m.varEditName)
			return nil
		}
	}

	// Handle text input with cursor support
	if m.varEditCursor == 0 {
		// Editing name field
		if _, shouldContinue := handleTextInputWithCursor(&m.varEditName, &m.varEditNamePos, msg); shouldContinue {
			return nil
		}
		// Insert character at cursor position
		if len(msg.String()) == 1 {
			m.varEditName = m.varEditName[:m.varEditNamePos] + msg.String() + m.varEditName[m.varEditNamePos:]
			m.varEditNamePos++
		}
	} else {
		// Editing value field
		if _, shouldContinue := handleTextInputWithCursor(&m.varEditValue, &m.varEditValuePos, msg); shouldContinue {
			return nil
		}
		// Insert character at cursor position
		if len(msg.String()) == 1 {
			m.varEditValue = m.varEditValue[:m.varEditValuePos] + msg.String() + m.varEditValue[m.varEditValuePos:]
			m.varEditValuePos++
		}
	}

	return nil
}

// handleVariableAliasKeys handles alias input for multi-value options
func (m *Model) handleVariableAliasKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()
	varValue, exists := profile.Variables[m.varEditName]
	if !exists || !varValue.IsMultiValue() {
		m.mode = ModeVariableManage
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeVariableManage
			m.varAliasInput = ""
			return nil

		case keybinds.ActionTextSubmit:
			if m.varAliasInput == "" {
				m.errorMsg = "Alias cannot be empty"
				return nil
			}

			// Initialize aliases map if needed
			if varValue.MultiValue.Aliases == nil {
				varValue.MultiValue.Aliases = make(map[string]int)
			}

			// Set the alias
			varValue.MultiValue.Aliases[m.varAliasInput] = m.varAliasTargetIdx
			profile.Variables[m.varEditName] = varValue
			m.sessionMgr.SaveProfiles()

			m.statusMsg = fmt.Sprintf("Alias '%s' set", m.varAliasInput)
			m.mode = ModeVariableManage
			m.varAliasInput = ""
			return nil

		case keybinds.ActionTextBackspace:
			if len(m.varAliasInput) > 0 {
				m.varAliasInput = m.varAliasInput[:len(m.varAliasInput)-1]
			}
			return nil
		}
	}

	// Only accept alphanumeric characters and common separators for aliases
	if len(msg.String()) == 1 {
		ch := msg.String()[0]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			m.varAliasInput += msg.String()
		}
	}

	return nil
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
