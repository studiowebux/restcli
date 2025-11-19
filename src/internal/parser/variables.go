package parser

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/studiowebux/restcli/internal/types"
)

var (
	// Variable placeholder pattern: {{varName}}
	varPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

	// Shell command pattern: $(command)
	shellPattern = regexp.MustCompile(`\$\(([^)]+)\)`)
)

// VariableResolver handles variable resolution for requests
type VariableResolver struct {
	// Variables are resolved in order: profile vars -> session vars
	profileVars map[string]types.VariableValue
	sessionVars map[string]string
}

// NewVariableResolver creates a new variable resolver
func NewVariableResolver(profileVars map[string]types.VariableValue, sessionVars map[string]string) *VariableResolver {
	if profileVars == nil {
		profileVars = make(map[string]types.VariableValue)
	}
	if sessionVars == nil {
		sessionVars = make(map[string]string)
	}

	return &VariableResolver{
		profileVars: profileVars,
		sessionVars: sessionVars,
	}
}

// ResolveRequest resolves all variables in a request (URL, headers, body)
func (vr *VariableResolver) ResolveRequest(req *types.HttpRequest) (*types.HttpRequest, error) {
	resolved := &types.HttpRequest{
		Name:          req.Name,
		Method:        req.Method,
		Headers:       make(map[string]string),
		Documentation: req.Documentation,
	}

	// Resolve URL
	url, err := vr.Resolve(req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve URL: %w", err)
	}
	resolved.URL = url

	// Resolve headers
	for key, value := range req.Headers {
		resolvedValue, err := vr.Resolve(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve header %s: %w", key, err)
		}
		resolved.Headers[key] = resolvedValue
	}

	// Resolve body
	if req.Body != "" {
		body, err := vr.Resolve(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve body: %w", err)
		}
		resolved.Body = body
	}

	return resolved, nil
}

// Resolve resolves variables and shell commands in a string
func (vr *VariableResolver) Resolve(input string) (string, error) {
	// First pass: resolve shell commands
	result, err := vr.resolveShellCommands(input)
	if err != nil {
		return "", err
	}

	// Second pass: resolve variables
	result = vr.resolveVariables(result)

	// Third pass: resolve any shell commands that were in variables
	result, err = vr.resolveShellCommands(result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// resolveVariables resolves {{varName}} placeholders
func (vr *VariableResolver) resolveVariables(input string) string {
	return varPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.TrimSpace(match[2 : len(match)-2])

		// Look up in session vars first (higher priority)
		if value, ok := vr.sessionVars[varName]; ok {
			return value
		}

		// Then look up in profile vars
		if value, ok := vr.profileVars[varName]; ok {
			return value.GetValue()
		}

		// If not found, return original placeholder
		return match
	})
}

// resolveShellCommands executes shell commands in $(command) syntax
func (vr *VariableResolver) resolveShellCommands(input string) (string, error) {
	var lastErr error

	result := shellPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract command (remove $( and ))
		command := strings.TrimSpace(match[2 : len(match)-1])

		// Execute with 5-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Use sh -c to execute the command
		cmd := exec.CommandContext(ctx, "sh", "-c", command)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			// Store error but continue (return original placeholder)
			lastErr = fmt.Errorf("shell command failed: %w\nstderr: %s", err, stderr.String())
			return match
		}

		// Return trimmed output
		return strings.TrimSpace(stdout.String())
	})

	return result, lastErr
}

// AddSessionVariable adds or updates a session variable
func (vr *VariableResolver) AddSessionVariable(name, value string) {
	vr.sessionVars[name] = value
}

// GetSessionVariables returns all session variables
func (vr *VariableResolver) GetSessionVariables() map[string]string {
	return vr.sessionVars
}

// ExtractTokens attempts to extract tokens from a response body
// This is used for auto-extracting authentication tokens
func ExtractTokens(body string, patterns map[string]string) map[string]string {
	tokens := make(map[string]string)

	for name, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := re.FindStringSubmatch(body)
		if len(matches) > 1 {
			tokens[name] = matches[1]
		}
	}

	return tokens
}

// ExtractJSONToken extracts a token from a JSON response by key path
// Example: ExtractJSONToken(body, "access_token") or "data.token"
func ExtractJSONToken(body string, keyPath string) (string, error) {
	// Simple JSON extraction - for complex paths, we could use a JSON library
	// For now, use regex to extract simple keys
	key := keyPath
	if strings.Contains(keyPath, ".") {
		parts := strings.Split(keyPath, ".")
		key = parts[len(parts)-1]
	}

	// Pattern: "key": "value" or "key":"value"
	pattern := fmt.Sprintf(`"%s"\s*:\s*"([^"]+)"`, key)
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("token not found for key: %s", keyPath)
}
