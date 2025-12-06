package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/filter"
	"github.com/studiowebux/restcli/internal/history"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/types"
	"gopkg.in/yaml.v3"
)

// promptForVariable prompts the user to enter a value for a variable
func promptForVariable(name string) (string, error) {
	fmt.Fprintf(os.Stderr, "Enter value for '%s': ", name)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

// isInteractive checks if stdin is a terminal (not piped)
func isInteractive() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// RunOptions contains options for running a request in CLI mode
type RunOptions struct {
	FilePath     string
	Profile      string
	OutputFormat string   // json, yaml, text
	SavePath     string
	BodyOverride string
	ShowFull     bool
	ExtraVars    []string // key=value pairs from -e flag
	EnvFile      string   // path to .env file
	Filter       string   // JMESPath filter expression
	Query        string   // JMESPath query or $(bash command)
}

// Run executes a request file in CLI mode
func Run(opts RunOptions) error {
	// Load session manager
	mgr := session.NewManager()
	if err := mgr.Load(); err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Determine if we're using a profile or interactive mode
	useProfile := opts.Profile != ""
	var profile *types.Profile
	var profileVars map[string]types.VariableValue
	var sessionVars map[string]string

	if useProfile {
		// Set active profile if specified
		if err := mgr.SetActiveProfile(opts.Profile); err != nil {
			return fmt.Errorf("failed to set profile: %w", err)
		}
		profile = mgr.GetActiveProfile()
		profileVars = profile.Variables
		sessionVars = mgr.GetSession().Variables
	} else {
		// No profile - use empty vars (will prompt for missing)
		profile = &types.Profile{}
		profileVars = make(map[string]types.VariableValue)
		sessionVars = make(map[string]string)
	}

	// Determine working directory
	workdir, err := config.GetWorkingDirectory(profile.Workdir)
	if err != nil {
		return err
	}

	// Resolve file path (supports extension-less names like "get-user" -> "get-user.http")
	filePath, err := resolveFilePath(opts.FilePath, workdir)
	if err != nil {
		return err
	}

	// Parse the request file
	requests, err := parser.Parse(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if len(requests) == 0 {
		return fmt.Errorf("no requests found in file: %s", filePath)
	}

	// Use first request (TODO: support selecting specific request by name)
	request := requests[0]

	// Check if confirmation is required
	if request.RequiresConfirmation {
		fmt.Printf("Request '%s' requires confirmation.\n", request.Name)
		fmt.Printf("Method: %s\n", request.Method)
		fmt.Printf("URL: %s\n", request.URL)
		fmt.Print("\nProceed? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("request execution cancelled by user")
		}
	}

	// Check if stdin is being piped (for body override)
	stdinPiped := false
	if opts.BodyOverride != "" {
		request.Body = opts.BodyOverride
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped in
			stdinPiped = true
			bodyBytes, err := io.ReadAll(os.Stdin)
			if err == nil && len(bodyBytes) > 0 {
				request.Body = string(bodyBytes)
			}
		}
	}

	// Merge profile headers with request headers (only if using profile)
	mergedHeaders := make(map[string]string)
	if useProfile {
		for k, v := range profile.Headers {
			mergedHeaders[k] = v
		}
	}
	for k, v := range request.Headers {
		mergedHeaders[k] = v
	}
	request.Headers = mergedHeaders

	// Parse CLI extra vars (key=value format) and resolve aliases
	cliVars := make(map[string]string)
	for _, ev := range opts.ExtraVars {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 2 {
			varName := parts[0]
			varValue := parts[1]

			// Check if the value is an alias for a multi-value variable (only if using profile)
			if useProfile {
				if profileVar, ok := profile.Variables[varName]; ok && profileVar.IsMultiValue() {
					if profileVar.MultiValue.Aliases != nil {
						if idx, aliasFound := profileVar.MultiValue.Aliases[varValue]; aliasFound {
							// Resolve alias to actual value
							if idx >= 0 && idx < len(profileVar.MultiValue.Options) {
								varValue = profileVar.MultiValue.Options[idx]
							}
						}
					}
				}
			}

			cliVars[varName] = varValue
		} else if len(parts) == 1 && parts[0] != "" {
			// Allow -e key (sets to empty string)
			cliVars[parts[0]] = ""
		}
	}

	// Load environment variables
	envVars := parser.LoadSystemEnv()

	// Load additional env vars from file if specified
	if opts.EnvFile != "" {
		fileEnvVars, err := parser.LoadEnvFile(opts.EnvFile)
		if err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}
		// File vars override system vars
		for k, v := range fileEnvVars {
			envVars[k] = v
		}
	}

	// If no profile specified, prompt for missing variables interactively
	if !useProfile {
		// Extract all variables required by the request
		requiredVars := parser.ExtractRequestVariables(&request)

		// Find variables that are not satisfied by cliVars or envVars
		var missingVars []string
		for _, varName := range requiredVars {
			// Skip if provided via -e flag
			if _, ok := cliVars[varName]; ok {
				continue
			}
			// Skip env.* variables if they exist in environment
			if strings.HasPrefix(varName, "env.") {
				envKey := varName[4:]
				if _, ok := envVars[envKey]; ok {
					continue
				}
			}
			missingVars = append(missingVars, varName)
		}

		// Prompt for missing variables
		if len(missingVars) > 0 {
			if stdinPiped {
				return fmt.Errorf("cannot prompt for variables while stdin is piped (missing: %s)", strings.Join(missingVars, ", "))
			}
			if !isInteractive() {
				return fmt.Errorf("missing variables (non-interactive mode): %s", strings.Join(missingVars, ", "))
			}

			for _, varName := range missingVars {
				value, err := promptForVariable(varName)
				if err != nil {
					return fmt.Errorf("failed to read input for '%s': %w", varName, err)
				}
				cliVars[varName] = value
			}
		}
	}

	// If using profile, check for multi-value variables that need selection
	if useProfile {
		// Extract all variables required by the request
		requiredVars := parser.ExtractRequestVariables(&request)

		for _, varName := range requiredVars {
			// Skip if already provided via -e flag
			if _, ok := cliVars[varName]; ok {
				continue
			}

			// Check if this is a multi-value variable in the profile
			if profileVar, ok := profileVars[varName]; ok && profileVar.IsMultiValue() {
				mv := profileVar.MultiValue

				// Always prompt for multi-value variables (unless -e was used)
				if stdinPiped {
					return fmt.Errorf("multi-value variable '%s' requires selection. Use -e %s=<value> or -e %s=<alias>", varName, varName, varName)
				}
				if !isInteractive() {
					return fmt.Errorf("multi-value variable '%s' requires selection (non-interactive mode). Use -e %s=<value>", varName, varName)
				}

				// Show interactive selector (active index will be the default)
				value, err := promptForMultiValueVariable(varName, mv)
				if err != nil {
					return fmt.Errorf("failed to select value for '%s': %w", varName, err)
				}
				cliVars[varName] = value
			}
		}

		// Check for interactive variables that always need prompting
		for varName, varValue := range profileVars {
			// Skip if already provided via -e flag
			if _, ok := cliVars[varName]; ok {
				continue
			}

			// If variable is marked as interactive, always prompt
			if varValue.Interactive {
				if stdinPiped {
					return fmt.Errorf("interactive variable '%s' requires input. Use -e %s=<value>", varName, varName)
				}
				if !isInteractive() {
					return fmt.Errorf("interactive variable '%s' requires input (non-interactive mode). Use -e %s=<value>", varName, varName)
				}

				// Prompt for the value
				value, err := promptForVariable(varName)
				if err != nil {
					return fmt.Errorf("failed to read input for interactive variable '%s': %w", varName, err)
				}
				cliVars[varName] = value
			}
		}
	}

	// Resolve variables (CLI vars have highest priority)
	resolver := parser.NewVariableResolver(profileVars, sessionVars, cliVars, envVars)
	resolvedRequest, err := resolver.ResolveRequest(&request)
	if err != nil {
		return fmt.Errorf("failed to resolve variables: %w", err)
	}

	// Warn about unresolved variables
	if unresolved := resolver.GetUnresolvedVariables(); len(unresolved) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: unresolved variables: %s\n", strings.Join(unresolved, ", "))
	}

	// Warn about shell command errors
	if shellErrs := resolver.GetShellErrors(); len(shellErrs) > 0 {
		for _, err := range shellErrs {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
	}

	// Merge TLS config: request-level overrides profile-level
	var tlsConfig *types.TLSConfig
	if useProfile && profile.TLS != nil {
		tlsConfig = profile.TLS
	}
	if request.TLS != nil {
		tlsConfig = request.TLS
	}

	// Execute request with streaming support (matches TUI behavior)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nStream cancelled by user")
		cancel()
	}()

	// Use streaming executor with real-time output callback
	var activeProfile *types.Profile
	if useProfile {
		activeProfile = profile
	}
	result, err := executor.ExecuteWithStreaming(ctx, resolvedRequest, tlsConfig, activeProfile, func(chunk []byte, done bool) {
		if !done {
			// Write chunks directly to stdout for real-time output
			os.Stdout.Write(chunk)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Save to history if enabled (check both global and profile settings)
	shouldSaveHistory := mgr.IsHistoryEnabled()
	if useProfile {
		profile := mgr.GetActiveProfile()
		if profile != nil && profile.HistoryEnabled != nil {
			// Profile setting overrides global
			shouldSaveHistory = *profile.HistoryEnabled
		}
	}
	if shouldSaveHistory {
		if err := history.Save(filePath, resolvedRequest, result); err != nil {
			// Don't fail if history save fails, just log
			fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
		}
	}

	// Auto-extract tokens if there's a token pattern in the response
	if result.Status >= 200 && result.Status < 300 {
		// Try to extract access_token from JSON response
		if token, err := parser.ExtractJSONToken(result.Body, "access_token"); err == nil {
			resolver.AddSessionVariable("token", token)
			mgr.SetSessionVariable("token", token)
		}
		// Also try "token" field
		if token, err := parser.ExtractJSONToken(result.Body, "token"); err == nil {
			resolver.AddSessionVariable("token", token)
			mgr.SetSessionVariable("token", token)
		}
	}

	// Apply filter and query to response body
	// Priority: CLI flags > request-level > profile defaults
	filterExpr := opts.Filter
	if filterExpr == "" {
		filterExpr = request.Filter
	}
	if filterExpr == "" && useProfile {
		filterExpr = profile.DefaultFilter
	}

	queryExpr := opts.Query
	if queryExpr == "" {
		queryExpr = request.Query
	}
	if queryExpr == "" && useProfile {
		queryExpr = profile.DefaultQuery
	}

	// Apply filter/query if specified
	if filterExpr != "" || queryExpr != "" {
		filteredBody, err := filter.Apply(result.Body, filterExpr, queryExpr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: filter/query error: %v\n", err)
		} else {
			result.Body = filteredBody
		}
	}

	// Parse escape sequences AFTER filter/query (as the final processing step)
	if request.ParseEscapes {
		result.Body = executor.ParseEscapeSequences(result.Body)
	}

	// Determine output format
	outputFormat := opts.OutputFormat
	if outputFormat == "" {
		// Use profile default or auto-detect based on TTY
		if profile.Output != "" {
			outputFormat = profile.Output
		} else {
			stat, _ := os.Stdout.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Output is being piped, just show body
				outputFormat = "body"
			} else {
				outputFormat = "text"
			}
		}
	}

	// Format and output response
	output, err := formatOutput(result, outputFormat, opts.ShowFull)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Save to file if specified
	if opts.SavePath != "" {
		if err := os.WriteFile(opts.SavePath, []byte(output), config.FilePermissions); err != nil {
			return fmt.Errorf("failed to save response: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Response saved to %s\n", opts.SavePath)
	} else {
		fmt.Print(output)
	}

	// Exit with error code if request failed
	if result.Error != "" || result.Status >= 400 {
		os.Exit(1)
	}

	return nil
}

// formatOutput formats the result based on the output format
func formatOutput(result *types.RequestResult, format string, showFull bool) (string, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "yaml":
		data, err := yaml.Marshal(result)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "body":
		// Just return the body
		return result.Body, nil

	case "text":
		fallthrough
	default:
		// Text format
		var sb strings.Builder

		// Status line
		statusColor := getStatusColor(result.Status)
		sb.WriteString(fmt.Sprintf("%s%s%s\n", statusColor, result.StatusText, colorReset))

		// Duration and size
		sb.WriteString(fmt.Sprintf("Duration: %s | Size: %s\n",
			executor.FormatDuration(result.Duration),
			executor.FormatSize(result.ResponseSize)))

		if showFull {
			// Headers
			if len(result.Headers) > 0 {
				sb.WriteString("\nHeaders:\n")
				for key, value := range result.Headers {
					sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
				}
			}
		}

		// Body
		if result.Body != "" {
			if showFull {
				sb.WriteString("\nBody:\n")
			} else if !showFull {
				sb.WriteString("\n")
			}
			sb.WriteString(result.Body)
			sb.WriteString("\n")
		}

		// Error
		if result.Error != "" {
			sb.WriteString(fmt.Sprintf("\n%sError: %s%s\n", colorRed, result.Error, colorReset))
		}

		return sb.String(), nil
	}
}

// ANSI color codes
const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
)

func getStatusColor(status int) string {
	if status >= 200 && status < 300 {
		return colorGreen
	} else if status >= 400 {
		return colorRed
	}
	return colorYellow
}

// resolveFilePath attempts to find the actual file path, trying common extensions
// if the exact path doesn't exist. Returns the resolved path and any error.
func resolveFilePath(basePath, workdir string) (string, error) {
	// Supported extensions in priority order (empty string = exact match first)
	extensions := []string{"", ".http", ".yaml", ".yml", ".json"}

	// If absolute path, only check with extensions
	if filepath.IsAbs(basePath) {
		for _, ext := range extensions {
			candidate := basePath + ext
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("file not found: %s (tried .http, .yaml, .yml, .json extensions)", basePath)
	}

	// Check in current directory first
	for _, ext := range extensions {
		candidate := basePath + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Check in workdir
	for _, ext := range extensions {
		candidate := filepath.Join(workdir, basePath+ext)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("file not found: %s (searched current directory and %s, tried .http, .yaml, .yml, .json extensions)", basePath, workdir)
}
