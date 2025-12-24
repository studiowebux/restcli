package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/stresstest"
)

// renderStressTestConfig renders the stress test configuration modal
func (m *Model) renderStressTestConfig() string {
	modalWidth := m.width - 10
	if modalWidth > 80 {
		modalWidth = 80
	}

	var content strings.Builder

	// Title
	content.WriteString(styleTitle.Render("Stress Test Configuration") + "\n\n")

	if m.stressTestState.GetConfigEdit() == nil {
		// Initialize new config if not editing
		m.stressTestState.SetConfigEdit(&stresstest.Config{
			Name:              "",
			RequestFile:       "",
			ConcurrentConns:   10,
			TotalRequests:     100,
			RampUpDurationSec: 0,
			TestDurationSec:   0,
		})

		// Pre-fill with current file if available
		currentFile := m.fileExplorer.GetCurrentFile()
		if currentFile != nil {
			m.stressTestState.GetConfigEdit().RequestFile = currentFile.Path
		}
	}

	// Field labels and values
	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Config Name:", m.stressTestState.GetConfigEdit().Name, "Unique name for this configuration"},
		{"Request File:", m.stressTestState.GetConfigEdit().RequestFile, "Path to .http file"},
		{"Concurrent Connections:", fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().ConcurrentConns), "Number of parallel workers (1-1000)"},
		{"Total Requests:", fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().TotalRequests), "Total number of requests to send"},
		{"Ramp-Up Duration (sec):", fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().RampUpDurationSec), "Time to gradually increase load (0=no ramp)"},
		{"Test Duration (sec):", fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().TestDurationSec), "Max test duration (0=unlimited)"},
	}

	for i, field := range fields {
		// Highlight current field
		labelStyle := styleSubtle
		valueStyle := lipgloss.NewStyle()
		isFocused := i == m.stressTestState.GetConfigField()

		if isFocused {
			labelStyle = styleTitleFocused
			valueStyle = styleTitleFocused.Underline(true)
		}

		content.WriteString(labelStyle.Render(fmt.Sprintf("%-30s", field.label)))

		// For Request File field, show just filename (not full path)
		if i == 1 {
			if isFocused {
				// Show just the filename, not full path
				value := m.stressTestState.GetConfigInput()
				if value != "" {
					// Extract just the filename for display
					lastSlash := -1
					for idx, ch := range value {
						if ch == '/' {
							lastSlash = idx
						}
					}
					if lastSlash >= 0 && lastSlash < len(value)-1 {
						value = value[lastSlash+1:]
					}
					value = valueStyle.Render(value)
				} else if m.stressTestState.GetFilePickerActive() && len(m.stressTestState.GetFilePickerFiles()) == 0 {
					// Show error if picker is active but no files found
					value = styleSubtle.Render("No compatible files found (.http, .yaml, .yml, .json, .jsonc)")
				} else {
					value = styleSubtle.Render("<use ↑/↓ to select, Enter to confirm>")
				}
				content.WriteString(value)
			} else {
				// Show just filename, not full path
				displayValue := field.value
				if displayValue != "" {
					lastSlash := -1
					for idx, ch := range displayValue {
						if ch == '/' {
							lastSlash = idx
						}
					}
					if lastSlash >= 0 && lastSlash < len(displayValue)-1 {
						displayValue = displayValue[lastSlash+1:]
					}
				}
				if displayValue == "" {
					displayValue = styleSubtle.Render("<empty>")
				}
				content.WriteString(displayValue)
			}
		} else {
			// Regular field display
			if isFocused {
				value := m.stressTestState.GetConfigInput()

				// Show placeholder if empty
				if value == "" {
					value = styleSubtle.Render("<empty>")
				} else {
					// Show cursor
					if m.stressTestState.GetConfigCursor() <= len(value) {
						value = value[:m.stressTestState.GetConfigCursor()] + "█" + value[m.stressTestState.GetConfigCursor():]
					}
					value = valueStyle.Render(value)
				}

				content.WriteString(value)
			} else {
				// Show current value
				displayValue := field.value
				if displayValue == "" {
					displayValue = styleSubtle.Render("<empty>")
				}
				content.WriteString(displayValue)
			}
		}

		content.WriteString("\n")

		// Show hint for current field
		if isFocused {
			content.WriteString(styleSubtle.Render("  ↳ " + field.hint) + "\n")
		}

		// Show file picker dropdown for Request File field
		if i == 1 && isFocused && m.stressTestState.GetFilePickerActive() && len(m.stressTestState.GetFilePickerFiles()) > 0 {
			content.WriteString("\n")
			content.WriteString(styleTitleFocused.Render("  Available Files:") + "\n\n")

			// Show up to 10 files
			maxDisplay := 10
			startIdx := m.stressTestState.GetFilePickerIndex()
			if startIdx > len(m.stressTestState.GetFilePickerFiles())-maxDisplay {
				startIdx = len(m.stressTestState.GetFilePickerFiles()) - maxDisplay
			}
			if startIdx < 0 {
				startIdx = 0
			}

			endIdx := startIdx + maxDisplay
			if endIdx > len(m.stressTestState.GetFilePickerFiles()) {
				endIdx = len(m.stressTestState.GetFilePickerFiles())
			}

			for j := startIdx; j < endIdx; j++ {
				file := m.stressTestState.GetFilePickerFiles()[j]
				line := fmt.Sprintf("  %s", file.Name)
				if j == m.stressTestState.GetFilePickerIndex() {
					content.WriteString(styleSelected.Render("> " + line) + "\n")
				} else {
					content.WriteString("  " + line + "\n")
				}
			}

			content.WriteString("\n")
			if len(m.stressTestState.GetFilePickerFiles()) > maxDisplay {
				showing := fmt.Sprintf("  Showing %d-%d of %d files", startIdx+1, endIdx, len(m.stressTestState.GetFilePickerFiles()))
				content.WriteString(styleSubtle.Render(showing) + "\n")
			}
		}
	}

	content.WriteString("\n")

	// Instructions
	var footer string
	if m.stressTestState.GetFilePickerActive() && m.stressTestState.GetConfigField() == 1 {
		if len(m.stressTestState.GetFilePickerFiles()) == 0 {
			footer = "No compatible files found | ESC: Cancel"
		} else {
			footer = "↑/↓: Select file | Enter: Confirm (required) | Ctrl+S: Save & Start | ESC: Cancel"
		}
	} else if m.stressTestState.GetConfigField() == 1 && m.stressTestState.GetConfigInput() == "" {
		footer = "File selection required | Navigate to select file | ESC: Cancel"
	} else {
		footer = "↑/↓: Navigate fields (auto-saves) | Type to edit | Ctrl+S: Save & Start | Ctrl+L: Load | ESC: Cancel"
	}
	content.WriteString(styleSubtle.Render(footer))

	// Center the modal
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content.String()),
	)
}

// updateStressTestConfigInput updates the config input buffer based on current field
func (m *Model) updateStressTestConfigInput() {
	if m.stressTestState.GetConfigEdit() == nil {
		return
	}

	switch m.stressTestState.GetConfigField() {
	case 0:
		m.stressTestState.SetConfigInput(m.stressTestState.GetConfigEdit().Name)
	case 1:
		m.stressTestState.SetConfigInput(m.stressTestState.GetConfigEdit().RequestFile)
	case 2:
		m.stressTestState.SetConfigInput(fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().ConcurrentConns))
	case 3:
		m.stressTestState.SetConfigInput(fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().TotalRequests))
	case 4:
		m.stressTestState.SetConfigInput(fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().RampUpDurationSec))
	case 5:
		m.stressTestState.SetConfigInput(fmt.Sprintf("%d", m.stressTestState.GetConfigEdit().TestDurationSec))
	}
	m.stressTestState.SetConfigCursor(len(m.stressTestState.GetConfigInput()))
}

// applyStressTestConfigInput applies the input buffer to the config field
func (m *Model) applyStressTestConfigInput() error {
	if m.stressTestState.GetConfigEdit() == nil {
		return nil
	}

	value := m.stressTestState.GetConfigInput()

	switch m.stressTestState.GetConfigField() {
	case 0: // Config Name
		m.stressTestState.GetConfigEdit().Name = value
	case 1: // Request File
		m.stressTestState.GetConfigEdit().RequestFile = value
	case 2: // Concurrent Connections
		if val, err := strconv.Atoi(value); err == nil && val > 0 && val <= 1000 {
			m.stressTestState.GetConfigEdit().ConcurrentConns = val
		} else {
			return fmt.Errorf("concurrent connections must be between 1 and 1000")
		}
	case 3: // Total Requests
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			m.stressTestState.GetConfigEdit().TotalRequests = val
		} else {
			return fmt.Errorf("total requests must be greater than 0")
		}
	case 4: // Ramp-Up Duration
		if val, err := strconv.Atoi(value); err == nil && val >= 0 {
			m.stressTestState.GetConfigEdit().RampUpDurationSec = val
		} else {
			return fmt.Errorf("ramp-up duration must be 0 or greater")
		}
	case 5: // Test Duration
		if val, err := strconv.Atoi(value); err == nil && val >= 0 {
			m.stressTestState.GetConfigEdit().TestDurationSec = val
		} else {
			return fmt.Errorf("test duration must be 0 or greater")
		}
	}

	return nil
}

// renderStressTestLoadConfig renders the load config selection modal
func (m *Model) renderStressTestLoadConfig() string {
	modalWidth := m.width - 10
	if modalWidth > 80 {
		modalWidth = 80
	}

	var content strings.Builder

	content.WriteString(styleTitle.Render("Load Stress Test Configuration") + "\n\n")

	if len(m.stressTestState.GetConfigs()) == 0 {
		content.WriteString("No saved configurations found.\n\n")
		content.WriteString("Create a new configuration to get started.\n")
	} else {
		content.WriteString(fmt.Sprintf("%d saved configuration(s):\n\n", len(m.stressTestState.GetConfigs())))

		for i, config := range m.stressTestState.GetConfigs() {
			line := fmt.Sprintf("%s | %d conns | %d reqs", config.Name, config.ConcurrentConns, config.TotalRequests)

			if i == m.stressTestState.GetConfigIndex() {
				content.WriteString(styleSelected.Render("> " + line))
			} else {
				content.WriteString("  " + line)
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	footer := "↑/↓: Navigate | Enter: Load | d: Delete | ESC: Cancel"
	content.WriteString(styleSubtle.Render(footer))

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content.String()),
	)
}
