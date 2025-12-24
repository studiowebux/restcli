package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/keybinds"
)

// handleMRUKeys handles keyboard input in MRU mode
func (m *Model) handleMRUKeys(msg tea.KeyMsg) tea.Cmd {
	recentFiles := m.sessionMgr.GetRecentFiles()

	// Handle quick select and enter before registry (special keys)
	switch msg.String() {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Quick select by number (1-9)
		num := int(msg.String()[0] - '0') // Convert '1'-'9' to 1-9
		index := num - 1                   // Convert to 0-based index
		if index < len(recentFiles) {
			m.mruIndex = index
			// Immediately select the file (simulate enter key)
			msg.Type = tea.KeyEnter
			return m.handleMRUKeys(tea.KeyMsg{Type: tea.KeyEnter})
		}

	case "0":
		// Quick select 10th item (0 = 10)
		index := 9
		if index < len(recentFiles) {
			m.mruIndex = index
			// Immediately select the file (simulate enter key)
			msg.Type = tea.KeyEnter
			return m.handleMRUKeys(tea.KeyMsg{Type: tea.KeyEnter})
		}

	case "enter":
		if len(recentFiles) == 0 {
			m.errorMsg = "No recent files"
			return nil
		}

		selectedFile := recentFiles[m.mruIndex]

		// Check if file exists
		if _, err := os.Stat(selectedFile); os.IsNotExist(err) {
			m.errorMsg = fmt.Sprintf("File not found: %s", selectedFile)
			return nil
		}

		// Find the file in the current file list
		found := false
		files := m.fileExplorer.GetFiles()
		for i, f := range files {
			if f.Path == selectedFile {
				// Navigate to this file
				currentIdx := m.fileExplorer.GetCurrentIndex()
				delta := i - currentIdx
				pageSize := m.getFileListHeight()
				m.fileExplorer.Navigate(delta, pageSize)
				m.loadRequestsFromCurrentFile()
				found = true
				break
			}
		}

		if !found {
			m.errorMsg = fmt.Sprintf("File not in current directory: %s", selectedFile)
			return nil
		}

		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Opened: %s", filepath.Base(selectedFile))
		m.errorMsg = ""
	}

	// Use registry for navigation
	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextModal, msg.String())
	if partial {
		return nil
	}
	if !ok {
		// Reset 'g' press on any other key
		m.gPressed = false
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal
		m.errorMsg = ""

	case keybinds.ActionNavigateDown:
		if len(recentFiles) > 0 {
			m.mruIndex = (m.mruIndex + 1) % len(recentFiles)
		}

	case keybinds.ActionNavigateUp:
		if len(recentFiles) > 0 {
			m.mruIndex = (m.mruIndex - 1 + len(recentFiles)) % len(recentFiles)
		}

	case keybinds.ActionGoToTop:
		// Go to top (triggered by gg)
		m.mruIndex = 0

	case keybinds.ActionGoToBottom:
		// Go to bottom
		if len(recentFiles) > 0 {
			m.mruIndex = len(recentFiles) - 1
		}
	}

	m.gPressed = false
	return nil
}

// renderMRUModal renders the MRU (Most Recently Used) files modal
func (m *Model) renderMRUModal() string {
	recentFiles := m.sessionMgr.GetRecentFiles()

	if len(recentFiles) == 0 {
		content := "No recent files\n\nPress ESC to close"
		return m.renderModal("Recent Files", content, 60, 10)
	}

	// Filter to only show files that exist
	existingFiles := []string{}
	for _, f := range recentFiles {
		if _, err := os.Stat(f); err == nil {
			existingFiles = append(existingFiles, f)
		}
	}

	if len(existingFiles) == 0 {
		content := "No recent files found (files may have been deleted)\n\nPress ESC to close"
		return m.renderModal("Recent Files", content, 60, 10)
	}

	// Build file list with selection
	var content strings.Builder
	profile := m.sessionMgr.GetActiveProfile()
	workdir := ""
	if profile != nil {
		workdir = profile.Workdir
	}

	// Render all files - let the modal viewport handle scrolling
	for i, filePath := range existingFiles {
		displayPath := filePath

		// Show relative path if in workdir
		if workdir != "" {
			if rel, err := filepath.Rel(workdir, filePath); err == nil {
				if !strings.HasPrefix(rel, "..") {
					displayPath = rel
				}
			}
		}

		// Add number prefix for first 10 items (1-9, 0 for 10th)
		var prefix string
		if i < 9 {
			prefix = fmt.Sprintf("%d. ", i+1)
		} else if i == 9 {
			prefix = "0. "
		} else {
			prefix = "   "
		}

		if i == m.mruIndex {
			content.WriteString(styleSelected.Render(fmt.Sprintf("%s%s", prefix, displayPath)) + "\n")
		} else {
			content.WriteString(fmt.Sprintf("  %s%s\n", prefix, displayPath))
		}
	}

	// Show error if present
	if m.errorMsg != "" {
		content.WriteString("\n\n")
		content.WriteString(styleError.Render(m.errorMsg))
	}

	footer := "[↑/↓ j/k] navigate [1-9,0] quick select [enter] open [esc] close"

	// Use auto-scroll to keep selected file visible
	return m.renderModalWithFooterAndScroll("Recent Files (MRU)", content.String(), footer, 70, 18, m.mruIndex)
}
