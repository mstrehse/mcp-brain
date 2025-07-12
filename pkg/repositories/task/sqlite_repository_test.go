package task

import (
	"path/filepath"
	"testing"
)

func setupTestSqliteRepo(t *testing.T) (*SqliteRepository, string) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	repo, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	return repo, dbPath
}

func TestSqliteRepository_GetTask(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "test-session-2"
	content := "Test task for getting"

	// Add a task
	addedTasks, err := repo.AddTasks(chatSessionID, []string{content})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	// Get the task
	retrievedTask, err := repo.GetTask(chatSessionID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if retrievedTask.ID != addedTasks[0].ID {
		t.Errorf("Expected task ID %d, got %d", addedTasks[0].ID, retrievedTask.ID)
	}
	if retrievedTask.Content != content {
		t.Errorf("Expected content %q, got %q", content, retrievedTask.Content)
	}
	if retrievedTask.Status != "pending" {
		t.Errorf("Expected status 'pending', got %q", retrievedTask.Status)
	}

	// Try to get another task - should fail
	_, err = repo.GetTask(chatSessionID)
	if err == nil {
		t.Error("Expected error when getting task from empty queue")
	}
}

func TestSqliteRepository_GetTask_FIFO(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "test-session-3"

	// Add multiple tasks
	tasks, err := repo.AddTasks(chatSessionID, []string{"Task 1", "Task 2"})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	// Get tasks - should follow FIFO order
	retrievedTask1, err := repo.GetTask(chatSessionID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if retrievedTask1.ID != tasks[0].ID {
		t.Errorf("Expected first task ID %d, got %d", tasks[0].ID, retrievedTask1.ID)
	}

	retrievedTask2, err := repo.GetTask(chatSessionID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if retrievedTask2.ID != tasks[1].ID {
		t.Errorf("Expected second task ID %d, got %d", tasks[1].ID, retrievedTask2.ID)
	}
}

func TestSqliteRepository_ChatSessionIsolation(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSession1 := "session-1"
	chatSession2 := "session-2"

	// Add tasks to different sessions
	_, err := repo.AddTasks(chatSession1, []string{"Session 1 Task"})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}
	_, err = repo.AddTasks(chatSession2, []string{"Session 2 Task"})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	// Get task from session 1 - should get the session 1 task
	task1, err := repo.GetTask(chatSession1)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if task1.Content != "Session 1 Task" {
		t.Errorf("Expected 'Session 1 Task', got %q", task1.Content)
	}

	// Get task from session 2 - should get the session 2 task
	task2, err := repo.GetTask(chatSession2)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if task2.Content != "Session 2 Task" {
		t.Errorf("Expected 'Session 2 Task', got %q", task2.Content)
	}

	// Both sessions should now be empty
	_, err = repo.GetTask(chatSession1)
	if err == nil {
		t.Error("Expected error when getting task from empty session 1")
	}
	_, err = repo.GetTask(chatSession2)
	if err == nil {
		t.Error("Expected error when getting task from empty session 2")
	}
}

func TestSqliteRepository_GetTask_NoTasks(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "empty-session"

	// Try to get task from empty queue
	_, err := repo.GetTask(chatSessionID)
	if err == nil {
		t.Error("Expected error when getting task from empty queue")
	}
}

func TestSqliteRepository_DatabasePersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persistence.db")

	// Create repo and add task
	repo1, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	chatSessionID := "persistent-session"
	content := "Persistent task"

	tasks, err := repo1.AddTasks(chatSessionID, []string{content})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}
	if err := repo1.Close(); err != nil {
		t.Fatalf("Failed to close repository: %v", err)
	}

	// Create new repo instance with same database
	repo2, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	// Task should still exist
	persistedTask, err := repo2.GetTask(chatSessionID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if persistedTask.ID != tasks[0].ID {
		t.Errorf("Expected task ID %d, got %d", tasks[0].ID, persistedTask.ID)
	}
	if persistedTask.Content != content {
		t.Errorf("Expected content %q, got %q", content, persistedTask.Content)
	}
}

func TestSqliteRepository_AddTasks(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "bulk-session-1"
	contents := []string{"Task 1", "Task 2", "Task 3"}

	tasks, err := repo.AddTasks(chatSessionID, contents)
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Check each task
	for i, task := range tasks {
		if task.ID == 0 {
			t.Errorf("Expected task %d ID to be set", i)
		}
		if task.ChatSessionID != chatSessionID {
			t.Errorf("Expected task %d chat session ID %q, got %q", i, chatSessionID, task.ChatSessionID)
		}
		if task.Content != contents[i] {
			t.Errorf("Expected task %d content %q, got %q", i, contents[i], task.Content)
		}
		if task.Status != "pending" {
			t.Errorf("Expected task %d status 'pending', got %q", i, task.Status)
		}
		if task.CreatedAt.IsZero() {
			t.Errorf("Expected task %d created_at to be set", i)
		}
	}
}

func TestSqliteRepository_AddTasks_EmptyArray(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "empty-session"
	contents := []string{}

	tasks, err := repo.AddTasks(chatSessionID, contents)
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestSqliteRepository_AddTasks_SingleTask(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "single-task-session"
	contents := []string{"Single task"}

	tasks, err := repo.AddTasks(chatSessionID, contents)
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Content != "Single task" {
		t.Errorf("Expected content 'Single task', got %q", task.Content)
	}
}

func TestSqliteRepository_AddTasks_FIFO_Order(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "fifo-session"
	contents := []string{"First", "Second", "Third"}

	_, err := repo.AddTasks(chatSessionID, contents)
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	// Retrieve tasks in FIFO order
	for i, expectedContent := range contents {
		task, err := repo.GetTask(chatSessionID)
		if err != nil {
			t.Fatalf("GetTask failed for task %d: %v", i, err)
		}
		if task.Content != expectedContent {
			t.Errorf("Expected task %d content %q, got %q", i, expectedContent, task.Content)
		}
	}

	// Queue should be empty now
	_, err = repo.GetTask(chatSessionID)
	if err == nil {
		t.Error("Expected error when getting task from empty queue")
	}
}

func TestSqliteRepository_AddTasks_MultipleBatches(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	chatSessionID := "mixed-session"

	// Add first batch
	_, err := repo.AddTasks(chatSessionID, []string{"First batch"})
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	// Add second batch
	bulkContents := []string{"Bulk 1", "Bulk 2"}
	_, err = repo.AddTasks(chatSessionID, bulkContents)
	if err != nil {
		t.Fatalf("AddTasks failed: %v", err)
	}

	// Retrieve tasks - should be in order of insertion
	expectedOrder := []string{"First batch", "Bulk 1", "Bulk 2"}
	for i, expectedContent := range expectedOrder {
		task, err := repo.GetTask(chatSessionID)
		if err != nil {
			t.Fatalf("GetTask failed for task %d: %v", i, err)
		}
		if task.Content != expectedContent {
			t.Errorf("Expected task %d content %q, got %q", i, expectedContent, task.Content)
		}
	}
}

func TestSqliteRepository_AddTasks_SessionIsolation(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	session1 := "session-1"
	session2 := "session-2"

	// Add tasks to different sessions
	_, err := repo.AddTasks(session1, []string{"Session 1 Task A", "Session 1 Task B"})
	if err != nil {
		t.Fatalf("AddTasks failed for session 1: %v", err)
	}

	_, err = repo.AddTasks(session2, []string{"Session 2 Task A", "Session 2 Task B"})
	if err != nil {
		t.Fatalf("AddTasks failed for session 2: %v", err)
	}

	// Get tasks from session 1
	task1A, err := repo.GetTask(session1)
	if err != nil {
		t.Fatalf("GetTask failed for session 1: %v", err)
	}
	if task1A.Content != "Session 1 Task A" {
		t.Errorf("Expected 'Session 1 Task A', got %q", task1A.Content)
	}

	// Get tasks from session 2
	task2A, err := repo.GetTask(session2)
	if err != nil {
		t.Fatalf("GetTask failed for session 2: %v", err)
	}
	if task2A.Content != "Session 2 Task A" {
		t.Errorf("Expected 'Session 2 Task A', got %q", task2A.Content)
	}

	// Continue with remaining tasks
	task1B, err := repo.GetTask(session1)
	if err != nil {
		t.Fatalf("GetTask failed for session 1 task B: %v", err)
	}
	if task1B.Content != "Session 1 Task B" {
		t.Errorf("Expected 'Session 1 Task B', got %q", task1B.Content)
	}

	task2B, err := repo.GetTask(session2)
	if err != nil {
		t.Fatalf("GetTask failed for session 2 task B: %v", err)
	}
	if task2B.Content != "Session 2 Task B" {
		t.Errorf("Expected 'Session 2 Task B', got %q", task2B.Content)
	}
}
