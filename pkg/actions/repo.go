package actions

import (
	"fmt"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/knowledge"
	"github.com/mstrehse/mcp-brain/pkg/repositories/task"
	"github.com/mstrehse/mcp-brain/pkg/repositories/template"
)

// Repositories holds all repository instances
type Repositories struct {
	Knowledge contracts.KnowledgeRepository
	Task      contracts.TaskRepository
	Template  contracts.TaskTemplateRepository
}

// NewRepositories creates a new instance of Repositories with all dependencies initialized
func NewRepositories(baseDir string) (*Repositories, error) {
	// Create all repositories with file-based storage
	knowledgeRepo, err := knowledge.NewFileRepository(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize knowledge repository: %w", err)
	}

	taskRepo, err := task.NewFileRepository(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize task repository: %w", err)
	}

	templateRepo, err := template.NewFileRepository(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template repository: %w", err)
	}

	return &Repositories{
		Knowledge: knowledgeRepo,
		Task:      taskRepo,
		Template:  templateRepo,
	}, nil
}

// Close closes all repositories and cleans up resources
func (r *Repositories) Close() error {
	// File repositories don't need explicit closing, but we keep the method for interface consistency
	return nil
}
