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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_chat_session ON tasks(chat_session_id);
	CREATE INDEX IF NOT EXISTS idx_created_at ON tasks(created_at);
	CREATE INDEX IF NOT EXISTS idx_chat_session_created_at ON tasks(chat_session_id, created_at);
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
		INSERT INTO tasks (chat_session_id, content, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
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
	SELECT id, chat_session_id, content, created_at
	FROM tasks
	WHERE chat_session_id = ?
	ORDER BY created_at ASC, id ASC
	LIMIT 1
	`

	var task contracts.Task
	err = tx.QueryRow(query, chatSessionID).Scan(
		&task.ID,
		&task.ChatSessionID,
		&task.Content,
		&task.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no pending tasks found for chat session: %s", chatSessionID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Delete the task from the database (actually removing it from the queue)
	deleteQuery := `DELETE FROM tasks WHERE id = ?`
	_, err = tx.Exec(deleteQuery, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete task: %w", err)
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
	SELECT id, chat_session_id, content, created_at
	FROM tasks
	WHERE id = ?
	`

	var task contracts.Task
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ChatSessionID,
		&task.Content,
		&task.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}

	return &task, nil
}

// ClearTasksForSession removes all tasks for a specific chat session
// This is useful for clearing the task queue when starting a new session
func (r *SqliteRepository) ClearTasksForSession(chatSessionID string) error {
	query := `DELETE FROM tasks WHERE chat_session_id = ?`
	result, err := r.db.Exec(query, chatSessionID)
	if err != nil {
		return fmt.Errorf("failed to clear tasks for session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("Cleared %d tasks for session %s\n", rowsAffected, chatSessionID)
	}

	return nil
}

// GetSessionSummary returns a summary of all chat sessions and their task counts
// This is useful for debugging session ID collisions
func (r *SqliteRepository) GetSessionSummary() (map[string]int, error) {
	query := `
	SELECT chat_session_id, COUNT(*) as task_count
	FROM tasks
	GROUP BY chat_session_id
	ORDER BY task_count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query session summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[string]int)
	for rows.Next() {
		var sessionID string
		var taskCount int
		if err := rows.Scan(&sessionID, &taskCount); err != nil {
			return nil, fmt.Errorf("failed to scan session summary row: %w", err)
		}
		summary[sessionID] = taskCount
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate session summary rows: %w", err)
	}

	return summary, nil
}

// GetAllTasksForSession returns all tasks for a specific session (for debugging)
func (r *SqliteRepository) GetAllTasksForSession(chatSessionID string) ([]*contracts.Task, error) {
	query := `
	SELECT id, chat_session_id, content, created_at
	FROM tasks
	WHERE chat_session_id = ?
	ORDER BY created_at ASC, id ASC
	`

	rows, err := r.db.Query(query, chatSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks for session: %w", err)
	}
	defer rows.Close()

	var tasks []*contracts.Task
	for rows.Next() {
		var task contracts.Task
		if err := rows.Scan(&task.ID, &task.ChatSessionID, &task.Content, &task.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate task rows: %w", err)
	}

	return tasks, nil
}
