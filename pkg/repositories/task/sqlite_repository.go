package task

import (
	"database/sql"
	"fmt"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	_ "modernc.org/sqlite"
)

// SqliteRepository handles SQLite-based storage for tasks
type SqliteRepository struct {
	db *sql.DB
}

// NewSqliteRepository creates a new SQLite repository with the given database file path
func NewSqliteRepository(dbPath string) (*SqliteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	repo := &SqliteRepository{db: db}
	if err := repo.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return repo, nil
}

// Close closes the database connection
func (r *SqliteRepository) Close() error {
	return r.db.Close()
}

// createTables creates the necessary tables for storing tasks
func (r *SqliteRepository) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_session_id TEXT NOT NULL,
		content TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_chat_session ON tasks(chat_session_id);
	CREATE INDEX IF NOT EXISTS idx_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_chat_session_status ON tasks(chat_session_id, status);
	`

	_, err := r.db.Exec(query)
	return err
}

// AddTasks adds multiple tasks to the queue for the given chat session
func (r *SqliteRepository) AddTasks(chatSessionID string, contents []string) ([]*contracts.Task, error) {
	if len(contents) == 0 {
		return []*contracts.Task{}, nil
	}

	// Start a transaction to ensure atomicity
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO tasks (chat_session_id, content, status, created_at)
		VALUES (?, ?, 'pending', CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			// Log error but don't override the main return error
			fmt.Printf("Error closing statement: %v\n", err)
		}
	}()

	// Insert all tasks and collect their IDs
	taskIDs := make([]int64, 0, len(contents))
	for _, content := range contents {
		result, err := stmt.Exec(chatSessionID, content)
		if err != nil {
			return nil, fmt.Errorf("failed to add task: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get task ID: %w", err)
		}
		taskIDs = append(taskIDs, id)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Retrieve all created tasks
	tasks := make([]*contracts.Task, 0, len(taskIDs))
	for _, id := range taskIDs {
		task, err := r.getTaskByID(int(id))
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTask retrieves and removes the next pending task from the queue for the given chat session
func (r *SqliteRepository) GetTask(chatSessionID string) (*contracts.Task, error) {
	// Start a transaction to ensure atomicity
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Find the oldest pending task for this chat session
	query := `
	SELECT id, chat_session_id, content, status, created_at
	FROM tasks
	WHERE chat_session_id = ? AND status = 'pending'
	ORDER BY created_at ASC, id ASC
	LIMIT 1
	`

	var task contracts.Task
	err = tx.QueryRow(query, chatSessionID).Scan(
		&task.ID,
		&task.ChatSessionID,
		&task.Content,
		&task.Status,
		&task.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no pending tasks found for chat session: %s", chatSessionID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update the task status to completed (effectively removing it from the queue)
	updateQuery := `UPDATE tasks SET status = 'completed' WHERE id = ?`
	_, err = tx.Exec(updateQuery, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &task, nil
}

// getTaskByID retrieves a task by its ID
func (r *SqliteRepository) getTaskByID(id int) (*contracts.Task, error) {
	query := `
	SELECT id, chat_session_id, content, status, created_at
	FROM tasks
	WHERE id = ?
	`

	var task contracts.Task
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ChatSessionID,
		&task.Content,
		&task.Status,
		&task.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}

	return &task, nil
}
