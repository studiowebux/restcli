package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// handleRenameKeys handles keyboard input in rename mode
func (m *Model) handleRenameKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.renameInput = ""
		m.errorMsg = ""

	case "enter":
		if m.renameInput == "" {
			m.errorMsg = "Filename cannot be empty"
			return nil
		}

		if len(m.files) == 0 {
			m.errorMsg = "No file selected"
			return nil
		}

		// Get absolute path of old file
		oldPath := m.files[m.fileIndex].Path
		if !filepath.IsAbs(oldPath) {
			var err error
			oldPath, err = filepath.Abs(oldPath)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Failed to get absolute path: %v", err)
				return nil
			}
		}

		dir := filepath.Dir(oldPath)

		// Ensure extension is preserved if not provided
		newName := m.renameInput
		if filepath.Ext(newName) == "" {
			newName += filepath.Ext(oldPath)
		}

		// Build absolute new path
		newPath := filepath.Join(dir, newName)

		// Check if file already exists
		if _, err := os.Stat(newPath); err == nil {
			m.errorMsg = fmt.Sprintf("File '%s' already exists", newName)
			return nil
		}

		// Create directories if the new path includes subdirectories
		newDir := filepath.Dir(newPath)
		if err := os.MkdirAll(newDir, 0755); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to create directory: %v", err)
			return nil
		}

		// Rename the file
		if err := os.Rename(oldPath, newPath); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to rename file: %v", err)
			return nil
		}

		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Renamed to: %s", newName)
		m.renameInput = ""

		// Refresh file list
		return m.refreshFiles()

	default:
		// Clear error when user starts typing
		m.errorMsg = ""

		// Handle common text input operations (paste, clear, backspace)
		if _, shouldContinue := handleTextInput(&m.renameInput, msg); shouldContinue {
			return nil
		}
		// Append character to rename input
		if len(msg.String()) == 1 {
			m.renameInput += msg.String()
		}
	}

	return nil
}

// renderRenameModal renders the file rename modal
func (m *Model) renderRenameModal() string {
	if len(m.files) == 0 {
		return m.renderModal("Rename", "No file selected\n\nPress ESC to close", 50, 10)
	}

	currentName := m.files[m.fileIndex].Name
	content := fmt.Sprintf("Current: %s\n\nNew name: %s",
		currentName, addCursor(m.renameInput))

	// Show error if present (wrapped to modal width)
	if m.errorMsg != "" {
		wrappedError := wrapText(m.errorMsg, 54) // Modal width (60) minus padding
		content += "\n\n" + styleError.Render(wrappedError)
	}

	content += "\n\nEnter new name, then press Enter to rename, ESC to cancel"

	return m.renderModal("Rename File", content, 60, 15)
}

// renderEditorConfigModal renders the editor configuration modal
func (m *Model) renderEditorConfigModal() string {
	// Show cursor in input
	inputField := m.inputValue[:m.inputCursor] + "█" + m.inputValue[m.inputCursor:]

	content := fmt.Sprintf("Editor: %s\n\nExamples: vim, nvim, nano, code, emacs", inputField)
	footer := "[Enter] save [ESC] cancel"

	return m.renderModalWithFooter("Configure Editor", content, footer, 50, 10)
}

// renderDeleteModal renders the delete file confirmation modal
func (m *Model) renderDeleteModal() string {
	if len(m.files) == 0 {
		return m.renderModal("Delete", "No file selected\n\nPress ESC to close", 50, 10)
	}

	fileName := m.files[m.fileIndex].Name
	content := fmt.Sprintf("Are you sure you want to delete:\n\n  %s\n\nThis action cannot be undone.", fileName)
	footer := "[y]es [n]o"

	return m.renderModalWithFooter("Delete File", content, footer, 60, 12)
}
