package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/types"
)

// updateWebSocketViews updates the viewport content (call this when data changes)
func (m *Model) updateWebSocketViews(width, height int) {
	// Calculate pane dimensions - WebSocket has header+status+spacing+footer = 5 lines overhead
	modalWidth := width - 6
	modalHeight := height - 3
	paneHeight := modalHeight - 5 // header(1) + status(1) + 2 empty lines(2) + footer(1) = 5

	// Split width: 60% history, 40% menu
	historyWidth := (modalWidth * 6) / 10
	menuWidth := modalWidth - historyWidth - 3 // -3 for border/padding

	// Update history viewport
	historyContentWidth := historyWidth - 4
	historyContentHeight := paneHeight - 2 // Account for title + empty line
	m.updateWebSocketHistoryView(historyContentWidth, historyContentHeight)

	// Update menu viewport
	menuContentWidth := menuWidth - 4
	menuContentHeight := paneHeight - 2  // Account for title + empty line
	m.updateWebSocketMenuView(menuContentWidth, menuContentHeight)
}

// renderWebSocketModal renders the split-pane WebSocket modal
// Left pane: message history
// Right pane: predefined message menu
func (m *Model) renderWebSocketModal() string {
	// Calculate pane dimensions - WebSocket has header+status+spacing+footer = 5 lines overhead
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 5 // header(1) + status(1) + 2 empty lines(2) + footer(1) = 5

	// Split width: 60% history, 40% menu
	historyWidth := (modalWidth * 6) / 10
	menuWidth := modalWidth - historyWidth - 3 // -3 for border/padding

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62")).
		Padding(0, 1)

	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	focusedPaneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Build header
	header := headerStyle.Render(fmt.Sprintf(" WebSocket: %s ", m.wsURL))

	// Color-code status based on connection state
	var statusColor string
	var statusIndicator string
	switch m.wsConnectionStatus {
	case "connected":
		statusColor = "10" // Green
		statusIndicator = "●"
	case "connecting":
		statusColor = "226" // Yellow
		statusIndicator = "◐"
	case "disconnected", "not connected":
		statusColor = "241" // Gray
		statusIndicator = "○"
	default:
		statusColor = "9" // Red for errors
		statusIndicator = "✖"
	}

	statusColorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true)

	statusText := fmt.Sprintf(" Status: %s %s | Messages: %d/%d ",
		statusIndicator,
		m.wsConnectionStatus,
		len(m.wsMessages),
		len(m.wsSendableMessages))

	status := statusColorStyle.Render(statusText)

	// Viewport dimensions are set in updateWebSocketViews(), not here

	// Build left pane: message history
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))

	historyTitle := titleStyle.Render("Message History")

	// Adjust viewport height based on whether search/filter bar is showing
	var historyContentHeight int
	if m.wsSearchMode || m.wsSearchQuery != "" {
		// Account for title + search/filter bar + empty line
		historyContentHeight = paneHeight - 3
	} else {
		// Account for title + empty line
		historyContentHeight = paneHeight - 2
	}
	m.wsHistoryView.Height = historyContentHeight

	// Add search bar if in search mode
	var historyContent string
	if m.wsSearchMode {
		searchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)
		searchBar := searchStyle.Render(fmt.Sprintf("Search: %s█", m.wsSearchQuery))
		historyContent = lipgloss.JoinVertical(lipgloss.Left, historyTitle, searchBar, "", m.wsHistoryView.View())
	} else if m.wsSearchQuery != "" {
		// Show active filter (not in edit mode)
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		filterBar := filterStyle.Render(fmt.Sprintf("Filter: %s (/ to edit, esc to clear)", m.wsSearchQuery))
		historyContent = lipgloss.JoinVertical(lipgloss.Left, historyTitle, filterBar, "", m.wsHistoryView.View())
	} else {
		historyContent = lipgloss.JoinVertical(lipgloss.Left, historyTitle, "", m.wsHistoryView.View())
	}

	historyPaneStyle := paneStyle
	if m.wsFocusedPane == "history" {
		historyPaneStyle = focusedPaneStyle
	}
	historyPane := historyPaneStyle.
		Width(historyWidth).
		Height(paneHeight).
		Render(historyContent)

	// Build right pane: predefined message menu
	menuTitle := titleStyle.Render("Predefined Messages")
	menuContent := lipgloss.JoinVertical(lipgloss.Left, menuTitle, "", m.wsMessageMenuView.View())

	menuPaneStyle := paneStyle
	if m.wsFocusedPane == "menu" {
		menuPaneStyle = focusedPaneStyle
	}
	menuPane := menuPaneStyle.
		Width(menuWidth).
		Height(paneHeight).
		Render(menuContent)

	// Combine panes side by side
	panes := lipgloss.JoinHorizontal(lipgloss.Top, historyPane, " ", menuPane)

	// Build footer with instructions
	var footer string

	// Show WebSocket-specific status message if set
	if m.wsStatusMsg != "" {
		statusMsgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)
		footer = statusMsgStyle.Render(fmt.Sprintf(" %s ", m.wsStatusMsg))
	} else if m.wsSearchMode {
		footer = statusStyle.Render(" Type to search | Enter: Apply | Esc: Cancel ")
	} else if m.wsComposerMode {
		footer = statusStyle.Render(" Type message | Enter: Send | Esc: Cancel ")
	} else {
		connectionAction := "r: Connect/Reconnect"
		if m.wsActive {
			connectionAction = "d: Disconnect | i: Compose"
		}

		if m.wsFocusedPane == "menu" {
			footer = statusStyle.Render(fmt.Sprintf(" j/k: Select | Enter: Send | /: Search | c: Copy | C: Clear | e: Export | %s | Tab: Switch | q: Close ", connectionAction))
		} else {
			footer = statusStyle.Render(fmt.Sprintf(" j/k: Scroll | /: Search | c: Copy | C: Clear | e: Export | %s | Tab: Switch | q: Close ", connectionAction))
		}
	}

	// Assemble the modal
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		status,
		"",
		panes,
		"",
		footer,
	)

	baseModal := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	// Show clear confirmation dialog if active
	if m.wsShowClearConfirm {
		confirmStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Background(lipgloss.Color("235")).
			Padding(1, 2).
			Width(50)

		confirmTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")).
			Render("Clear Message History")

		confirmText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Render(fmt.Sprintf("\nAre you sure you want to clear all %d messages?\n\n[y] Yes   [n] No", len(m.wsMessages)))

		confirmContent := lipgloss.JoinVertical(lipgloss.Left, confirmTitle, confirmText)
		confirmBox := confirmStyle.Render(confirmContent)

		// Place confirmation dialog centered over base modal
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, confirmBox)
	}

	// Show composer input if active
	if m.wsComposerMode {
		composerStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Background(lipgloss.Color("235")).
			Padding(1, 2).
			Width(60)

		composerTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Render("Compose Custom Message")

		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true)

		composerInput := inputStyle.Render(fmt.Sprintf("\n> %s█\n", m.wsComposerMessage))

		composerHint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Press Enter to send, Esc to cancel")

		composerContent := lipgloss.JoinVertical(lipgloss.Left, composerTitle, composerInput, composerHint)
		composerBox := composerStyle.Render(composerContent)

		// Place composer dialog centered over base modal
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, composerBox)
	}

	return baseModal
}

// updateWebSocketHistoryView updates the message history viewport content
func (m *Model) updateWebSocketHistoryView(width, height int) {
	// Update viewport dimensions to fill available space
	m.wsHistoryView.Width = width
	m.wsHistoryView.Height = height

	sentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10"))

	receivedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12"))

	systemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	timestampStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	if len(m.wsMessages) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No messages yet...")
		m.wsHistoryView.SetContent(emptyMsg)
		return
	}

	// Filter messages by search query
	var filteredMessages []types.ReceivedMessage
	if m.wsSearchQuery != "" {
		query := strings.ToLower(m.wsSearchQuery)
		for _, msg := range m.wsMessages {
			if strings.Contains(strings.ToLower(msg.Content), query) ||
				strings.Contains(strings.ToLower(msg.Direction), query) ||
				strings.Contains(strings.ToLower(msg.Type), query) {
				filteredMessages = append(filteredMessages, msg)
			}
		}
	} else {
		filteredMessages = m.wsMessages
	}

	// Show empty state if no matches
	if len(filteredMessages) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(fmt.Sprintf("No messages matching '%s'...", m.wsSearchQuery))
		m.wsHistoryView.SetContent(emptyMsg)
		return
	}

	var messages []string
	for _, msg := range filteredMessages {
		var directionStyle lipgloss.Style
		var directionLabel string

		switch msg.Direction {
		case "sent":
			directionStyle = sentStyle
			directionLabel = "→"
		case "received":
			directionStyle = receivedStyle
			directionLabel = "←"
		case "system":
			directionStyle = systemStyle
			directionLabel = "●"
		default:
			directionStyle = timestampStyle
			directionLabel = "·"
		}

		// Format timestamp (show time only)
		timestamp := msg.Timestamp
		if len(timestamp) >= 19 {
			timestamp = timestamp[11:19] // Extract HH:MM:SS
		}

		// Format message content with word wrapping
		content := msg.Content
		maxWidth := width - 15 // Leave space for timestamp + direction

		// Create a wrapped style for the content
		contentStyle := lipgloss.NewStyle().
			Width(maxWidth)

		// Wrap the content
		wrappedContent := contentStyle.Render(content)

		// Split wrapped content into lines
		contentLines := strings.Split(wrappedContent, "\n")

		// First line includes timestamp and direction
		firstLine := fmt.Sprintf("%s %s %s",
			timestampStyle.Render(timestamp),
			directionStyle.Render(directionLabel),
			contentLines[0],
		)
		messages = append(messages, firstLine)

		// Subsequent lines are indented to align with content
		for i := 1; i < len(contentLines); i++ {
			indentedLine := fmt.Sprintf("%s   %s",
				strings.Repeat(" ", 8), // 8 spaces for timestamp
				contentLines[i],
			)
			messages = append(messages, indentedLine)
		}
	}

	messageList := strings.Join(messages, "\n")
	m.wsHistoryView.SetContent(messageList)
}

// updateWebSocketMenuView updates the message menu viewport content
func (m *Model) updateWebSocketMenuView(width, height int) {
	// Update viewport dimensions to fill available space
	m.wsMessageMenuView.Width = width
	m.wsMessageMenuView.Height = height

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	if len(m.wsPredefinedMessages) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No predefined messages...")
		m.wsMessageMenuView.SetContent(emptyMsg)
		return
	}

	if len(m.wsSendableMessages) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No sendable messages...")
		m.wsMessageMenuView.SetContent(emptyMsg)
		return
	}

	var menuItems []string
	for i, msg := range m.wsSendableMessages {
		// Build menu item
		label := msg.Name
		if label == "" {
			label = fmt.Sprintf("Message %d", i+1)
		}

		typeLabel := typeStyle.Render(fmt.Sprintf("[%s]", msg.Type))
		itemText := fmt.Sprintf(" → %s %s", label, typeLabel)

		// Apply selection styling
		if i == m.wsSelectedMessageIndex {
			itemText = selectedStyle.Render(itemText)
		} else {
			itemText = normalStyle.Render(itemText)
		}

		menuItems = append(menuItems, itemText)
	}

	menuList := strings.Join(menuItems, "\n")
	m.wsMessageMenuView.SetContent(menuList)

	// Ensure selected item is visible
	selectedLine := m.wsSelectedMessageIndex
	viewportHeight := m.wsMessageMenuView.Height

	// Scroll to keep selected item visible
	if selectedLine < m.wsMessageMenuView.YOffset {
		m.wsMessageMenuView.YOffset = selectedLine
	} else if selectedLine >= m.wsMessageMenuView.YOffset+viewportHeight {
		m.wsMessageMenuView.YOffset = selectedLine - viewportHeight + 1
	}
}

// handleWebSocketKeys handles key presses in WebSocket mode
func (m *Model) handleWebSocketKeys(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Handle confirmation dialog if showing
	if m.wsShowClearConfirm {
		switch key {
		case "y", "Y":
			// Clear history
			m.wsMessages = []types.ReceivedMessage{}
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			m.updateWebSocketHistoryView(historyWidth-4, paneHeight-2)
			m.wsShowClearConfirm = false
			return nil
		case "n", "N", "esc":
			// Cancel
			m.wsShowClearConfirm = false
			return nil
		}
		return nil
	}

	// Handle search mode
	if m.wsSearchMode {
		switch key {
		case "esc":
			// Exit search mode
			m.wsSearchMode = false
			m.wsSearchQuery = ""
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			m.updateWebSocketHistoryView(historyWidth-4, paneHeight-2)
			return nil
		case "enter":
			// Exit search mode but keep filter
			m.wsSearchMode = false
			return nil
		case "backspace":
			// Remove last character
			if len(m.wsSearchQuery) > 0 {
				m.wsSearchQuery = m.wsSearchQuery[:len(m.wsSearchQuery)-1]
				modalWidth := m.width - 6
				modalHeight := m.height - 3
				paneHeight := modalHeight - 5
				historyWidth := (modalWidth * 6) / 10
				m.updateWebSocketHistoryView(historyWidth-4, paneHeight-2)
			}
			return nil
		default:
			// Add character to search query
			if len(key) == 1 {
				m.wsSearchQuery += key
				modalWidth := m.width - 6
				modalHeight := m.height - 3
				paneHeight := modalHeight - 5
				historyWidth := (modalWidth * 6) / 10
				m.updateWebSocketHistoryView(historyWidth-4, paneHeight-2)
			}
		}
		return nil
	}

	// Handle composer mode
	if m.wsComposerMode {
		switch key {
		case "esc":
			// Cancel composer
			m.wsComposerMode = false
			m.wsComposerMessage = ""
			return nil
		case "enter":
			// Send custom message
			if m.wsActive && m.wsSendChannel != nil && m.wsComposerMessage != "" {
				message := m.wsComposerMessage
				m.wsComposerMode = false
				m.wsComposerMessage = ""
				// Send message via channel
				go func() {
					select {
					case m.wsSendChannel <- message:
					case <-time.After(1 * time.Second):
					}
				}()
			}
			return nil
		case "backspace":
			// Remove last character
			if len(m.wsComposerMessage) > 0 {
				m.wsComposerMessage = m.wsComposerMessage[:len(m.wsComposerMessage)-1]
			}
			return nil
		case "space":
			// Add space
			m.wsComposerMessage += " "
			return nil
		default:
			// Add character to message
			if len(key) == 1 {
				m.wsComposerMessage += key
			}
		}
		return nil
	}

	switch key {
	case "q", "esc":
		// Close WebSocket modal
		m.mode = ModeNormal
		if m.wsCancelFunc != nil {
			m.wsCancelFunc()
		}
		m.wsActive = false
		m.wsConnectionStatus = "disconnected"
		m.wsMessageChannel = nil
		m.wsCancelFunc = nil
		m.wsLastKey = ""
		return nil

	case "tab":
		// Switch focused pane
		if m.wsFocusedPane == "menu" {
			m.wsFocusedPane = "history"
		} else {
			m.wsFocusedPane = "menu"
		}
		m.wsLastKey = ""
		return nil

	case "d":
		// Disconnect WebSocket (keep modal open with history)
		m.wsLastKey = ""
		if m.wsActive {
			if m.wsCancelFunc != nil {
				m.wsCancelFunc()
			}
			m.wsActive = false
			m.wsConnectionStatus = "disconnected"
			m.wsMessageChannel = nil
			m.wsCancelFunc = nil
		}
		return nil

	case "r":
		// Reconnect WebSocket
		m.wsLastKey = ""
		if !m.wsActive {
			return m.connectWebSocket()
		}
		return nil

	case "C":
		// Show clear history confirmation
		m.wsLastKey = ""
		if len(m.wsMessages) > 0 {
			m.wsShowClearConfirm = true
		}
		return nil

	case "/":
		// Enter search mode
		m.wsLastKey = ""
		m.wsSearchMode = true
		m.wsSearchQuery = ""
		return nil

	case "i":
		// Enter composer mode (only when connected)
		m.wsLastKey = ""
		if m.wsActive {
			m.wsComposerMode = true
			m.wsComposerMessage = ""
		}
		return nil

	case "c":
		// Copy last message to clipboard
		m.wsLastKey = ""
		if len(m.wsMessages) > 0 {
			lastMsg := m.wsMessages[len(m.wsMessages)-1]
			if err := clipboard.WriteAll(lastMsg.Content); err != nil {
				m.wsStatusMsg = fmt.Sprintf("Failed to copy: %v", err)
			} else {
				m.wsStatusMsg = "Last message copied to clipboard"
			}
			// Clear status message after 2 seconds
			return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return clearWSStatusMsg{}
			})
		}
		return nil

	case "e":
		// Export message history to file
		m.wsLastKey = ""
		if len(m.wsMessages) > 0 {
			return m.exportWebSocketMessages()
		}
		return nil

	case "g":
		// Check for gg (go to top)
		if m.wsLastKey == "g" {
			if m.wsFocusedPane == "menu" {
				// Go to first menu item
				m.wsSelectedMessageIndex = 0
				// Update menu to show new highlighting
				modalWidth := m.width - 6
				modalHeight := m.height - 3
				paneHeight := modalHeight - 5
				historyWidth := (modalWidth * 6) / 10
				menuWidth := modalWidth - historyWidth - 3
				m.updateWebSocketMenuView(menuWidth-4, paneHeight-2)
			} else {
				// Go to top of history
				m.wsHistoryView.GotoTop()
			}
			m.wsLastKey = ""
		} else {
			m.wsLastKey = "g"
		}
		return nil

	case "G":
		// Go to bottom
		if m.wsFocusedPane == "menu" {
			// Go to last menu item
			if len(m.wsSendableMessages) > 0 {
				m.wsSelectedMessageIndex = len(m.wsSendableMessages) - 1
			}
			// Update menu to show new highlighting
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			menuWidth := modalWidth - historyWidth - 3
			m.updateWebSocketMenuView(menuWidth-4, paneHeight-2)
		} else {
			// Go to bottom of history
			m.wsHistoryView.GotoBottom()
		}
		m.wsLastKey = ""
		return nil

	case "up", "k":
		if m.wsFocusedPane == "menu" {
			// Navigate menu
			if m.wsSelectedMessageIndex > 0 {
				m.wsSelectedMessageIndex--
			}
			// Update menu to show new highlighting
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			menuWidth := modalWidth - historyWidth - 3
			m.updateWebSocketMenuView(menuWidth-4, paneHeight-2)
		} else {
			// Scroll history viewport up
			m.wsHistoryView.LineUp(1)
		}
		m.wsLastKey = ""
		return nil

	case "down", "j":
		if m.wsFocusedPane == "menu" {
			// Navigate menu
			if m.wsSelectedMessageIndex < len(m.wsSendableMessages)-1 {
				m.wsSelectedMessageIndex++
			}
			// Update menu to show new highlighting
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			menuWidth := modalWidth - historyWidth - 3
			m.updateWebSocketMenuView(menuWidth-4, paneHeight-2)
		} else {
			// Scroll history viewport down
			m.wsHistoryView.LineDown(1)
		}
		m.wsLastKey = ""
		return nil

	case "ctrl+d":
		// Page down (half page)
		if m.wsFocusedPane == "menu" {
			// Navigate menu down by half viewport height
			halfPage := m.wsMessageMenuView.Height / 2
			if halfPage < 1 {
				halfPage = 1
			}
			m.wsSelectedMessageIndex += halfPage
			if m.wsSelectedMessageIndex >= len(m.wsSendableMessages) {
				m.wsSelectedMessageIndex = len(m.wsSendableMessages) - 1
			}
			// Update menu to show new highlighting
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			menuWidth := modalWidth - historyWidth - 3
			m.updateWebSocketMenuView(menuWidth-4, paneHeight-2)
		} else {
			// Scroll history viewport down by half page
			m.wsHistoryView.HalfViewDown()
		}
		m.wsLastKey = ""
		return nil

	case "ctrl+u":
		// Page up (half page)
		if m.wsFocusedPane == "menu" {
			// Navigate menu up by half viewport height
			halfPage := m.wsMessageMenuView.Height / 2
			if halfPage < 1 {
				halfPage = 1
			}
			m.wsSelectedMessageIndex -= halfPage
			if m.wsSelectedMessageIndex < 0 {
				m.wsSelectedMessageIndex = 0
			}
			// Update menu to show new highlighting
			modalWidth := m.width - 6
			modalHeight := m.height - 3
			paneHeight := modalHeight - 5
			historyWidth := (modalWidth * 6) / 10
			menuWidth := modalWidth - historyWidth - 3
			m.updateWebSocketMenuView(menuWidth-4, paneHeight-2)
		} else {
			// Scroll history viewport up by half page
			m.wsHistoryView.HalfViewUp()
		}
		m.wsLastKey = ""
		return nil

	case "enter":
		// Send selected message (only when menu is focused)
		m.wsLastKey = ""
		if m.wsFocusedPane == "menu" {
			return m.sendWebSocketMessage(m.wsSelectedMessageIndex)
		}
		return nil

	default:
		// Clear last key for any other key
		m.wsLastKey = ""
	}

	return nil
}
