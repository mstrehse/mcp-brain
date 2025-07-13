package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// FileRepository handles file-based storage for knowledge using markdown files
type FileRepository struct {
	baseDir string
}

// NewFileRepository creates a new file-based repository
func NewFileRepository(baseDir string) (*FileRepository, error) {
	knowledgeDir := filepath.Join(baseDir, "knowledge")

	// Ensure the knowledge directory exists
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create knowledge directory: %w", err)
	}

	return &FileRepository{
		baseDir: knowledgeDir,
	}, nil
}

// Close is a no-op for file-based storage
func (r *FileRepository) Close() error {
	return nil
}

// List returns a json representation of the directory and file structure
func (r *FileRepository) List() (contracts.DirStructure, error) {
	result := contracts.DirStructure{}

	err := filepath.Walk(r.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == r.baseDir {
			return nil
		}

		// Get relative path from base directory
		relPath, err := filepath.Rel(r.baseDir, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)

		r.insertPathIntoStructure(result, relPath, info.IsDir())
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return result, nil
}

// insertPathIntoStructure inserts a path into the directory structure
func (r *FileRepository) insertPathIntoStructure(structure contracts.DirStructure, path string, isDirectory bool) {
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
			// This is an intermediate directory
			if current[part] == nil {
				current[part] = make(contracts.DirStructure)
			}
			current = current[part]
		}
	}
}

// Write knowledge to the filesystem
func (r *FileRepository) Write(path string, content string) error {
	// Ensure the path uses forward slashes and add .md extension if not present
	normalizedPath := filepath.ToSlash(path)
	if !strings.HasSuffix(normalizedPath, ".md") {
		normalizedPath += ".md"
	}

	fullPath := filepath.Join(r.baseDir, normalizedPath)

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Read knowledge from the filesystem
func (r *FileRepository) Read(path string) (string, error) {
	// Normalize path and add .md extension if not present
	normalizedPath := filepath.ToSlash(path)
	if !strings.HasSuffix(normalizedPath, ".md") {
		normalizedPath += ".md"
	}

	fullPath := filepath.Join(r.baseDir, normalizedPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("knowledge file not found: %s", path)
		}
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// Delete knowledge from the filesystem
func (r *FileRepository) Delete(path string) error {
	// Normalize path and add .md extension if not present
	normalizedPath := filepath.ToSlash(path)
	if !strings.HasSuffix(normalizedPath, ".md") {
		normalizedPath += ".md"
	}

	fullPath := filepath.Join(r.baseDir, normalizedPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("knowledge file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty parent directories
	r.removeEmptyParentDirs(filepath.Dir(fullPath))

	return nil
}

// removeEmptyParentDirs removes empty parent directories up to but not including the base directory
func (r *FileRepository) removeEmptyParentDirs(dir string) {
	for dir != r.baseDir && dir != filepath.Dir(dir) {
		if err := os.Remove(dir); err != nil {
			// Stop if directory is not empty or other error
			break
		}
		dir = filepath.Dir(dir)
	}
}
