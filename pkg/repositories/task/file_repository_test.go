package task

import (
	"os"
	"testing"
	"time"
)

func TestFileRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_task_repo")
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

	// Test AddTasks operation
	contents := []string{"Task 1", "Task 2", "Task 3"}
	tasks, err := repo.AddTasks(contents)
	if err != nil {
		t.Fatalf("Failed to add tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify task contents
	for i, task := range tasks {
		if task.Content != contents[i] {
			t.Errorf("Expected content %s, got %s", contents[i], task.Content)
		}
	}

	// Test GetTask operation (FIFO order)
	task1, err := repo.GetTask()
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task1 == nil {
		t.Fatal("Expected task, got nil")
	}
	if task1.Content != "Task 1" {
		t.Errorf("Expected first task content 'Task 1', got '%s'", task1.Content)
	}

	task2, err := repo.GetTask()
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task2 == nil {
		t.Fatal("Expected task, got nil")
	}
	if task2.Content != "Task 2" {
		t.Errorf("Expected second task content 'Task 2', got '%s'", task2.Content)
	}

	// Test GetTaskCount
	count, err := repo.GetTaskCount()
	if err != nil {
		t.Fatalf("Failed to get task count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 remaining task, got %d", count)
	}

	// Get the last task
	task3, err := repo.GetTask()
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task3 == nil {
		t.Fatal("Expected task, got nil")
	}
	if task3.Content != "Task 3" {
		t.Errorf("Expected third task content 'Task 3', got '%s'", task3.Content)
	}

	// Test empty queue
	emptyTask, err := repo.GetTask()
	if err != nil {
		t.Fatalf("Failed to get task from empty queue: %v", err)
	}
	if emptyTask != nil {
		t.Fatal("Expected nil from empty queue, got task")
	}
}

func TestFileRepositoryEmptyTaskList(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_task_repo")
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

	// Test adding empty task list
	tasks, err := repo.AddTasks([]string{})
	if err != nil {
		t.Fatalf("Failed to add empty task list: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}

	// Test getting from empty queue
	task, err := repo.GetTask()
	if err != nil {
		t.Fatalf("Failed to get task from empty queue: %v", err)
	}
	if task != nil {
		t.Fatal("Expected nil from empty queue, got task")
	}
}

func TestFileRepositoryPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_task_repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create repository and add tasks
	repo1, err := NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	contents := []string{"Persistent Task 1", "Persistent Task 2"}
	_, err = repo1.AddTasks(contents)
	if err != nil {
		t.Fatalf("Failed to add tasks: %v", err)
	}
	_ = repo1.Close()

	// Create new repository instance pointing to same directory
	repo2, err := NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	// Verify tasks persisted
	count, err := repo2.GetTaskCount()
	if err != nil {
		t.Fatalf("Failed to get task count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 persisted tasks, got %d", count)
	}

	// Get first task
	task, err := repo2.GetTask()
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if task == nil {
		t.Fatal("Expected task, got nil")
	}
	if task.Content != "Persistent Task 1" {
		t.Errorf("Expected 'Persistent Task 1', got '%s'", task.Content)
	}
}

func TestFileRepositoryTimestamps(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_task_repo")
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

	// Record time before adding tasks
	beforeTime := time.Now()

	// Add tasks
	contents := []string{"Timestamped Task"}
	tasks, err := repo.AddTasks(contents)
	if err != nil {
		t.Fatalf("Failed to add tasks: %v", err)
	}

	// Record time after adding tasks
	afterTime := time.Now()

	// Verify timestamp is within reasonable range
	task := tasks[0]
	if task.CreatedAt.Before(beforeTime) || task.CreatedAt.After(afterTime) {
		t.Errorf("Task creation time %v is not within expected range %v to %v",
			task.CreatedAt, beforeTime, afterTime)
	}
}
