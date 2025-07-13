package actions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/repositories/task"
)

func TestTasksAddHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := task.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTasksAddHandler(repo)

	t.Run("successful add tasks", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tasks-add",
				Arguments: map[string]interface{}{
					"contents": []string{"Task 1", "Task 2", "Task 3"},
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

		// Verify the result is valid JSON
		var taskResult map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &taskResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Check that tasks_added field exists
		if tasksAdded, ok := taskResult["tasks_added"]; ok {
			if tasksAddedFloat, ok := tasksAdded.(float64); ok {
				if int(tasksAddedFloat) != 3 {
					t.Errorf("Expected 3 tasks added, got %d", int(tasksAddedFloat))
				}
			} else {
				t.Error("tasks_added should be a number")
			}
		} else {
			t.Error("Expected tasks_added field in result")
		}
	})

	t.Run("missing contents parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "tasks-add",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing contents")
		}
	})

	t.Run("empty contents array", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tasks-add",
				Arguments: map[string]interface{}{
					"contents": []string{},
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for empty contents")
		}
	})
}

func TestTaskGetHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := task.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTaskGetHandler(repo)

	// Setup test data
	tasks, err := repo.AddTasks([]string{"Test task 1", "Test task 2"})
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}

	t.Run("successful get task", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-get",
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
		var taskResult interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &taskResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Verify it's a task object (should have fields like content)
		if taskMap, ok := taskResult.(map[string]interface{}); ok {
			if content, hasContent := taskMap["content"]; hasContent {
				if contentStr, ok := content.(string); ok {
					if contentStr != "Test task 1" {
						t.Errorf("Expected task content 'Test task 1', got %s", contentStr)
					}
				} else {
					t.Error("Task content should be a string")
				}
			} else {
				t.Error("Expected content field in task")
			}
		} else {
			t.Error("Expected task result to be an object")
		}
	})

	t.Run("get task when queue is empty", func(t *testing.T) {
		// Get all remaining tasks to empty the queue
		for {
			task, err := repo.GetTask()
			if err != nil || task == nil {
				break
			}
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Should return null for empty queue
		if textContent.Text != "null" {
			t.Errorf("Expected null for empty queue, got: %s", textContent.Text)
		}
	})
}
