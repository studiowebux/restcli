package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/analytics"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/history"
	"github.com/studiowebux/restcli/internal/jsonpath"
	"github.com/studiowebux/restcli/internal/keybinds"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/proxy"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/stresstest"
	"github.com/studiowebux/restcli/internal/types"
)

// New creates a new TUI model
func New(mgr *session.Manager, version string) (Model, error) {
	// Load files
	files, err := loadFiles(mgr)
	if err != nil {
		return Model{}, err
	}

	// Initialize analytics, history, stress test, and bookmark managers using the same database file
	analyticsManager, err := analytics.NewManager(config.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: analytics disabled: %v\n", err)
	}
	historyManager, err := history.NewManager(config.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: history disabled: %v\n", err)
	}
	stressTestManager, err := stresstest.NewManager(config.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: stress test disabled: %v\n", err)
	}
	bookmarkManager, err := jsonpath.NewBookmarkManager(config.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: jsonpath bookmarks disabled: %v\n", err)
	}

	// Initialize keybindings (load user config or use defaults)
	configPath, err := keybinds.GetDefaultConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: using default keybindings: %v\n", err)
		configPath = "" // Will trigger defaults
	}

	// Auto-create keybinds.json on first run if it doesn't exist
	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Create the config file with examples
			if err := keybinds.CreateExampleConfig(configPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not create keybinds.json: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Created example keybinds configuration: %s\n", configPath)
			}
		}
	}

	keybindRegistry, err := keybinds.LoadOrDefault(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: keybinds config error, using defaults: %v\n", err)
		keybindRegistry = keybinds.NewDefaultRegistry()
	}

	// Initialize file explorer state
	fileExplorer := NewFileExplorerState()
	fileExplorer.SetFiles(files, files)

	// Initialize documentation state
	docState := NewDocumentationState()

	// Initialize history state
	historyState := NewHistoryState()

	// Initialize analytics state
	analyticsState := NewAnalyticsState(analyticsManager)

	// Initialize stress test state
	stressTestState := NewStressTestState(stressTestManager)

	// Initialize profile edit state
	profileEditState := NewProfileEditState()

	// Initialize mock server state
	mockServerState := NewMockServerState()

	// Initialize proxy server state
	proxyServerState := NewProxyServerState()

	// Initialize rename state
	renameState := NewRenameState()

	m := Model{
		sessionMgr:            mgr,
		analyticsManager:      analyticsManager,
		historyManager:        historyManager,
		bookmarkManager:       bookmarkManager,
		keybinds:              keybindRegistry,
		mode:                  ModeNormal,
		version:               version,
		fileExplorer:          fileExplorer,
		historyState:          historyState,
		analyticsState:        analyticsState,
		stressTestState:       stressTestState,
		docState:              docState,
		profileEditState:      profileEditState,
		mockServerState:       mockServerState,
		proxyServerState:      proxyServerState,
		renameState:           renameState,
		showHeaders:           false,
		showBody:              true,
		fullscreen:            false,
		focusedPanel:            "sidebar", // Start with sidebar focused
		streamState:             &StreamState{},
		requestState:            &RequestState{},
		wsState:                 &WebSocketState{},
		responseView:            viewport.New(80, 20),
		modalView:               viewport.New(80, 20), // For scrollable modals
		wsHistoryView:           viewport.New(80, 20), // Left pane: message history
		wsMessageMenuView:       viewport.New(80, 20), // Right pane: predefined messages
		wsFocusedPane:           "menu",               // Start with menu focused
	}

	// Load requests from first file
	if len(files) > 0 {
		m.loadRequestsFromCurrentFile()
	}

	return m, nil
}

// Run starts the TUI
func Run(version string) error {
	// Initialize config
	if err := config.Initialize(); err != nil {
		return err
	}

	// Load session
	mgr := session.NewManager()
	if err := mgr.Load(); err != nil {
		return err
	}

	// Create model
	m, err := New(mgr, version)
	if err != nil {
		return err
	}

	// Start TUI (pass pointer since Update uses pointer receiver)
	// Enable mouse cell motion to capture scroll events (which we'll discard to prevent terminal scrolling)
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// loadFiles loads all .http, .yaml, .yml, .json, .jsonc files from the working directory
func loadFiles(mgr *session.Manager) ([]types.FileInfo, error) {
	profile := mgr.GetActiveProfile()
	workdir, err := config.GetWorkingDirectory(profile.Workdir)
	if err != nil {
		return nil, err
	}

	var files []types.FileInfo

	// Walk the directory
	err = filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files that cause errors
		}

		if info.IsDir() {
			// Skip only hidden directories and common dependency/build directories
			// that are definitely not API paths (only skip at root level to avoid false positives)
			dirName := filepath.Base(path)
			relPath, _ := filepath.Rel(workdir, path)
			depth := strings.Count(relPath, string(filepath.Separator))

			// Only skip these at depth 0 or 1 (at or near root level)
			if depth <= 1 {
				if dirName == ".git" || dirName == "node_modules" || dirName == ".venv" ||
					dirName == ".idea" || dirName == ".vscode" || dirName == "__pycache__" {
					return filepath.SkipDir
				}
			}

			// Always skip hidden directories (except root ".")
			if strings.HasPrefix(dirName, ".") && path != workdir {
				return filepath.SkipDir
			}

			return nil // Continue into other directories
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".http" || ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".jsonc" || ext == ".ws" {
			relPath, _ := filepath.Rel(workdir, path)

			// Parse file to get first HTTP method and tags
			httpMethod := ""
			tags := []string{}

			// WebSocket files are handled differently
			if ext == ".ws" {
				// Use "WS" as the method indicator for WebSocket files
				httpMethod = "WS"

				// Try to parse WebSocket file for tags
				if wsReq, err := parser.ParseWebSocketFile(path); err == nil && wsReq.Documentation != nil {
					for _, tag := range wsReq.Documentation.Tags {
						tags = append(tags, tag)
					}
				}
			} else {
				// Regular HTTP request files
				if requests, err := parser.Parse(path); err == nil && len(requests) > 0 {
					httpMethod = requests[0].Method

					// Collect unique tags from all requests in file
					tagSet := make(map[string]bool)
					for _, req := range requests {
						// Ensure documentation is parsed from DocumentationLines
						req.EnsureDocumentationParsed(parser.ParseDocumentationLines)

						if req.Documentation != nil {
							for _, tag := range req.Documentation.Tags {
								tagSet[tag] = true
							}
						}
					}
					for tag := range tagSet {
						tags = append(tags, tag)
					}
				}
			}

			files = append(files, types.FileInfo{
				Path:         path,
				Name:         relPath,
				RequestCount: 0, // TODO: Count requests in file
				ModifiedTime: info.ModTime(),
				HTTPMethod:   httpMethod,
				Tags:         tags,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// SetProxy sets the proxy server for the model
func (m *Model) SetProxy(p *proxy.Proxy) {
	if p != nil {
		m.proxyServerState.Start(p)
	} else {
		m.proxyServerState.Stop()
	}
}

type tickMsg time.Time
