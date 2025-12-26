/*
Package types defines core data structures used throughout RestCLI.

# Overview

The types package provides shared type definitions for:
  - HTTP requests and responses
  - WebSocket requests and connections
  - Configuration (profiles, TLS)
  - Analytics and history
  - Request chaining and dependencies

# Request Types

HttpRequest:
  - Standard HTTP request definition
  - Parsed from .http files
  - Supports variables, headers, body
  - Validation fields for stress testing

WebSocketRequest:
  - WebSocket connection definition
  - Parsed from .ws files
  - Message sequences (send/receive patterns)
  - Subprotocol support

# Response Types

RequestResult:
  - HTTP response data
  - Status, headers, body
  - Duration and size metrics
  - Error information

WebSocketResult:
  - WebSocket session data
  - Message history
  - Connection statistics
  - Disconnect reason

# Configuration

Profile:
  - Environment-specific settings
  - Base URL, headers, timeouts
  - TLS configuration
  - Variables

TLSConfig:
  - Client certificates (mTLS)
  - CA certificates
  - InsecureSkipVerify flag

# Analytics

AnalyticsEntry:
  - Request execution record
  - Timestamp, duration, status
  - Profile and file path
  - Request/response sizes

# Request Chaining

Requests can reference variables from other requests:
  - Execution order determined by dependencies
  - Results passed via variable resolution
  - Circular dependency detection

# Field Tags

All types use JSON and YAML tags for serialization:
  - File persistence (.http, .ws files)
  - Configuration files (profiles.json)
  - Analytics database

The `omitempty` tag is used extensively to keep serialized data clean.

# Validation

HttpRequest includes validation helpers:
  - IsExpectedStatus: Check if status code matches expectations
  - ExpectedBodyExact, ExpectedBodyContains, ExpectedBodyPattern
  - ExpectedBodyFields for JSON validation

# Constants

Common constants:
  - Default timeouts
  - Buffer sizes
  - HTTP methods
  - WebSocket message types

# Documentation Types

Documentation:
  - Request documentation metadata
  - Parsed from comments in .http files
  - Description, tags, examples
  - Field-level documentation

# Example Structures

HTTP Request:
	{
	  "name": "Create User",
	  "method": "POST",
	  "url": "{{baseUrl}}/users",
	  "headers": {
	    "Content-Type": "application/json"
	  },
	  "body": "{\"name\":\"John\"}"
	}

Profile:
	{
	  "name": "production",
	  "baseUrl": "https://api.example.com",
	  "headers": {
	    "Authorization": "Bearer {{token}}"
	  },
	  "variables": {
	    "token": "secret"
	  },
	  "requestTimeout": "30s"
	}

# Type Safety

All types are designed to be:
  - Immutable after parsing (except for runtime state)
  - Safe for concurrent reads
  - Serializable to/from JSON and YAML

# Extension

When adding new fields:
  - Use JSON and YAML tags
  - Add omitempty for optional fields
  - Update parser if needed
  - Document in struct comments
  - Update validation if applicable
*/
package types
