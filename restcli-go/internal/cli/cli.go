package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/history"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/types"
	"gopkg.in/yaml.v3"
)

// RunOptions contains options for running a request in CLI mode
type RunOptions struct {
	FilePath    string
	Profile     string
	OutputFormat string  // json, yaml, text
	SavePath    string
	BodyOverride string
	ShowFull    bool
}

// Run executes a request file in CLI mode
func Run(opts RunOptions) error {
	// Load session manager
	mgr := session.NewManager()
	if err := mgr.Load(); err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Set active profile if specified
	if opts.Profile != "" {
		if err := mgr.SetActiveProfile(opts.Profile); err != nil {
			return fmt.Errorf("failed to set profile: %w", err)
		}
	}

	// Get active profile
	profile := mgr.GetActiveProfile()

	// Determine working directory
	workdir, err := config.GetWorkingDirectory(profile.Workdir)
	if err != nil {
		return err
	}

	// Resolve file path (relative to workdir or absolute)
	filePath := opts.FilePath
	if !strings.HasPrefix(filePath, "/") {
		// Check if file exists in current directory first
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Try in workdir
			filePath = fmt.Sprintf("%s/%s", workdir, filePath)
		}
	}

	// Parse the request file
	requests, err := parser.Parse(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if len(requests) == 0 {
		return fmt.Errorf("no requests found in file")
	}

	// Use first request (TODO: support selecting specific request by name)
	request := requests[0]

	// Override body if specified
	if opts.BodyOverride != "" {
		request.Body = opts.BodyOverride
	} else {
		// Check for stdin body override (for piping)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped in
			bodyBytes, err := io.ReadAll(os.Stdin)
			if err == nil && len(bodyBytes) > 0 {
				request.Body = string(bodyBytes)
			}
		}
	}

	// Merge profile headers with request headers
	mergedHeaders := make(map[string]string)
	for k, v := range profile.Headers {
		mergedHeaders[k] = v
	}
	for k, v := range request.Headers {
		mergedHeaders[k] = v
	}
	request.Headers = mergedHeaders

	// Resolve variables
	resolver := parser.NewVariableResolver(profile.Variables, mgr.GetSession().Variables)
	resolvedRequest, err := resolver.ResolveRequest(&request)
	if err != nil {
		return fmt.Errorf("failed to resolve variables: %w", err)
	}

	// Execute request
	result, err := executor.Execute(resolvedRequest)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Save to history if enabled
	if mgr.IsHistoryEnabled() {
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
		if err := os.WriteFile(opts.SavePath, []byte(output), 0644); err != nil {
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
