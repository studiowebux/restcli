package jsonpath

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Bookmark represents a saved JSONPath expression
type Bookmark struct {
	ID         int
	Expression string
	CreatedAt  time.Time
}

// BookmarkManager handles JSONPath bookmark persistence
type BookmarkManager struct {
	db *sql.DB
}

// NewBookmarkManager creates a new bookmark manager
func NewBookmarkManager(dbPath string) (*BookmarkManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &BookmarkManager{db: db}, nil
}

// Save adds a new bookmark (returns nil if already exists)
func (m *BookmarkManager) Save(expression string) (bool, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return false, fmt.Errorf("expression cannot be empty")
	}

	// Check if bookmark already exists
	var exists bool
	err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM jsonpath_bookmarks WHERE expression = ?)", expression).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark: %w", err)
	}

	if exists {
		return false, nil // Already exists, not an error
	}

	// Insert new bookmark
	_, err = m.db.Exec(`
		INSERT INTO jsonpath_bookmarks (expression, created_at)
		VALUES (?, CURRENT_TIMESTAMP)
	`, expression)

	if err != nil {
		return false, fmt.Errorf("failed to save bookmark: %w", err)
	}

	return true, nil
}

// Delete removes a bookmark by ID
func (m *BookmarkManager) Delete(id int) error {
	result, err := m.db.Exec("DELETE FROM jsonpath_bookmarks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}

// List returns all bookmarks ordered by creation date (newest first)
func (m *BookmarkManager) List() ([]Bookmark, error) {
	rows, err := m.db.Query(`
		SELECT id, expression, created_at
		FROM jsonpath_bookmarks
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.Expression, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookmarks: %w", err)
	}

	return bookmarks, nil
}

// Search filters bookmarks by substring match (case-insensitive)
func (m *BookmarkManager) Search(query string) ([]Bookmark, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return m.List()
	}

	rows, err := m.db.Query(`
		SELECT id, expression, created_at
		FROM jsonpath_bookmarks
		WHERE expression LIKE ?
		ORDER BY created_at DESC
	`, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.Expression, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookmarks: %w", err)
	}

	return bookmarks, nil
}

// Close closes the database connection
func (m *BookmarkManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
