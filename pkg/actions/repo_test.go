package actions

import (
	"path/filepath"
	"testing"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/knowledge"
)

// TestRepositoryInterface verifies that the SqliteRepository correctly implements the interface
func TestRepositoryInterface(t *testing.T) {
	// Test that the concrete implementation satisfies the interface
	dbPath := filepath.Join(t.TempDir(), "test.db")
	sqliteRepo, err := knowledge.NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}
	defer func() {
		if err := sqliteRepo.Close(); err != nil {
			t.Logf("Warning: Failed to close repository: %v", err)
		}
	}()

	var repo contracts.KnowledgeRepository = sqliteRepo

	// Test basic operations through the interface
	project := "test-project"
	path := "test.md"
	content := "# Test Content"

	// Test Write
	if err := repo.Write(project, path, content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Test Read
	readContent, err := repo.Read(project, path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if readContent != content {
		t.Errorf("Content mismatch: got %q, want %q", readContent, content)
	}

	// Test List
	list, err := repo.List(project)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if _, ok := list[path]; !ok {
		t.Error("Expected file in list")
	}

	// Test Delete
	if err := repo.Delete(project, path); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
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
	project := "init-test"
	path := "init.md"
	content := "Initialization test"

	if err := repositories.Knowledge.Write(project, path, content); err != nil {
		t.Fatalf("Write failed with initialized repo: %v", err)
	}

	readContent, err := repositories.Knowledge.Read(project, path)
	if err != nil {
		t.Fatalf("Read failed with initialized repo: %v", err)
	}
	if readContent != content {
		t.Errorf("Content mismatch: got %q, want %q", readContent, content)
	}

	// Test that the task repository is usable
	chatSessionID := "test-session"
	tasks, err := repositories.Task.AddTasks(chatSessionID, []string{"test task 1", "test task 2"})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Test retrieving a task
	task, err := repositories.Task.GetTask(chatSessionID)
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
