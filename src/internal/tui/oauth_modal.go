package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/keybinds"
	"github.com/studiowebux/restcli/internal/types"
)

// OAuth field indices
const (
	oauthFieldEnabled = iota
	oauthFieldAuthURL
	oauthFieldTokenURL
	oauthFieldClientID
	oauthFieldClientSecret
	oauthFieldRedirectURI
	oauthFieldScope
	oauthFieldResponseType
	oauthFieldPort
	oauthFieldTokenKey
	oauthFieldCount
)

// renderOAuthConfig renders the OAuth configuration editor
func (m *Model) renderOAuthConfig() string {
	profile := m.sessionMgr.GetActiveProfile()

	if profile.OAuth == nil {
		profile.OAuth = &types.OAuthConfig{
			Enabled:         false,
			RedirectURI:     "http://localhost:8888/callback",
			Scope:           "openid",
			ResponseType:    "code",
			WebhookPort:     8888,
			TokenStorageKey: "token",
		}
	}

	oauth := profile.OAuth

	var content strings.Builder
	content.WriteString("OAuth 2.0 Configuration\n\n")

	fields := []struct {
		label string
		value string
	}{
		{"Enabled", fmt.Sprintf("%v", oauth.Enabled)},
		{"Auth URL", oauth.AuthURL},
		{"Token URL", oauth.TokenURL},
		{"Client ID", oauth.ClientID},
		{"Client Secret", strings.Repeat("*", len(oauth.ClientSecret))},
		{"Redirect URI", oauth.RedirectURI},
		{"Scope", oauth.Scope},
		{"Response Type", oauth.ResponseType},
		{"Webhook Port", fmt.Sprintf("%d", oauth.WebhookPort)},
		{"Token Storage Key", oauth.TokenStorageKey},
	}

	for i, field := range fields {
		line := fmt.Sprintf("  %-18s %s", field.label+":", field.value)
		if i == m.oauthField {
			line = styleSelected.Render(line)
		}
		content.WriteString(line + "\n")
	}

	content.WriteString("\n↑/↓ navigate, [e]dit field, [t]oggle enabled, [s]ave, ESC cancel")

	return m.renderModal("OAuth Configuration", content.String(), 70, 25)
}

// renderOAuthEdit renders the edit input for an OAuth field
func (m *Model) renderOAuthEdit() string {
	profile := m.sessionMgr.GetActiveProfile()
	if profile.OAuth == nil {
		return m.renderOAuthConfig()
	}

	var content strings.Builder
	content.WriteString("Edit OAuth Field\n\n")

	// Get field label
	fieldLabels := []string{
		"Enabled",
		"Auth URL",
		"Token URL",
		"Client ID",
		"Client Secret",
		"Redirect URI",
		"Scope",
		"Response Type",
		"Webhook Port",
		"Token Storage Key",
	}

	if m.oauthField >= 0 && m.oauthField < len(fieldLabels) {
		content.WriteString(fmt.Sprintf("%s:\n", fieldLabels[m.oauthField]))
	}

	// Show input with cursor
	inputWithCursor := m.inputValue[:m.oauthCursor] + "█" + m.inputValue[m.oauthCursor:]
	content.WriteString(inputWithCursor + "\n")

	content.WriteString("\nEnter to save, ESC to cancel")
	content.WriteString("\nCtrl+V/Shift+Insert to paste, Ctrl+K to clear")

	return m.renderModal("Edit Field", content.String(), 70, 12)
}

// handleOAuthKeys handles keyboard input in OAuth config mode
func (m *Model) handleOAuthKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	if profile.OAuth == nil {
		profile.OAuth = &types.OAuthConfig{
			Enabled:         false,
			RedirectURI:     "http://localhost:8888/callback",
			Scope:           "openid",
			ResponseType:    "code",
			WebhookPort:     8888,
			TokenStorageKey: "token",
		}
	}

	// Handle special keys not in registry
	switch msg.String() {
	case "t":
		if m.oauthField == oauthFieldEnabled {
			profile.OAuth.Enabled = !profile.OAuth.Enabled
		}
		return nil

	case "e":
		// Enter edit mode for the current field
		m.inputValue = m.getOAuthFieldValue(profile.OAuth, m.oauthField)
		m.inputCursor = 0
		return m.editOAuthField()

	case "s":
		// Save OAuth config
		m.sessionMgr.SaveProfiles()
		m.mode = ModeNormal
		m.statusMsg = "OAuth configuration saved"
		return nil
	}

	// Use registry for navigation
	action, ok := m.keybinds.Match(keybinds.ContextModal, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionNavigateUp:
		if m.oauthField > 0 {
			m.oauthField--
		}

	case keybinds.ActionNavigateDown:
		if m.oauthField < oauthFieldCount-1 {
			m.oauthField++
		}
	}

	return nil
}

// editOAuthField opens an input dialog to edit an OAuth field
func (m *Model) editOAuthField() tea.Cmd {
	// Switch to edit mode for text input
	m.mode = ModeOAuthEdit
	m.oauthCursor = len(m.inputValue) // Start cursor at end
	return nil
}

// handleOAuthEditKeys handles keyboard input in OAuth edit mode
func (m *Model) handleOAuthEditKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			// Cancel edit
			m.mode = ModeOAuthConfig
			return nil

		case keybinds.ActionTextSubmit:
			// Save the edited value
			if profile.OAuth != nil {
				m.setOAuthFieldValue(profile.OAuth, m.oauthField, m.inputValue)
			}
			m.mode = ModeOAuthConfig
			return nil

		case keybinds.ActionTextMoveLeft:
			if m.oauthCursor > 0 {
				m.oauthCursor--
			}
			return nil

		case keybinds.ActionTextMoveRight:
			if m.oauthCursor < len(m.inputValue) {
				m.oauthCursor++
			}
			return nil

		case keybinds.ActionTextMoveHome:
			m.oauthCursor = 0
			return nil

		case keybinds.ActionTextMoveEnd:
			m.oauthCursor = len(m.inputValue)
			return nil

		case keybinds.ActionTextBackspace:
			if m.oauthCursor > 0 {
				m.inputValue = m.inputValue[:m.oauthCursor-1] + m.inputValue[m.oauthCursor:]
				m.oauthCursor--
			}
			return nil

		case keybinds.ActionTextDelete:
			if m.oauthCursor < len(m.inputValue) {
				m.inputValue = m.inputValue[:m.oauthCursor] + m.inputValue[m.oauthCursor+1:]
			}
			return nil

		case keybinds.ActionTextClearAfter:
			// Clear input
			m.inputValue = ""
			m.oauthCursor = 0
			return nil

		case keybinds.ActionTextPaste:
			// Paste from clipboard
			if clipText, err := clipboard.ReadAll(); err == nil {
				m.inputValue = m.inputValue[:m.oauthCursor] + clipText + m.inputValue[m.oauthCursor:]
				m.oauthCursor += len(clipText)
			}
			return nil
		}
	}

	// Regular character input
	if len(msg.String()) == 1 {
		char := msg.String()
		m.inputValue = m.inputValue[:m.oauthCursor] + char + m.inputValue[m.oauthCursor:]
		m.oauthCursor++
	}

	return nil
}

// getOAuthFieldValue returns the string value of an OAuth field
func (m *Model) getOAuthFieldValue(oauth *types.OAuthConfig, fieldIndex int) string {
	switch fieldIndex {
	case oauthFieldEnabled:
		return fmt.Sprintf("%v", oauth.Enabled)
	case oauthFieldAuthURL:
		return oauth.AuthURL
	case oauthFieldTokenURL:
		return oauth.TokenURL
	case oauthFieldClientID:
		return oauth.ClientID
	case oauthFieldClientSecret:
		return oauth.ClientSecret
	case oauthFieldRedirectURI:
		return oauth.RedirectURI
	case oauthFieldScope:
		return oauth.Scope
	case oauthFieldResponseType:
		return oauth.ResponseType
	case oauthFieldPort:
		return fmt.Sprintf("%d", oauth.WebhookPort)
	case oauthFieldTokenKey:
		return oauth.TokenStorageKey
	default:
		return ""
	}
}

// setOAuthFieldValue sets the value of an OAuth field from a string
func (m *Model) setOAuthFieldValue(oauth *types.OAuthConfig, fieldIndex int, value string) {
	switch fieldIndex {
	case oauthFieldAuthURL:
		oauth.AuthURL = value
	case oauthFieldTokenURL:
		oauth.TokenURL = value
	case oauthFieldClientID:
		oauth.ClientID = value
	case oauthFieldClientSecret:
		oauth.ClientSecret = value
	case oauthFieldRedirectURI:
		oauth.RedirectURI = value
	case oauthFieldScope:
		oauth.Scope = value
	case oauthFieldResponseType:
		oauth.ResponseType = value
	case oauthFieldPort:
		if port, err := strconv.Atoi(value); err == nil {
			oauth.WebhookPort = port
		}
	case oauthFieldTokenKey:
		oauth.TokenStorageKey = value
	}
}
