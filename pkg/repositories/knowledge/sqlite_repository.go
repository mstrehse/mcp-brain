package knowledge

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	_ "modernc.org/sqlite"
)

// SqliteRepository handles SQLite-based storage for knowledge
type SqliteRepository struct {
	db *sql.DB
}

// NewSqliteRepository creates a new SQLite repository with the given database file path
func NewSqliteRepository(dbPath string) (*SqliteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	repo := &SqliteRepository{db: db}
	if err := repo.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return repo, nil
}

// Close closes the database connection
func (r *SqliteRepository) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

// createTables creates the necessary tables for storing knowledge
func (r *SqliteRepository) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS knowledge (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		path TEXT NOT NULL,
		content TEXT,
		is_directory BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(project, path)
	);
	
	CREATE INDEX IF NOT EXISTS idx_project ON knowledge(project);
	CREATE INDEX IF NOT EXISTS idx_project_path ON knowledge(project, path);
	`

	_, err := r.db.Exec(query)
	return err
}

// List returns a json representation of the directory and file structure within the project's knowledge
func (r *SqliteRepository) List(project string) (contracts.DirStructure, error) {
	query := `SELECT path, is_directory FROM knowledge WHERE project = ? ORDER BY path`
	rows, err := r.db.Query(query, project)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override the main return error
			fmt.Printf("Error closing rows: %v\n", err)
		}
	}()

	result := contracts.DirStructure{}

	for rows.Next() {
		var path string
		var isDirectory bool

		if err := rows.Scan(&path, &isDirectory); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		r.insertPathIntoStructure(result, path, isDirectory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// insertPathIntoStructure inserts a path into the directory structure
func (r *SqliteRepository) insertPathIntoStructure(structure contracts.DirStructure, path string, isDirectory bool) {
	parts := strings.Split(path, "/")

	// Build the structure level by level
	current := structure
	for i, part := range parts {
		if i == len(parts)-1 {
			// This is the final element
			if isDirectory {
				if current[part] == nil {
					current[part] = make(contracts.DirStructure)
				}
			} else {
				current[part] = nil
			}
		} else {
			// This is a directory in the path
			if current[part] == nil {
				current[part] = make(contracts.DirStructure)
			}
			// Use interface{} to work around the type system
			var nextMap interface{} = current[part]
			current = nextMap.(contracts.DirStructure)
		}
	}
}

// Write knowledge to the database
func (r *SqliteRepository) Write(project string, path string, content string) error {
	// First, ensure parent directories exist
	if err := r.ensureParentDirectories(project, path); err != nil {
		return fmt.Errorf("failed to ensure parent directories: %w", err)
	}

	// Insert or update the file
	query := `
	INSERT INTO knowledge (project, path, content, is_directory, updated_at)
	VALUES (?, ?, ?, 0, CURRENT_TIMESTAMP)
	ON CONFLICT(project, path) 
	DO UPDATE SET content = excluded.content, updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Exec(query, project, path, content)
	if err != nil {
		return fmt.Errorf("failed to write knowledge: %w", err)
	}

	return nil
}

// ensureParentDirectories ensures all parent directories exist in the database
func (r *SqliteRepository) ensureParentDirectories(project string, path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}

	parts := strings.Split(dir, "/")
	currentPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		insertQuery := `
		INSERT INTO knowledge (project, path, is_directory, updated_at)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(project, path) DO NOTHING
		`
		_, err := r.db.Exec(insertQuery, project, currentPath)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", currentPath, err)
		}
	}
	return nil
}

// Read knowledge from the database
func (r *SqliteRepository) Read(project string, path string) (string, error) {
	query := `SELECT content FROM knowledge WHERE project = ? AND path = ? AND is_directory = 0`

	var content string
	err := r.db.QueryRow(query, project, path).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("file not found: %s", path)
		}
		return "", fmt.Errorf("failed to read knowledge: %w", err)
	}

	return content, nil
}

// Delete knowledge from the database
func (r *SqliteRepository) Delete(project string, path string) error {
	// Check if the path exists
	var exists bool
	checkQuery := `SELECT 1 FROM knowledge WHERE project = ? AND path = ?`
	err := r.db.QueryRow(checkQuery, project, path).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Delete the file/directory and all its children
	deleteQuery := `DELETE FROM knowledge WHERE project = ? AND (path = ? OR path LIKE ?)`
	pathPrefix := path + "/%"

	_, err = r.db.Exec(deleteQuery, project, path, pathPrefix)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge: %w", err)
	}

	return nil
}
