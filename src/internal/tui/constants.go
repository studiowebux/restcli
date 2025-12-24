package tui

// UI Layout Constants
// These constants define spacing, margins, and dimensions for the TUI layout

const (
	// Modal Dimensions - Standard margins for modal dialogs
	ModalWidthMargin       = 6  // Standard horizontal margin (m.width - 6)
	ModalHeightMargin      = 3  // Standard vertical margin (m.height - 3)
	ModalWidthMarginNarrow = 10 // Narrow horizontal margin for focused modals (m.width - 10)
	ModalHeightMarginSmall = 2  // Small vertical margin (m.height - 2)
	ModalHeightMarginMed   = 4  // Medium vertical margin (m.height - 4)

	// Viewport Padding and Borders
	ViewportBorderWidth      = 2  // Width consumed by borders
	ViewportPaddingHorizontal = 4  // Horizontal padding (left + right)
	ViewportPaddingVertical   = 2  // Vertical padding (top + bottom)

	// Content Area Offsets
	// These are calculated offsets used in viewport sizing
	ContentOffsetStandard = 7  // m.height - 7 for standard viewports
	ContentOffsetLarge    = 9  // m.height - 9 for modals with footers
	ContentOffsetHelp     = 10 // m.height - 10 for help viewer
	ContentOffsetSidebar  = 9  // m.height - 9 for file sidebar
	MainViewHeightOffset  = 5  // m.height - 5 for main render (status + borders + top visibility)

	// Layout Margins
	MinimalBorderMargin   = 2  // m.width - 2 or m.height - 2 for minimal borders
	HelpViewWidthOffset   = 14 // m.width - 14 for help viewport width

	// Modal Content Calculations
	ModalOverheadLines     = 6 // Title (2) + padding (2) + border (2)
	ModalOverheadMinimal   = 4 // Border + title for minimal modals
	ModalFooterLines       = 2 // Footer + blank line

	// Buffer Sizes
	WebSocketMessageBuffer = 100 // Buffer size for WebSocket message channel
	WebSocketSendBuffer    = 10  // Buffer size for WebSocket send channel
	StreamMessageBuffer    = 100 // Buffer size for streaming response channel

	// Split View Ratios
	SplitViewEqual = 0.5 // Equal 50/50 split for split-pane modals

	// Split Pane Layout
	SplitPaneBorderWidth = 3 // Border width between split panes

	// WebSocket Modal Layout
	WebSocketHistoryWidthRatio = 0.6 // 60% for history pane, 40% for menu pane
	WebSocketPaneHeightOffset  = 3   // Breathing room for header, status, footer
)
