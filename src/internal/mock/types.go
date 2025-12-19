package mock

import "time"

// Config represents the mock server configuration
type Config struct {
	Port    int     `json:"port" yaml:"port"`       // Server port (default: 8080)
	Host    string  `json:"host" yaml:"host"`       // Server host (default: localhost)
	Routes  []Route `json:"routes" yaml:"routes"`   // Route definitions
	Logging bool    `json:"logging" yaml:"logging"` // Enable request logging (default: true)
}

// Route represents a mock route configuration
type Route struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`               // Route description
	Method      string            `json:"method" yaml:"method"`                               // HTTP method (GET, POST, etc.)
	Path        string            `json:"path" yaml:"path"`                                   // URL path pattern
	PathType    string            `json:"pathType,omitempty" yaml:"pathType,omitempty"`       // exact, prefix, regex (default: exact)
	Status      int               `json:"status" yaml:"status"`                               // HTTP status code
	Headers     map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`         // Response headers
	Body        string            `json:"body,omitempty" yaml:"body,omitempty"`               // Response body (string or file path)
	BodyFile    string            `json:"bodyFile,omitempty" yaml:"bodyFile,omitempty"`       // Path to response body file
	Delay       int               `json:"delay,omitempty" yaml:"delay,omitempty"`             // Response delay in milliseconds
	Description string            `json:"description,omitempty" yaml:"description,omitempty"` // Route documentation
}

// RequestLog represents a logged request
type RequestLog struct {
	Timestamp   time.Time         `json:"timestamp"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	MatchedRule string            `json:"matchedRule"`
	Status      int               `json:"status"`
	Duration    time.Duration     `json:"duration"`
}
