package chain

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/types"
)

// Node represents a request in the dependency graph
type Node struct {
	FilePath string
	Request  *types.HttpRequest
	Children []*Node
	Executed bool
}

// Graph represents the dependency graph for request chaining
type Graph struct {
	Nodes     map[string]*Node // Key is file path
	workdir   string
}

// NewGraph creates a new dependency graph
func NewGraph(workdir string) *Graph {
	return &Graph{
		Nodes:   make(map[string]*Node),
		workdir: workdir,
	}
}

// AddRequest adds a request to the graph
func (g *Graph) AddRequest(filePath string, req *types.HttpRequest) error {
	// Normalize file path
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(g.workdir, filePath)
	}
	filePath = filepath.Clean(filePath)

	// Create node if doesn't exist
	if _, exists := g.Nodes[filePath]; !exists {
		g.Nodes[filePath] = &Node{
			FilePath: filePath,
			Request:  req,
			Children: []*Node{},
			Executed: false,
		}
	}

	return nil
}

// BuildGraph builds the dependency graph for a given file
func (g *Graph) BuildGraph(startFile string) error {
	// Normalize start file path
	if !filepath.IsAbs(startFile) {
		startFile = filepath.Join(g.workdir, startFile)
	}
	startFile = filepath.Clean(startFile)

	// Parse the start file
	requests, err := parser.Parse(startFile)
	if err != nil {
		return fmt.Errorf("failed to parse start file %s: %w", startFile, err)
	}

	if len(requests) == 0 {
		return fmt.Errorf("no requests found in %s", startFile)
	}

	// Use first request
	startReq := &requests[0]

	// Add to graph
	if err := g.AddRequest(startFile, startReq); err != nil {
		return err
	}

	// Recursively build dependencies
	return g.buildDependencies(startFile, startReq)
}

// buildDependencies recursively builds the dependency tree
func (g *Graph) buildDependencies(filePath string, req *types.HttpRequest) error {
	if req.DependsOn == nil || len(req.DependsOn) == 0 {
		return nil
	}

	currentNode := g.Nodes[filePath]

	for _, depPath := range req.DependsOn {
		// Normalize dependency path relative to workdir
		if !filepath.IsAbs(depPath) {
			depPath = filepath.Join(g.workdir, depPath)
		}
		depPath = filepath.Clean(depPath)

		// Check for circular dependencies
		if depPath == filePath {
			return fmt.Errorf("circular dependency detected: %s depends on itself", filePath)
		}

		// Parse dependency file
		depRequests, err := parser.Parse(depPath)
		if err != nil {
			return fmt.Errorf("failed to parse dependency %s: %w", depPath, err)
		}

		if len(depRequests) == 0 {
			return fmt.Errorf("no requests found in dependency %s", depPath)
		}

		depReq := &depRequests[0]

		// Add to graph
		if err := g.AddRequest(depPath, depReq); err != nil {
			return err
		}

		// Add as child
		depNode := g.Nodes[depPath]
		currentNode.Children = append(currentNode.Children, depNode)

		// Recursively process dependencies of this dependency
		if err := g.buildDependencies(depPath, depReq); err != nil {
			return err
		}
	}

	return nil
}

// GetExecutionOrder returns the execution order using topological sort
func (g *Graph) GetExecutionOrder(startFile string) ([]string, error) {
	// Normalize start file path
	if !filepath.IsAbs(startFile) {
		startFile = filepath.Join(g.workdir, startFile)
	}
	startFile = filepath.Clean(startFile)

	startNode, exists := g.Nodes[startFile]
	if !exists {
		return nil, fmt.Errorf("start file %s not found in graph", startFile)
	}

	// Reset executed flags
	for _, node := range g.Nodes {
		node.Executed = false
	}

	// Perform DFS to get execution order
	var order []string
	visited := make(map[string]bool)
	var dfs func(*Node) error

	dfs = func(node *Node) error {
		if visited[node.FilePath] {
			return nil
		}

		// Visit children first (dependencies must execute before this node)
		for _, child := range node.Children {
			if err := dfs(child); err != nil {
				return err
			}
		}

		// Add this node after all dependencies
		if !visited[node.FilePath] {
			order = append(order, node.FilePath)
			visited[node.FilePath] = true
		}

		return nil
	}

	if err := dfs(startNode); err != nil {
		return nil, err
	}

	return order, nil
}

// HasDependencies checks if a request has dependencies
func HasDependencies(req *types.HttpRequest) bool {
	return req.DependsOn != nil && len(req.DependsOn) > 0
}

// HasExtractions checks if a request has variable extractions
func HasExtractions(req *types.HttpRequest) bool {
	return req.Extract != nil && len(req.Extract) > 0
}

// FormatDependencyInfo returns a formatted string showing dependency info
func FormatDependencyInfo(req *types.HttpRequest) string {
	var info []string

	if HasDependencies(req) {
		deps := make([]string, len(req.DependsOn))
		for i, dep := range req.DependsOn {
			deps[i] = filepath.Base(dep)
		}
		info = append(info, fmt.Sprintf("Depends: %s", strings.Join(deps, ", ")))
	}

	if HasExtractions(req) {
		extracts := make([]string, 0, len(req.Extract))
		for varName, jmesPath := range req.Extract {
			extracts = append(extracts, fmt.Sprintf("%s=%s", varName, jmesPath))
		}
		info = append(info, fmt.Sprintf("Extract: %s", strings.Join(extracts, ", ")))
	}

	if len(info) == 0 {
		return ""
	}

	return strings.Join(info, " | ")
}
