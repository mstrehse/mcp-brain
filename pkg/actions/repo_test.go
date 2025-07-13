package actions

import (
	"testing"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/knowledge"
)

// TestRepositoryInterface verifies that the FileRepository correctly implements the interface
func TestRepositoryInterface(t *testing.T) {
	baseDir := t.TempDir()

	fileRepo, err := knowledge.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create file repository: %v", err)
	}

	if err := fileRepo.Close(); err != nil {
		t.Fatalf("Failed to close repository: %v", err)
	}

	// Test interface compliance
	var repo contracts.KnowledgeRepository = fileRepo
	// If this compiles, it means the interface is implemented correctly
	_ = repo
}

// TestNewRepositories verifies that the new dependency injection pattern works correctly
func TestNewRepositories(t *testing.T) {
	// Test that initialization works correctly with the new pattern
	baseDir := t.TempDir()
	repositories, err := NewRepositories(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repositories: %v", err)
	}

	// Clean up repositories after test
	defer func() {
		if err := repositories.Close(); err != nil {
			t.Logf("Warning: Failed to close repositories: %v", err)
		}
	}()

	if repositories.Knowledge == nil {
		t.Fatal("Knowledge repository should not be nil after initialization")
	}

	if repositories.Task == nil {
		t.Fatal("Task repository should not be nil after initialization")
	}

	// Test that the knowledge repository is actually usable
	path := "init.md"
	content := "Initialization test"

	if err := repositories.Knowledge.Write(path, content); err != nil {
		t.Fatalf("Write failed with initialized repo: %v", err)
	}

	readContent, err := repositories.Knowledge.Read(path)
	if err != nil {
		t.Fatalf("Read failed with initialized repo: %v", err)
	}
	if readContent != content {
		t.Errorf("Content mismatch: got %q, want %q", readContent, content)
	}

	// Test that the task repository is usable
	tasks, err := repositories.Task.AddTasks([]string{"test task 1", "test task 2"})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Test retrieving a task
	task, err := repositories.Task.GetTask()
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if task == nil {
		t.Fatal("Task should not be nil")
	}
	if task.Content != "test task 1" {
		t.Errorf("Expected task content 'test task 1', got %q", task.Content)
	}
}
