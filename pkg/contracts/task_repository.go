package contracts

import "time"

// Task represents a task in the queue
type Task struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskRepository defines the interface for task queue operations
type TaskRepository interface {
	// AddTasks adds multiple tasks to the queue
	AddTasks(contents []string) ([]*Task, error)

	// GetTask retrieves and removes the next pending task from the queue
	GetTask() (*Task, error)

	// Close closes the repository and cleans up resources
	Close() error
}
