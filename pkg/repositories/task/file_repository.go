package task

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"gopkg.in/yaml.v3"
)

// TasksFile represents the structure of the tasks.yaml file
type TasksFile struct {
	Tasks      []*contracts.Task `yaml:"tasks"`
	NextID     int               `yaml:"next_id"`
	LastUpdate time.Time         `yaml:"last_update"`
}

// FileRepository handles file-based storage for tasks using a single YAML file
type FileRepository struct {
	filePath string
	mutex    sync.RWMutex
}

// NewFileRepository creates a new file-based task repository
func NewFileRepository(baseDir string) (*FileRepository, error) {
	// Ensure the brain directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create brain directory: %w", err)
	}

	filePath := filepath.Join(baseDir, "tasks.yaml")

	repo := &FileRepository{
		filePath: filePath,
	}

	// Initialize file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := repo.saveTasksFile(&TasksFile{
			Tasks:      []*contracts.Task{},
			NextID:     1,
			LastUpdate: time.Now(),
		}); err != nil {
			return nil, fmt.Errorf("failed to initialize tasks file: %w", err)
		}
	}

	return repo, nil
}

// Close is a no-op for file-based storage
func (r *FileRepository) Close() error {
	return nil
}

// loadTasksFile loads the tasks file from disk
func (r *FileRepository) loadTasksFile() (*TasksFile, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TasksFile{
				Tasks:      []*contracts.Task{},
				NextID:     1,
				LastUpdate: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	var tasksFile TasksFile
	if err := yaml.Unmarshal(data, &tasksFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks file: %w", err)
	}

	return &tasksFile, nil
}

// saveTasksFile saves the tasks file to disk
func (r *FileRepository) saveTasksFile(tasksFile *TasksFile) error {
	tasksFile.LastUpdate = time.Now()

	data, err := yaml.Marshal(tasksFile)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks file: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// AddTasks adds multiple tasks to the queue
func (r *FileRepository) AddTasks(contents []string) ([]*contracts.Task, error) {
	if len(contents) == 0 {
		return []*contracts.Task{}, nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	tasksFile, err := r.loadTasksFile()
	if err != nil {
		return nil, err
	}

	var newTasks []*contracts.Task
	now := time.Now()

	for _, content := range contents {
		task := &contracts.Task{
			ID:        tasksFile.NextID,
			Content:   content,
			CreatedAt: now,
		}
		newTasks = append(newTasks, task)
		tasksFile.Tasks = append(tasksFile.Tasks, task)
		tasksFile.NextID++
	}

	if err := r.saveTasksFile(tasksFile); err != nil {
		return nil, err
	}

	return newTasks, nil
}

// GetTask retrieves and removes the next pending task from the queue
func (r *FileRepository) GetTask() (*contracts.Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tasksFile, err := r.loadTasksFile()
	if err != nil {
		return nil, err
	}

	if len(tasksFile.Tasks) == 0 {
		return nil, nil
	}

	// Get the first task (FIFO)
	task := tasksFile.Tasks[0]
	tasksFile.Tasks = tasksFile.Tasks[1:]

	if err := r.saveTasksFile(tasksFile); err != nil {
		return nil, err
	}

	return task, nil
}

// Additional methods for testing/debugging purposes (not part of the interface)

// ClearAllTasks removes all tasks from the queue
func (r *FileRepository) ClearAllTasks() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tasksFile := &TasksFile{
		Tasks:      []*contracts.Task{},
		NextID:     1,
		LastUpdate: time.Now(),
	}

	return r.saveTasksFile(tasksFile)
}

// GetTaskCount returns the number of tasks in the queue
func (r *FileRepository) GetTaskCount() (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasksFile, err := r.loadTasksFile()
	if err != nil {
		return 0, err
	}

	return len(tasksFile.Tasks), nil
}

// GetAllTasks returns all tasks in the queue (for testing purposes)
func (r *FileRepository) GetAllTasks() ([]*contracts.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasksFile, err := r.loadTasksFile()
	if err != nil {
		return nil, err
	}

	// Return a copy to avoid external modification
	tasks := make([]*contracts.Task, len(tasksFile.Tasks))
	copy(tasks, tasksFile.Tasks)

	return tasks, nil
}
