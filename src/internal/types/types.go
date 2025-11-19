package types

import "time"

// HttpRequest represents an HTTP request definition from .http files
type HttpRequest struct {
	Name                string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Method              string                 `json:"method" yaml:"method"`
	URL                 string                 `json:"url" yaml:"url"`
	Headers             map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body                string                 `json:"body,omitempty" yaml:"body,omitempty"`
	Documentation       *Documentation         `json:"documentation,omitempty" yaml:"documentation,omitempty"`
	DocumentationLines  []string               `json:"-" yaml:"-"` // Raw documentation comment lines for lazy loading
	documentationParsed bool                   `json:"-" yaml:"-"` // Whether documentation has been parsed (unexported for internal use)
}

// EnsureDocumentationParsed parses documentation lines if not already parsed
// This is called on demand when documentation is first accessed
func (r *HttpRequest) EnsureDocumentationParsed(parseFunc func([]string) *Documentation) {
	if r.documentationParsed || len(r.DocumentationLines) == 0 {
		return
	}
	r.Documentation = parseFunc(r.DocumentationLines)
	r.documentationParsed = true
	// Clear the lines to free memory
	r.DocumentationLines = nil
}

// Documentation contains request documentation metadata
type Documentation struct {
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string    `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses   []Response  `json:"responses,omitempty" yaml:"responses,omitempty"`
}

// Parameter represents a request parameter
type Parameter struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Example     string `json:"example,omitempty" yaml:"example,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Response represents an expected API response
type Response struct {
	Code        string          `json:"code" yaml:"code"`
	Description string          `json:"description" yaml:"description"`
	ContentType string          `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Fields      []ResponseField `json:"fields,omitempty" yaml:"fields,omitempty"`
	Example     string          `json:"example,omitempty" yaml:"example,omitempty"`
}

// ResponseField represents a field in a response schema
type ResponseField struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Session represents ephemeral session state
type Session struct {
	Variables      map[string]string `json:"variables,omitempty"`
	ActiveProfile  string            `json:"activeProfile,omitempty"`
	HistoryEnabled *bool             `json:"historyEnabled,omitempty"`
}

// Profile represents a header/variable profile
type Profile struct {
	Name      string                    `json:"name"`
	Headers   map[string]string         `json:"headers,omitempty"`
	Variables map[string]VariableValue  `json:"variables,omitempty"`
	Workdir   string                    `json:"workdir,omitempty"`
	OAuth     *OAuthConfig              `json:"oauth,omitempty"`
	Editor    string                    `json:"editor,omitempty"`
	Output    string                    `json:"output,omitempty"` // json, yaml, text
}

// VariableValue can be a simple string or a multi-value variable
type VariableValue struct {
	// Simple string value
	StringValue *string

	// Multi-value variable
	MultiValue *MultiValueVariable
}

// MultiValueVariable represents a variable with multiple options
type MultiValueVariable struct {
	Options     []string `json:"options"`
	Active      int      `json:"active"`
	Description string   `json:"description,omitempty"`
}

// OAuthConfig contains OAuth 2.0 configuration
type OAuthConfig struct {
	Enabled bool `json:"enabled"`

	// Manual mode - complete auth URL
	AuthEndpoint string `json:"authEndpoint,omitempty"`

	// Auto-build mode
	AuthURL          string `json:"authUrl,omitempty"`
	TokenURL         string `json:"tokenUrl,omitempty"`
	ClientID         string `json:"clientId,omitempty"`
	ClientSecret     string `json:"clientSecret,omitempty"`
	RedirectURI      string `json:"redirectUri,omitempty"`
	Scope            string `json:"scope,omitempty"`
	ResponseType     string `json:"responseType,omitempty"` // code or token
	WebhookPort      int    `json:"webhookPort,omitempty"`
	TokenStorageKey  string `json:"tokenStorageKey,omitempty"`
}

// RequestResult contains the HTTP response data
type RequestResult struct {
	Status         int               `json:"status"`
	StatusText     string            `json:"statusText"`
	Headers        map[string]string `json:"headers"`
	Body           string            `json:"body"`
	Duration       int64             `json:"duration"`       // milliseconds
	RequestSize    int               `json:"requestSize"`    // bytes
	ResponseSize   int               `json:"responseSize"`   // bytes
	Error          string            `json:"error,omitempty"`
}

// HistoryEntry represents a saved request/response pair
type HistoryEntry struct {
	Timestamp          string            `json:"timestamp"`
	RequestFile        string            `json:"requestFile"`
	RequestName        string            `json:"requestName,omitempty"`
	Method             string            `json:"method"`
	URL                string            `json:"url"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body,omitempty"`
	ResponseStatus     int               `json:"responseStatus"`
	ResponseStatusText string            `json:"responseStatusText"`
	ResponseHeaders    map[string]string `json:"responseHeaders"`
	ResponseBody       string            `json:"responseBody"`
	Duration           int64             `json:"duration"`
	RequestSize        int               `json:"requestSize,omitempty"`
	ResponseSize       int               `json:"responseSize,omitempty"`
	Error              string            `json:"error,omitempty"`
}

// RequestFile represents a parsed .http file
type RequestFile struct {
	Path     string
	Requests []HttpRequest
}

// FileInfo represents a file in the TUI file list
type FileInfo struct {
	Path          string
	Name          string
	RequestCount  int
	ModifiedTime  time.Time
}
