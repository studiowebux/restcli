/*
Package tui implements the terminal user interface for RestCLI.

# Architecture

The TUI follows the Bubble Tea framework's Model-Update-View pattern:
  - Model: Maintains all application state
  - Update: Processes messages and returns commands
  - View: Renders the current state to the terminal

# Key Components

  - model.go: Core state and initialization, defines the Model struct
  - keys.go: Keyboard input handling and keybind routing
  - render.go: View rendering logic for the main interface
  - actions.go: Business logic and side effects (HTTP requests, file operations)

# State Management

The application is decomposed into focused state objects (see sync_state.go):
  - FileExplorerState: File navigation and filtering
  - HistoryState: Request history tracking
  - AnalyticsState: Analytics data and visualization
  - StressTestState: Stress testing execution
  - DocumentationState: Documentation viewer state
  - ProfileEditState: Profile editing forms
  - StreamState/RequestState/WebSocketState: Thread-safe request lifecycle

All state objects use sync.RWMutex for thread safety.

# Modal System

The application uses a mode-based system with 79+ distinct modes.
Each modal has associated handlers in keys.go and render functions.

Modes are organized by category:
  - Normal operation (ModeNormal, ModeLoading, etc.)
  - Request editing (ModeVariableEdit, ModeHeaderEdit, etc.)
  - Modals (ModeHistory, ModeAnalytics, ModeStressTest, etc.)
  - Confirmations (ModeDeleteConfirm, ModeHistoryClearConfirm, etc.)

# Keybind System

Keybinds are managed through the keybinds.Registry:
  - Context-aware bindings (global, normal, modal-specific)
  - User-customizable via keybinds.json
  - Reserved keys protection
  - Shadowing detection for hierarchical contexts

# Threading Model

The TUI runs in a single goroutine (Bubble Tea's event loop), but spawns
goroutines for:
  - HTTP request execution
  - WebSocket connections
  - Stress test workers
  - File operations

Communication with background goroutines uses channels and tea.Cmd functions.

# Example Usage

	config := &Config{
		WorkDir: "/path/to/requests",
		Profile: profileMgr.GetActiveProfile(),
	}

	model := NewModel(config)
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if err := program.Start(); err != nil {
		log.Fatal(err)
	}
*/
package tui
