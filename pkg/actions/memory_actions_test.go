package actions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/repositories/knowledge"
)

func TestMemoryStoreHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := knowledge.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewMemoryStoreHandler(repo)

	t.Run("successful store", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "memory-store",
				Arguments: map[string]interface{}{
					"path":    "test.md",
					"content": "# Test Content\nThis is a test.",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		if textContent.Text != "Memory stored successfully." {
			t.Errorf("Expected success message, got: %s", textContent.Text)
		}

		// Verify the content was actually stored
		stored, err := repo.Read("test.md")
		if err != nil {
			t.Fatalf("Failed to read stored content: %v", err)
		}
		if stored != "# Test Content\nThis is a test." {
			t.Errorf("Content mismatch: got %q", stored)
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "memory-store",
				Arguments: map[string]interface{}{
					"content": "test content",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing path")
		}
	})

	t.Run("missing content parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "memory-store",
				Arguments: map[string]interface{}{
					"path": "test.md",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing content")
		}
	})
}

func TestMemoryGetHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := knowledge.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewMemoryGetHandler(repo)

	// Setup test data
	testPath := "test.md"
	testContent := "# Test Memory\nThis is test content."
	if err := repo.Write(testPath, testContent); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful get", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "memory-get",
				Arguments: map[string]interface{}{
					"path": testPath,
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		if textContent.Text != testContent {
			t.Errorf("Content mismatch: got %q, want %q", textContent.Text, testContent)
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "memory-get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing path")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "memory-get",
				Arguments: map[string]interface{}{
					"path": "non-existent.md",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for non-existent file")
		}
	})
}

func TestMemoriesListHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := knowledge.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewMemoriesListHandler(repo)

	// Setup test data
	if err := repo.Write("file1.md", "content1"); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}
	if err := repo.Write("dir/file2.md", "content2"); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful list", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "memories-list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON
		var dirStructure interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &dirStructure); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}
	})
}

func TestMemoryDeleteHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := knowledge.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewMemoryDeleteHandler(repo)

	// Setup test data
	testPath := "test.md"
	if err := repo.Write(testPath, "test content"); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "memory-delete",
				Arguments: map[string]interface{}{
					"path": testPath,
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		if textContent.Text != "Memory deleted successfully." {
			t.Errorf("Expected success message, got: %s", textContent.Text)
		}

		// Verify the file was actually deleted
		_, err = repo.Read(testPath)
		if err == nil {
			t.Error("Expected file to be deleted")
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "memory-delete",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing path")
		}
	})
}
