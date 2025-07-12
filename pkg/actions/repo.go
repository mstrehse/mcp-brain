package actions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/knowledge"
	"github.com/mstrehse/mcp-brain/pkg/repositories/task"
)

// Repositories holds all repository instances
type Repositories struct {
	Knowledge contracts.KnowledgeRepository
	Task      contracts.TaskRepository
}

// NewRepositories creates and initializes all repositories with the provided base directory
func NewRepositories(baseDir string) (*Repositories, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", baseDir, err)
	}

	// Place the SQLite database inside the .mcp-brain folder
	dbPath := filepath.Join(baseDir, "brain.db")

	knowledgeRepo, err := knowledge.NewSqliteRepository(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite knowledge repository: %w", err)
	}

	taskRepo, err := task.NewSqliteRepository(dbPath)
	if err != nil {
		// Close the knowledge repo if task repo fails
		_ = knowledgeRepo.Close()
		return nil, fmt.Errorf("failed to initialize task repository: %w", err)
	}

	return &Repositories{
		Knowledge: knowledgeRepo,
		Task:      taskRepo,
	}, nil
}

// Close closes all repositories and cleans up resources
func (r *Repositories) Close() error {
	var knowledgeErr, taskErr error

	if repo, ok := r.Knowledge.(*knowledge.SqliteRepository); ok {
		knowledgeErr = repo.Close()
	}

	if repo, ok := r.Task.(*task.SqliteRepository); ok {
		taskErr = repo.Close()
	}

	// Return the first error encountered
	if knowledgeErr != nil {
		return knowledgeErr
	}
	return taskErr
}
