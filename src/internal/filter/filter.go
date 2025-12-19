package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jmespath/go-jmespath"
	"github.com/studiowebux/restcli/internal/types"
)

const (
	// QueryShellTimeout is the maximum time allowed for query shell command execution
	QueryShellTimeout = 30 * time.Second
)

var (
	// Shell command pattern: $(command)
	shellPattern = regexp.MustCompile(`^\$\((.+)\)$`)
)

// Apply applies filter and query expressions to a response body
// Filter narrows results (e.g., items[?status==`active`])
// Query transforms/selects fields (e.g., [].name)
// If query starts with $(...), it's executed as a bash command with body piped to stdin
func Apply(body string, filter string, query string) (string, error) {
	result := body

	// Apply filter first (if specified)
	if filter != "" {
		filtered, err := applyJMESPath(result, filter)
		if err != nil {
			return "", fmt.Errorf("failed to apply filter: %w", err)
		}
		result = filtered
	}

	// Apply query (can be JMESPath or shell command)
	if query != "" {
		// Check if it's a shell command
		if matches := shellPattern.FindStringSubmatch(query); len(matches) > 1 {
			command := matches[1]
			queried, err := executeShellCommand(result, command)
			if err != nil {
				return "", fmt.Errorf("failed to execute query shell command: %w", err)
			}
			result = queried
		} else {
			// Apply as JMESPath query
			queried, err := applyJMESPath(result, query)
			if err != nil {
				return "", fmt.Errorf("failed to apply query: %w", err)
			}
			result = queried
		}
	}

	return result, nil
}

// applyJMESPath applies a JMESPath expression to a JSON string
func applyJMESPath(jsonStr string, expression string) (string, error) {
	// Parse the JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Compile the JMESPath expression
	jp, err := jmespath.Compile(expression)
	if err != nil {
		return "", fmt.Errorf("invalid JMESPath expression '%s': %w", expression, err)
	}

	// Search/apply the expression
	result, err := jp.Search(data)
	if err != nil {
		return "", fmt.Errorf("JMESPath search failed: %w", err)
	}

	// Handle null result
	if result == nil {
		return "null", nil
	}

	// Convert result back to JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// executeShellCommand executes a shell command with the body piped to stdin
func executeShellCommand(body string, command string) (string, error) {
	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), QueryShellTimeout)
	defer cancel()

	// Use sh -c to execute the command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	// Pipe body to stdin
	cmd.Stdin = strings.NewReader(body)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := err.Error()
		if stderr.Len() > 0 {
			errMsg = strings.TrimSpace(stderr.String())
		}
		return "", fmt.Errorf("command '%s' failed: %s", command, errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// IsValidJMESPath checks if an expression is valid JMESPath syntax
func IsValidJMESPath(expression string) bool {
	_, err := jmespath.Compile(expression)
	return err == nil
}

// IsShellCommand checks if a query is a shell command (starts with $(...))
func IsShellCommand(query string) bool {
	return shellPattern.MatchString(query)
}

// FilterByTags filters files that contain ANY of the specified tags
func FilterByTags(files []types.FileInfo, tags []string) []types.FileInfo {
	if len(tags) == 0 {
		return files
	}

	var filtered []types.FileInfo
	for _, file := range files {
		if hasAnyTag(file.Tags, tags) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// FilterByCollection filters files based on collection criteria
func FilterByCollection(files []types.FileInfo, collection *types.Collection) []types.FileInfo {
	if collection == nil {
		return files
	}

	var filtered []types.FileInfo
	for _, file := range files {
		// Check tag filter
		if len(collection.TagFilter) > 0 && !hasAnyTag(file.Tags, collection.TagFilter) {
			continue
		}

		// Check path pattern
		if collection.PathPattern != "" {
			matched, err := filepath.Match(collection.PathPattern, file.Path)
			if err != nil || !matched {
				continue
			}
		}

		filtered = append(filtered, file)
	}
	return filtered
}

// GetAllTags extracts all unique tags from a list of files
func GetAllTags(files []types.FileInfo) []string {
	tagSet := make(map[string]bool)
	for _, file := range files {
		for _, tag := range file.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

// hasAnyTag checks if the file has any of the specified tags
func hasAnyTag(fileTags, filterTags []string) bool {
	for _, filterTag := range filterTags {
		for _, fileTag := range fileTags {
			if strings.EqualFold(fileTag, filterTag) {
				return true
			}
		}
	}
	return false
}
