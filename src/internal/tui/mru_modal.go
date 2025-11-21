package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleMRUKeys handles keyboard input in MRU mode
func (m *Model) handleMRUKeys(msg tea.KeyMsg) tea.Cmd {
	recentFiles := m.sessionMgr.GetRecentFiles()

	switch msg.String() {
	case "esc", "q":
		m.mode = ModeNormal
		m.errorMsg = ""

	case "j", "down":
		if len(recentFiles) > 0 {
			m.mruIndex = (m.mruIndex + 1) % len(recentFiles)
		}

	case "k", "up":
		if len(recentFiles) > 0 {
			m.mruIndex = (m.mruIndex - 1 + len(recentFiles)) % len(recentFiles)
		}

	case "g":
		// Go to top
		if m.gPressed {
			m.mruIndex = 0
			m.gPressed = false
		} else {
			m.gPressed = true
		}

	case "G":
		// Go to bottom
		if len(recentFiles) > 0 {
			m.mruIndex = len(recentFiles) - 1
		}
		m.gPressed = false

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
		for i, f := range m.files {
			if f.Path == selectedFile {
				m.fileIndex = i
				m.adjustScrollOffset()
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

	default:
		// Reset 'g' press on any other key
		m.gPressed = false
	}

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

	maxVisible := 15 // Maximum files to show at once
	startIdx := 0
	endIdx := len(existingFiles)

	// Adjust scroll window if needed
	if len(existingFiles) > maxVisible {
		// Center selected item in view
		startIdx = m.mruIndex - maxVisible/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxVisible
		if endIdx > len(existingFiles) {
			endIdx = len(existingFiles)
			startIdx = endIdx - maxVisible
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		filePath := existingFiles[i]
		displayPath := filePath

		// Show relative path if in workdir
		if workdir != "" {
			if rel, err := filepath.Rel(workdir, filePath); err == nil {
				if !strings.HasPrefix(rel, "..") {
					displayPath = rel
				}
			}
		}

		if i == m.mruIndex {
			content.WriteString(fmt.Sprintf("→ %s\n", displayPath))
		} else {
			content.WriteString(fmt.Sprintf("  %s\n", displayPath))
		}
	}

	// Show scroll indicator if needed
	if len(existingFiles) > maxVisible {
		content.WriteString(fmt.Sprintf("\n[%d-%d of %d files]", startIdx+1, endIdx, len(existingFiles)))
	}

	// Show error if present
	if m.errorMsg != "" {
		content.WriteString("\n\n")
		content.WriteString(styleError.Render(m.errorMsg))
	}

	footer := "[↑/↓ j/k] navigate [enter] open [esc] close"

	return m.renderModalWithFooter("Recent Files (MRU)", content.String(), footer, 70, 20)
}
