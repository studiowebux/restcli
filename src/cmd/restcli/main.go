package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/studiowebux/restcli/internal/cli"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/converter"
	"github.com/studiowebux/restcli/internal/mock"
	"github.com/studiowebux/restcli/internal/proxy"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/tui"
)

var (
	version = "0.0.34"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "restcli [file]",
	Short: "REST CLI - HTTP request testing tool",
	Long: `REST CLI is a powerful HTTP request testing tool with an interactive TUI.

Run without arguments to start the TUI, or provide a .http file to execute it directly.
File extension is optional - 'get-user' resolves to 'get-user.http' automatically.

When running without -p (profile), you'll be prompted for any missing variables.
Use -p to load a profile with predefined variables and headers.

Examples:
  restcli                              # Start interactive TUI
  restcli get-user                     # Execute and prompt for missing vars
  restcli run request.http             # Execute and prompt for missing vars
  restcli run api -p dev               # Use 'dev' profile (no prompts)
  restcli run api -e userId=123        # Provide var, prompt for others
  restcli run api -e env=dev -e v=2    # Multiple variables
  restcli --help                       # Show help`,
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		if err := config.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		// If a file is provided, run in CLI mode
		if len(args) > 0 {
			return runCLI(cmd, args[0])
		}

		// Otherwise, start the TUI
		return runTUI(cmd)
	},
}

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Execute an HTTP request file in CLI mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		return runCLI(cmd, args[0])
	},
}

var curl2httpCmd = &cobra.Command{
	Use:   "curl2http [curl command]",
	Short: "Convert cURL command to .http file",
	Long: `Convert cURL commands to .http file format.

You can pipe a cURL command from stdin or provide it as an argument.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCurl2Http(cmd, args)
	},
}

var openapi2httpCmd = &cobra.Command{
	Use:   "openapi2http <spec-file-or-url>",
	Short: "Convert OpenAPI specification to .http files",
	Long: `Convert OpenAPI/Swagger specifications to .http files.

Supports both local files and remote URLs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOpenapi2Http(cmd, args[0])
	},
}

var har2httpCmd = &cobra.Command{
	Use:   "har2http <har-file>",
	Short: "Convert HAR file to .http files",
	Long: `Convert HTTP Archive (HAR) file to .http files.

HAR files are exported by browsers (Chrome DevTools, Firefox, etc.) and contain
recorded HTTP requests. This command converts them to editable request files.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHar2Http(cmd, args[0])
	},
}

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Manage mock HTTP server",
	Long: `Manage mock HTTP server for testing.

Create a .mock.yaml or .mock.json file with your mock routes and responses.
The mock server will match incoming requests and return configured responses.`,
}

var mockStartCmd = &cobra.Command{
	Use:   "start [config-file]",
	Short: "Start mock HTTP server",
	Long: `Start a mock HTTP server with the specified configuration file.

If no config file is provided, looks for .mock.yaml or .mock.json files in:
  - mocks/ directory
  - current directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMockStart(cmd, args)
	},
}

var mockStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop running mock server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMockStop(cmd)
	},
}

var mockLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show mock server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMockLogs(cmd)
	},
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage debug HTTP proxy",
	Long: `Manage debug HTTP proxy for capturing and inspecting traffic.

The proxy captures all HTTP requests passing through it, allowing you to inspect
request/response details in real-time. Useful for debugging and troubleshooting.`,
}

var proxyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start debug HTTP proxy",
	Long: `Start a debug HTTP proxy on the specified port.

Configure your application to use this proxy:
  export HTTP_PROXY=http://localhost:8888
  export http_proxy=http://localhost:8888

The proxy will capture all HTTP traffic and display it in the TUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProxyStart(cmd)
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for restcli.

To load completions:

Bash (Linux):
  $ restcli completion bash > /etc/bash_completion.d/restcli

Zsh (macOS/Linux):
  # Create completions directory
  $ mkdir -p ~/.zsh/completions

  # Add to ~/.zshrc (if not already present)
  $ echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
  $ echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

  # Generate completions
  $ restcli completion zsh > ~/.zsh/completions/_restcli

  # Reload shell
  $ source ~/.zshrc
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		var buf bytes.Buffer
		var err error

		switch args[0] {
		case "bash":
			err = rootCmd.GenBashCompletion(&buf)
			if err != nil {
				return err
			}
			fmt.Print(buf.String())
		case "zsh":
			err = rootCmd.GenZshCompletion(&buf)
			if err != nil {
				return err
			}
			fmt.Print(buf.String())
		case "fish":
			err = rootCmd.GenFishCompletion(&buf, true)
			if err != nil {
				return err
			}
			fmt.Print(buf.String())
		case "powershell":
			// PowerShell doesn't need wrapping
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

// Flags for root/run command
var (
	flagProfile   string
	flagOutput    string
	flagSave      string
	flagBody      string
	flagFull      bool
	flagExtraVars []string
	flagEnvFile   string
	flagFilter    string
	flagQuery     string
)

// Flags for curl2http
var (
	curlOutputFile    string
	curlImportHeaders bool
	curlFormat        string
)

// Flags for openapi2http
var (
	openapiOutputDir  string
	openapiOrganizeBy string
	openapiFormat     string
)

// Flags for har2http
var (
	harOutputDir     string
	harImportHeaders bool
	harFormat        string
	harFilter        string
)

// Flags for proxy
var (
	proxyPort int
)

func init() {
	// Root command flags
	rootCmd.PersistentFlags().StringVarP(&flagProfile, "profile", "p", "", "Profile to use")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output format (json/yaml/text)")
	rootCmd.Flags().StringVarP(&flagSave, "save", "s", "", "Save response to file")
	rootCmd.Flags().StringVarP(&flagBody, "body", "b", "", "Override request body")
	rootCmd.Flags().BoolVarP(&flagFull, "full", "f", false, "Show full output (status, headers, body)")
	rootCmd.Flags().StringArrayVarP(&flagExtraVars, "extra-vars", "e", []string{}, "Set variable (key=value), can be repeated")
	rootCmd.Flags().StringVar(&flagEnvFile, "env-file", "", "Load environment variables from file")
	rootCmd.Flags().StringVar(&flagFilter, "filter", "", "JMESPath filter expression to apply to response")
	rootCmd.Flags().StringVarP(&flagQuery, "query", "q", "", "JMESPath query or $(bash command) to transform response")

	// Run command flags (same as root)
	runCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output format (json/yaml/text)")
	runCmd.Flags().StringVarP(&flagSave, "save", "s", "", "Save response to file")
	runCmd.Flags().StringVarP(&flagBody, "body", "b", "", "Override request body")
	runCmd.Flags().BoolVarP(&flagFull, "full", "f", false, "Show full output (status, headers, body)")
	runCmd.Flags().StringArrayVarP(&flagExtraVars, "extra-vars", "e", []string{}, "Set variable (key=value), can be repeated")
	runCmd.Flags().StringVar(&flagEnvFile, "env-file", "", "Load environment variables from file")
	runCmd.Flags().StringVar(&flagFilter, "filter", "", "JMESPath filter expression to apply to response")
	runCmd.Flags().StringVarP(&flagQuery, "query", "q", "", "JMESPath query or $(bash command) to transform response")

	// curl2http flags
	curl2httpCmd.Flags().StringVarP(&curlOutputFile, "output", "o", "", "Output file path")
	curl2httpCmd.Flags().BoolVar(&curlImportHeaders, "import-headers", false, "Include sensitive headers")
	curl2httpCmd.Flags().StringVarP(&curlFormat, "format", "f", "http", "Output format (http/json/yaml)")

	// openapi2http flags
	openapi2httpCmd.Flags().StringVarP(&openapiOutputDir, "output", "o", "requests", "Output directory")
	openapi2httpCmd.Flags().StringVar(&openapiOrganizeBy, "organize-by", "tags", "Organization strategy (tags/paths/flat)")
	openapi2httpCmd.Flags().StringVarP(&openapiFormat, "format", "f", "http", "Output format (http/json/yaml)")

	// har2http flags
	har2httpCmd.Flags().StringVarP(&harOutputDir, "output", "o", "requests", "Output directory")
	har2httpCmd.Flags().BoolVar(&harImportHeaders, "import-headers", false, "Include sensitive headers (Auth, Cookie)")
	har2httpCmd.Flags().StringVarP(&harFormat, "format", "f", "http", "Output format (http/json/yaml)")
	har2httpCmd.Flags().StringVar(&harFilter, "filter", "", "Filter requests by URL pattern")

	// Helper function to get .http files in a directory
	getHttpFilesInDir := func(dir string) []string {
		var httpFiles []string
		files, err := os.ReadDir(dir)
		if err != nil {
			return httpFiles
		}

		for _, file := range files {
			if !file.IsDir() {
				name := file.Name()
				// Support .http, .yaml, .json, .jsonc files
				if strings.HasSuffix(name, ".http") {
					// Return filename without .http extension (since it's optional)
					httpFiles = append(httpFiles, strings.TrimSuffix(name, ".http"))
				} else if strings.HasSuffix(name, ".yaml") ||
					strings.HasSuffix(name, ".json") ||
					strings.HasSuffix(name, ".jsonc") {
					// Return these with extension
					httpFiles = append(httpFiles, name)
				}
			}
		}
		return httpFiles
	}

	// Register autocomplete for --profile flag
	profileCompletionFunc := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Initialize config to get proper paths
		if err := config.Initialize(); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Load profiles
		mgr := session.NewManager()
		if err := mgr.LoadProfiles(); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Extract profile names
		profiles := mgr.GetProfiles()
		names := make([]string, len(profiles))
		for i, p := range profiles {
			names[i] = p.Name
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}

	// Register autocomplete for file argument (scans profile workdir)
	fileCompletionFunc := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Don't complete if already have an argument
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// If user is typing a path (starts with . / or ~), use default file completion
		if strings.HasPrefix(toComplete, "./") ||
			strings.HasPrefix(toComplete, "../") ||
			strings.HasPrefix(toComplete, "/") ||
			strings.HasPrefix(toComplete, "~/") {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Initialize config to get proper paths
		if err := config.Initialize(); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Get profile from persistent/inherited flags (since --profile is on root command)
		var profileName string
		// First try inherited flags (from parent commands)
		if pf := cmd.InheritedFlags().Lookup("profile"); pf != nil && pf.Changed {
			profileName = pf.Value.String()
		}
		// Fallback to local flags
		if profileName == "" {
			if cmd.Flags().Changed("profile") {
				profileName, _ = cmd.Flags().GetString("profile")
			}
		}

		// Load session manager
		mgr := session.NewManager()
		if err := mgr.LoadProfiles(); err != nil {
			// Fallback to current directory if profiles can't be loaded
			files := getHttpFilesInDir(".")
			if len(files) == 0 {
				return []string{"(no request files in current directory)"}, cobra.ShellCompDirectiveNoFileComp
			}
			return files, cobra.ShellCompDirectiveNoFileComp
		}

		// Determine workdir to scan
		var workdir string
		if profileName != "" {
			// Find profile by name
			profiles := mgr.GetProfiles()
			for _, p := range profiles {
				if p.Name == profileName && p.Workdir != "" {
					workdir = p.Workdir
					break
				}
			}
		}

		// If no profile workdir, use current directory
		if workdir == "" {
			workdir = "."
		}

		// Expand home directory if workdir starts with ~
		if strings.HasPrefix(workdir, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				workdir = strings.Replace(workdir, "~", home, 1)
			}
		}

		// Get .http files from workdir
		httpFiles := getHttpFilesInDir(workdir)

		// Show helpful message if no files found
		if len(httpFiles) == 0 {
			return []string{"(no request files found in: " + workdir + ")"}, cobra.ShellCompDirectiveNoFileComp
		}

		// Filter based on what user has typed
		var filtered []string
		for _, f := range httpFiles {
			if toComplete == "" || strings.HasPrefix(f, toComplete) {
				filtered = append(filtered, f)
			}
		}

		// Show helpful message if filtered list is empty
		if len(filtered) == 0 {
			return []string{"(no matches for '" + toComplete + "' in: " + workdir + ")"}, cobra.ShellCompDirectiveNoFileComp
		}

		return filtered, cobra.ShellCompDirectiveNoFileComp
	}

	// Register for both root and run commands
	rootCmd.RegisterFlagCompletionFunc("profile", profileCompletionFunc)
	runCmd.RegisterFlagCompletionFunc("profile", profileCompletionFunc)
	rootCmd.ValidArgsFunction = fileCompletionFunc
	runCmd.ValidArgsFunction = fileCompletionFunc

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(curl2httpCmd)
	rootCmd.AddCommand(openapi2httpCmd)
	rootCmd.AddCommand(har2httpCmd)
	rootCmd.AddCommand(completionCmd)

	// Add mock subcommands
	mockCmd.AddCommand(mockStartCmd)
	mockCmd.AddCommand(mockStopCmd)
	mockCmd.AddCommand(mockLogsCmd)
	rootCmd.AddCommand(mockCmd)

	// Add proxy subcommands
	proxyStartCmd.Flags().IntVar(&proxyPort, "proxy-port", 8888, "Proxy port")
	proxyCmd.AddCommand(proxyStartCmd)
	rootCmd.AddCommand(proxyCmd)
}

// runCLI executes a request file in CLI mode
func runCLI(cmd *cobra.Command, filePath string) error {
	opts := cli.RunOptions{
		FilePath:     filePath,
		Profile:      flagProfile,
		OutputFormat: flagOutput,
		SavePath:     flagSave,
		BodyOverride: flagBody,
		ShowFull:     flagFull,
		ExtraVars:    flagExtraVars,
		EnvFile:      flagEnvFile,
		Filter:       flagFilter,
		Query:        flagQuery,
	}
	return cli.Run(opts)
}

// runTUI starts the interactive TUI
func runTUI(cmd *cobra.Command) error {
	return tui.Run(version)
}

// runCurl2Http converts cURL to .http format
func runCurl2Http(cmd *cobra.Command, args []string) error {
	var curlCommand string

	// Read from stdin if no args provided
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		curlCommand = string(data)
	} else if len(args) > 0 {
		// Use argument
		curlCommand = args[0]
	} else {
		return fmt.Errorf("no cURL command provided (pipe it or provide as argument)")
	}

	// Validate output file
	if err := converter.ValidateOutputFile(curlOutputFile); err != nil {
		return err
	}

	opts := converter.CurlToHttpOptions{
		CurlCommand:   curlCommand,
		OutputFile:    curlOutputFile,
		ImportHeaders: curlImportHeaders,
		Format:        curlFormat,
	}

	return converter.Curl2Http(opts)
}

// runOpenapi2Http converts OpenAPI spec to .http files
func runOpenapi2Http(cmd *cobra.Command, specPath string) error {
	opts := converter.OpenAPI2HttpOptions{
		SpecPath:   specPath,
		OutputDir:  openapiOutputDir,
		OrganizeBy: openapiOrganizeBy,
		Format:     openapiFormat,
	}

	return converter.Openapi2Http(opts)
}

// findMockConfig finds mock config files
func findMockConfig(configPath string) (string, error) {
	// If config path provided, use it
	if configPath != "" {
		if _, err := os.Stat(configPath); err != nil {
			return "", fmt.Errorf("config file not found: %s", configPath)
		}
		return configPath, nil
	}

	// Search common locations
	searchPaths := []string{
		"mocks",
		".",
		"../mocks",  // Parent dir (common when running from src/)
		"..",        // Parent directory
	}

	patterns := []string{"*.mock.yaml", "*.mock.yml", "*.mock.json"}

	for _, dir := range searchPaths {
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(dir, pattern))
			if err == nil && len(matches) > 0 {
				return matches[0], nil
			}
		}
	}

	return "", fmt.Errorf("no mock config files found in mocks/, current, or parent directory")
}

// runMockStart starts the mock server
func runMockStart(cmd *cobra.Command, args []string) error {
	var configPath string
	if len(args) > 0 {
		configPath = args[0]
	}

	// Find config file
	foundPath, err := findMockConfig(configPath)
	if err != nil {
		return err
	}

	// Load config
	config, err := mock.LoadConfig(foundPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get workdir for resolving relative paths
	workdir := filepath.Dir(foundPath)

	// Create and start server
	server := mock.NewServer(config, workdir)
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	fmt.Printf("Mock server started at %s\n", server.GetAddress())
	fmt.Printf("Config: %s\n", foundPath)
	fmt.Printf("Routes: %d\n", len(config.Routes))
	fmt.Println("\nPress Ctrl+C to stop")

	// Wait indefinitely
	select {}
}

// runMockStop stops the mock server
func runMockStop(cmd *cobra.Command) error {
	return fmt.Errorf("stop command requires server management - use TUI (press 'M') or Ctrl+C on running server")
}

// runMockLogs shows mock server logs
func runMockLogs(cmd *cobra.Command) error {
	return fmt.Errorf("logs command requires active server - use TUI (press 'M') to view real-time logs")
}

// runProxyStart starts the debug proxy server
func runProxyStart(cmd *cobra.Command) error {
	// Create proxy
	p := proxy.NewProxy(proxyPort)

	// Start proxy
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	defer p.Stop()

	fmt.Fprintf(os.Stderr, "Debug proxy started at http://localhost:%d\n", proxyPort)
	fmt.Fprintf(os.Stderr, "\nConfigure your application to use this proxy:\n")
	fmt.Fprintf(os.Stderr, "  export HTTP_PROXY=http://localhost:%d\n", proxyPort)
	fmt.Fprintf(os.Stderr, "  export http_proxy=http://localhost:%d\n\n", proxyPort)
	fmt.Fprintf(os.Stderr, "Press 'y' in TUI to view captured traffic\n")
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n\n")

	// Initialize config
	if err := config.Initialize(); err != nil {
		return err
	}

	// Load session
	mgr := session.NewManager()
	if err := mgr.Load(); err != nil {
		return err
	}

	// Create TUI model
	m, err := tui.New(mgr, version)
	if err != nil {
		return err
	}

	// Set proxy in model
	m.SetProxy(p)

	// Run TUI
	program := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// runHar2Http converts HAR file to .http files
func runHar2Http(cmd *cobra.Command, harFile string) error {
	opts := converter.Har2HttpOptions{
		HarFile:       harFile,
		OutputDir:     harOutputDir,
		ImportHeaders: harImportHeaders,
		Format:        harFormat,
		Filter:        harFilter,
	}

	return converter.Har2Http(opts)
}
