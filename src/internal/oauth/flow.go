package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// OAuthCallbackTimeout is the maximum time to wait for OAuth callback
	OAuthCallbackTimeout = 5 * time.Minute
	// TokenRequestTimeout is the timeout for token exchange HTTP requests
	TokenRequestTimeout = 30 * time.Second
)

// Config holds OAuth configuration
type Config struct {
	AuthURL      string // Base auth URL (auto-build mode)
	AuthEndpoint string // Complete auth URL (manual mode)
	TokenURL     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scope        string
	CallbackPort int
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// StartFlow initiates the OAuth PKCE flow
func StartFlow(config *Config) (*TokenResponse, error) {
	// Generate PKCE pair
	pkce, err := GeneratePKCEPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	// Generate state for CSRF protection
	state, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start callback server
	server := NewCallbackServer(config.CallbackPort)
	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Shutdown(context.Background())

	// Build authorization URL
	authURL := buildAuthURL(config, pkce.Challenge, state)

	// Open browser
	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %w\nPlease visit: %s", err, authURL)
	}

	// Wait for callback
	result, err := server.WaitForCallback(OAuthCallbackTimeout)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("authorization failed: %s", result.Error)
	}

	if result.Code == "" {
		return nil, fmt.Errorf("no authorization code received")
	}

	// Verify state matches (CSRF protection)
	if result.State != state {
		return nil, fmt.Errorf("state mismatch (possible CSRF attack)")
	}

	// Exchange code for token
	token, err := exchangeCodeForToken(config, result.Code, pkce.Verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// buildAuthURL builds the authorization URL
func buildAuthURL(config *Config, codeChallenge, state string) string {
	// Manual mode - use complete authEndpoint and append PKCE params
	if config.AuthEndpoint != "" {
		// Parse existing URL to add PKCE parameters
		baseURL := config.AuthEndpoint
		separator := "&"
		if !strings.Contains(baseURL, "?") {
			separator = "?"
		}

		// Add PKCE parameters to the existing URL
		pkceParams := url.Values{}
		pkceParams.Set("state", state)
		pkceParams.Set("code_challenge", codeChallenge)
		pkceParams.Set("code_challenge_method", "S256")

		return baseURL + separator + pkceParams.Encode()
	}

	// Auto-build mode - construct from base AuthURL
	params := url.Values{}
	params.Set("client_id", config.ClientID)
	params.Set("redirect_uri", config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", config.Scope)
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	return config.AuthURL + "?" + params.Encode()
}

// exchangeCodeForToken exchanges the authorization code for an access token
func exchangeCodeForToken(config *Config, code, verifier string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", config.RedirectURL)
	data.Set("client_id", config.ClientID)
	data.Set("code_verifier", verifier)

	// Include client secret if provided
	if config.ClientSecret != "" {
		data.Set("client_secret", config.ClientSecret)
	}

	req, err := http.NewRequest("POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: TokenRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &token, nil
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
