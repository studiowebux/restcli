package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/keybinds"
)

var fileTypes = []string{"http", "json", "yaml", "jsonc"}

// handleCreateFileKeys handles keyboard input in create file mode
func (m *Model) handleCreateFileKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle tab specially (cycle file types)
	if msg.String() == "tab" {
		// Cycle through file types
		m.createFileType = (m.createFileType + 1) % len(fileTypes)
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeNormal
			m.createFileInput = ""
			m.errorMsg = ""
			return nil

		case keybinds.ActionTextSubmit:
			if m.createFileInput == "" {
				m.errorMsg = "Filename cannot be empty"
				return nil
			}

			profile := m.sessionMgr.GetActiveProfile()
			if profile == nil {
				m.errorMsg = "No active profile"
				return nil
			}

			// Get working directory from profile
			workdir := profile.Workdir
			if workdir == "" {
				m.errorMsg = "Profile workdir not set"
				return nil
			}

			// Expand tilde in workdir
			if strings.HasPrefix(workdir, "~/") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					m.errorMsg = fmt.Sprintf("Failed to get home directory: %v", err)
					return nil
				}
				workdir = filepath.Join(homeDir, workdir[2:])
			}

			// Ensure extension is added
			filename := m.createFileInput
			ext := "." + fileTypes[m.createFileType]
			if !strings.HasSuffix(filename, ext) {
				filename += ext
			}

			// Expand tilde to home directory
			if strings.HasPrefix(filename, "~/") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					m.errorMsg = fmt.Sprintf("Failed to get home directory: %v", err)
					return nil
				}
				filename = filepath.Join(homeDir, filename[2:])
			}

			// Build absolute path
			var fullPath string
			if filepath.IsAbs(filename) {
				fullPath = filename
			} else {
				fullPath = filepath.Join(workdir, filename)
			}

			// Check if file already exists
			if _, err := os.Stat(fullPath); err == nil {
				m.errorMsg = fmt.Sprintf("File '%s' already exists", filename)
				return nil
			}

			// Create directories if needed
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, config.DirPermissions); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to create directory: %v", err)
				return nil
			}

			// Create the file with basic template based on type
			content := getFileTemplate(fileTypes[m.createFileType])
			if err := os.WriteFile(fullPath, []byte(content), config.FilePermissions); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to create file: %v", err)
				return nil
			}

			m.mode = ModeNormal
			m.statusMsg = fmt.Sprintf("Created: %s", filename)
			m.createFileInput = ""

			// Refresh file list
			return m.refreshFiles()
		}
	}

	// Clear error when user starts typing
	m.errorMsg = ""

	// Handle text input with cursor support (arrow keys, etc.)
	if _, shouldContinue := handleTextInputWithCursor(&m.createFileInput, &m.createFileCursor, msg); shouldContinue {
		return nil
	}
	// Insert character at cursor position
	if len(msg.String()) == 1 {
		m.createFileInput = m.createFileInput[:m.createFileCursor] + msg.String() + m.createFileInput[m.createFileCursor:]
		m.createFileCursor++
	}

	return nil
}

// renderCreateFileModal renders the create file modal
func (m *Model) renderCreateFileModal() string {
	profile := m.sessionMgr.GetActiveProfile()
	workdir := "not set"
	if profile != nil && profile.Workdir != "" {
		workdir = profile.Workdir
		// Expand tilde for display
		if strings.HasPrefix(workdir, "~/") {
			if homeDir, err := os.UserHomeDir(); err == nil {
				workdir = filepath.Join(homeDir, workdir[2:])
			}
		}
	}

	// Build file type selector
	fileTypeDisplay := ""
	for i, ft := range fileTypes {
		if i == m.createFileType {
			fileTypeDisplay += fmt.Sprintf("[%s] ", ft)
		} else {
			fileTypeDisplay += fmt.Sprintf(" %s  ", ft)
		}
	}

	// Show cursor at correct position
	inputWithCursor := m.createFileInput[:m.createFileCursor] + "█" + m.createFileInput[m.createFileCursor:]

	// Wrap working directory path if it's too long
	wrappedWorkdir := wrapText(workdir, 64)

	content := fmt.Sprintf("Working directory:\n%s\n\nFilename: %s\n\nFile type: %s\n(Press TAB to cycle)",
		wrappedWorkdir, inputWithCursor, fileTypeDisplay)

	// Show error if present (wrapped to modal width)
	if m.errorMsg != "" {
		wrappedError := wrapText(m.errorMsg, 64)
		content += "\n\n" + styleError.Render(wrappedError)
	}

	// Wrap instruction text to modal width
	instruction := wrapText("Enter filename (with optional path), then press Enter to create, ESC to cancel", 64)
	content += "\n\n" + instruction

	return m.renderModal("Create New File", content, 70, 18)
}

// getFileTemplate returns a basic template for each file type
func getFileTemplate(fileType string) string {
	switch fileType {
	case "http":
		return `### New Request
GET https://example.com
`
	case "json":
		return `[
  {
    "name": "New Request",
    "method": "GET",
    "url": "https://example.com"
  }
]
`
	case "yaml":
		return `- name: New Request
  method: GET
  url: https://example.com
`
	case "jsonc":
		return `// JSON with Comments
[
  {
    "name": "New Request",
    "method": "GET",
    "url": "https://example.com"
  }
]
`
	default:
		return ""
	}
}
