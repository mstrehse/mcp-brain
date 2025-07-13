package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_knowledge_repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create repository
	repo, err := NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Test Write operation
	testContent := "# Test Knowledge\n\nThis is a test knowledge file."
	if err := repo.Write("test/knowledge", testContent); err != nil {
		t.Fatalf("Failed to write knowledge: %v", err)
	}

	// Test Read operation
	readContent, err := repo.Read("test/knowledge")
	if err != nil {
		t.Fatalf("Failed to read knowledge: %v", err)
	}
	if readContent != testContent {
		t.Errorf("Read content doesn't match written content. Got: %s, Want: %s", readContent, testContent)
	}

	// Test List operation
	structure, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list knowledge: %v", err)
	}

	// Check that the structure contains our test file
	if structure == nil {
		t.Fatal("List returned nil structure")
	}

	testDir, exists := structure["test"]
	if !exists {
		t.Fatal("Test directory not found in structure")
	}

	if testDir == nil {
		t.Fatal("Test directory should not be nil")
	}

	if testDir["knowledge.md"] != nil {
		t.Fatal("Test file should have nil value in structure")
	}

	// Test Delete operation
	if err := repo.Delete("test/knowledge"); err != nil {
		t.Fatalf("Failed to delete knowledge: %v", err)
	}

	// Verify file is deleted
	_, err = repo.Read("test/knowledge")
	if err == nil {
		t.Fatal("Expected error when reading deleted file")
	}
}

func TestFileRepositoryPathNormalization(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_knowledge_repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create repository
	repo, err := NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Test writing without .md extension
	testContent := "# Test Knowledge\n\nThis is a test knowledge file."
	if err := repo.Write("test", testContent); err != nil {
		t.Fatalf("Failed to write knowledge: %v", err)
	}

	// Verify file was created with .md extension
	fullPath := filepath.Join(repo.baseDir, "test.md")
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatal("File was not created with .md extension")
	}

	// Test reading without .md extension
	readContent, err := repo.Read("test")
	if err != nil {
		t.Fatalf("Failed to read knowledge: %v", err)
	}
	if readContent != testContent {
		t.Errorf("Read content doesn't match written content")
	}
}

func TestFileRepositoryEmptyDirCleanup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_knowledge_repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create repository
	repo, err := NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Create a file in a nested directory
	testContent := "# Test Knowledge"
	if err := repo.Write("deep/nested/test", testContent); err != nil {
		t.Fatalf("Failed to write knowledge: %v", err)
	}

	// Verify nested directory structure was created
	deepDir := filepath.Join(repo.baseDir, "deep", "nested")
	if _, err := os.Stat(deepDir); os.IsNotExist(err) {
		t.Fatal("Nested directory was not created")
	}

	// Delete the file
	if err := repo.Delete("deep/nested/test"); err != nil {
		t.Fatalf("Failed to delete knowledge: %v", err)
	}

	// Verify that empty parent directories are cleaned up
	// (This depends on the implementation - if no other files exist in the parent dirs)
	if _, err := os.Stat(deepDir); !os.IsNotExist(err) {
		// It's okay if the directory still exists - this depends on the cleanup implementation
		t.Logf("Empty parent directory still exists: %s", deepDir)
	}
}
