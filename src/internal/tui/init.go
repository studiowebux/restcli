package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/types"
)

// New creates a new TUI model
func New(mgr *session.Manager) (Model, error) {
	// Load files
	files, err := loadFiles(mgr)
	if err != nil {
		return Model{}, err
	}

	m := Model{
		sessionMgr:   mgr,
		mode:         ModeNormal,
		files:        files,
		fileIndex:    0,
		fileOffset:   0,
		showHeaders:  false,
		showBody:     true,
		fullscreen:   false,
		focusedPanel: "sidebar", // Start with sidebar focused
		docCollapsed: make(map[int]bool),
		responseView: viewport.New(80, 20),
		modalView:    viewport.New(80, 20), // For scrollable modals
	}

	// Load requests from first file
	if len(files) > 0 {
		m.loadRequestsFromCurrentFile()
	}

	return m, nil
}

// Run starts the TUI
func Run() error {
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
	m, err := New(mgr)
	if err != nil {
		return err
	}

	// Start TUI (pass pointer since Update uses pointer receiver)
	// Note: Mouse is disabled by default in bubbletea
	p := tea.NewProgram(&m, tea.WithAltScreen())
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
		if ext == ".http" || ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".jsonc" {
			relPath, _ := filepath.Rel(workdir, path)

			files = append(files, types.FileInfo{
				Path:         path,
				Name:         relPath,
				RequestCount: 0, // TODO: Count requests in file
				ModifiedTime: info.ModTime(),
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

type tickMsg time.Time
