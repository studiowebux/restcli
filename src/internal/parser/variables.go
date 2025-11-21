package parser

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
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
	// Variables are resolved in order: cliVars (highest) -> envVars -> session vars -> profile vars (lowest)
	profileVars map[string]types.VariableValue
	sessionVars map[string]string
	cliVars     map[string]string // CLI vars from -e flag (highest priority)
	envVars     map[string]string // Environment variables (accessed via {{env.VAR_NAME}})
	unresolved  []string          // Track unresolved variable names
	shellErrors []string          // Track shell command errors
}

// NewVariableResolver creates a new variable resolver
// cliVars and envVars can be nil if not using them
func NewVariableResolver(profileVars map[string]types.VariableValue, sessionVars map[string]string, cliVars map[string]string, envVars map[string]string) *VariableResolver {
	if profileVars == nil {
		profileVars = make(map[string]types.VariableValue)
	}
	if sessionVars == nil {
		sessionVars = make(map[string]string)
	}
	if cliVars == nil {
		cliVars = make(map[string]string)
	}
	if envVars == nil {
		envVars = make(map[string]string)
	}

	return &VariableResolver{
		profileVars: profileVars,
		sessionVars: sessionVars,
		cliVars:     cliVars,
		envVars:     envVars,
		unresolved:  []string{},
		shellErrors: []string{},
	}
}

// GetUnresolvedVariables returns a list of variable names that couldn't be resolved
func (vr *VariableResolver) GetUnresolvedVariables() []string {
	// Return unique values
	seen := make(map[string]bool)
	unique := []string{}
	for _, v := range vr.unresolved {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}

// GetShellErrors returns a list of shell command errors that occurred during resolution
func (vr *VariableResolver) GetShellErrors() []string {
	return vr.shellErrors
}

// ExtractVariableNames extracts all unique variable names from a string
// Returns variable names without the {{ }} brackets
func ExtractVariableNames(input string) []string {
	matches := varPattern.FindAllStringSubmatch(input, -1)
	seen := make(map[string]bool)
	var names []string
	for _, match := range matches {
		if len(match) > 1 {
			name := strings.TrimSpace(match[1])
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

// ExtractRequestVariables extracts all unique variable names from a request
// Includes variables from URL, headers, and body
func ExtractRequestVariables(req *types.HttpRequest) []string {
	seen := make(map[string]bool)
	var names []string

	// Helper to add unique names
	addNames := func(vars []string) {
		for _, name := range vars {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}

	// Extract from URL
	addNames(ExtractVariableNames(req.URL))

	// Extract from headers
	for _, v := range req.Headers {
		addNames(ExtractVariableNames(v))
	}

	// Extract from body
	addNames(ExtractVariableNames(req.Body))

	return names
}

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(path string) (map[string]string, error) {
	envVars := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env file: %w", err)
	}

	return envVars, nil
}

// LoadSystemEnv loads all system environment variables
func LoadSystemEnv() map[string]string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}
	return envVars
}

// ResolveRequest resolves all variables in a request (URL, headers, body)
func (vr *VariableResolver) ResolveRequest(req *types.HttpRequest) (*types.HttpRequest, error) {
	resolved := &types.HttpRequest{
		Name:          req.Name,
		Method:        req.Method,
		Headers:       make(map[string]string),
		Documentation: req.Documentation,
		Filter:        req.Filter,
		Query:         req.Query,
		TLS:           req.TLS,
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

		// Check for env.VAR_NAME syntax
		if strings.HasPrefix(varName, "env.") {
			envKey := varName[4:] // Remove "env." prefix
			if value, ok := vr.envVars[envKey]; ok {
				return value
			}
			// Track unresolved env variable
			vr.unresolved = append(vr.unresolved, varName)
			return match
		}

		// Look up in CLI vars first (highest priority - from -e flag)
		if value, ok := vr.cliVars[varName]; ok {
			return value
		}

		// Then look up in session vars
		if value, ok := vr.sessionVars[varName]; ok {
			return value
		}

		// Then look up in profile vars (lowest priority)
		if value, ok := vr.profileVars[varName]; ok {
			return value.GetValue()
		}

		// Track unresolved variable
		vr.unresolved = append(vr.unresolved, varName)
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
			// Track error for display
			errMsg := fmt.Sprintf("$(%s): %v", command, err)
			if stderr.Len() > 0 {
				errMsg = fmt.Sprintf("$(%s): %s", command, strings.TrimSpace(stderr.String()))
			}
			vr.shellErrors = append(vr.shellErrors, errMsg)
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
