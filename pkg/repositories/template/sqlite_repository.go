package template

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	_ "modernc.org/sqlite"
)

// SqliteRepository handles SQLite-based storage for task templates
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

// createTables creates the necessary tables for storing task templates
func (r *SqliteRepository) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS task_templates (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		category TEXT NOT NULL,
		parameters TEXT NOT NULL DEFAULT '{}',
		tasks TEXT NOT NULL DEFAULT '[]',
		estimated_time TEXT,
		prerequisites TEXT DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_template_category ON task_templates(category);
	CREATE INDEX IF NOT EXISTS idx_template_name ON task_templates(name);
	`

	_, err := r.db.Exec(query)
	return err
}

// CreateTemplate creates a new task template
func (r *SqliteRepository) CreateTemplate(template *contracts.TaskTemplate) error {
	if template.ID == "" {
		template.ID = generateTemplateID(template.Name)
	}

	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Serialize JSON fields
	parametersJSON, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	tasksJSON, err := json.Marshal(template.Tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	prerequisitesJSON, err := json.Marshal(template.Prerequisites)
	if err != nil {
		return fmt.Errorf("failed to marshal prerequisites: %w", err)
	}

	query := `
	INSERT INTO task_templates (id, name, description, category, parameters, tasks, estimated_time, prerequisites, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		template.ID,
		template.Name,
		template.Description,
		template.Category,
		string(parametersJSON),
		string(tasksJSON),
		template.EstimatedTime,
		string(prerequisitesJSON),
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// GetTemplate retrieves a template by ID
func (r *SqliteRepository) GetTemplate(id string) (*contracts.TaskTemplate, error) {
	query := `
	SELECT id, name, description, category, parameters, tasks, estimated_time, prerequisites, created_at, updated_at
	FROM task_templates
	WHERE id = ?
	`

	var template contracts.TaskTemplate
	var parametersJSON, tasksJSON, prerequisitesJSON string

	err := r.db.QueryRow(query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.Category,
		&parametersJSON,
		&tasksJSON,
		&template.EstimatedTime,
		&prerequisitesJSON,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal([]byte(parametersJSON), &template.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	if err := json.Unmarshal([]byte(tasksJSON), &template.Tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	if err := json.Unmarshal([]byte(prerequisitesJSON), &template.Prerequisites); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prerequisites: %w", err)
	}

	return &template, nil
}

// ListTemplates lists all templates, optionally filtered by category
func (r *SqliteRepository) ListTemplates(category string) ([]*contracts.TaskTemplate, error) {
	var query string
	var args []interface{}

	if category == "" {
		query = `
		SELECT id, name, description, category, parameters, tasks, estimated_time, prerequisites, created_at, updated_at
		FROM task_templates
		ORDER BY category, name
		`
	} else {
		query = `
		SELECT id, name, description, category, parameters, tasks, estimated_time, prerequisites, created_at, updated_at
		FROM task_templates
		WHERE category = ?
		ORDER BY name
		`
		args = append(args, category)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
		}
	}()

	var templates []*contracts.TaskTemplate

	for rows.Next() {
		var template contracts.TaskTemplate
		var parametersJSON, tasksJSON, prerequisitesJSON string

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.Category,
			&parametersJSON,
			&tasksJSON,
			&template.EstimatedTime,
			&prerequisitesJSON,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		// Deserialize JSON fields
		if err := json.Unmarshal([]byte(parametersJSON), &template.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		if err := json.Unmarshal([]byte(tasksJSON), &template.Tasks); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
		}

		if err := json.Unmarshal([]byte(prerequisitesJSON), &template.Prerequisites); err != nil {
			return nil, fmt.Errorf("failed to unmarshal prerequisites: %w", err)
		}

		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

// UpdateTemplate updates an existing template
func (r *SqliteRepository) UpdateTemplate(template *contracts.TaskTemplate) error {
	// Set updated timestamp
	template.UpdatedAt = time.Now()

	// Serialize JSON fields
	parametersJSON, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	tasksJSON, err := json.Marshal(template.Tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	prerequisitesJSON, err := json.Marshal(template.Prerequisites)
	if err != nil {
		return fmt.Errorf("failed to marshal prerequisites: %w", err)
	}

	query := `
	UPDATE task_templates
	SET name = ?, description = ?, category = ?, parameters = ?, tasks = ?, estimated_time = ?, prerequisites = ?, updated_at = ?
	WHERE id = ?
	`

	result, err := r.db.Exec(query,
		template.Name,
		template.Description,
		template.Category,
		string(parametersJSON),
		string(tasksJSON),
		template.EstimatedTime,
		string(prerequisitesJSON),
		template.UpdatedAt,
		template.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	return nil
}

// DeleteTemplate deletes a template by ID
func (r *SqliteRepository) DeleteTemplate(id string) error {
	query := `DELETE FROM task_templates WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", id)
	}

	return nil
}

// InstantiateTemplate creates a template instance with resolved parameters
func (r *SqliteRepository) InstantiateTemplate(templateID string, parameters map[string]string) (*contracts.TemplateInstance, error) {
	// Get the template
	template, err := r.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Fill in default values for missing optional parameters
	resolvedParameters := make(map[string]string)
	for paramName, param := range template.Parameters {
		if value, exists := parameters[paramName]; exists {
			resolvedParameters[paramName] = value
		} else if param.Default != "" {
			resolvedParameters[paramName] = param.Default
		}
	}

	// Resolve task templates
	resolvedTasks := make([]string, len(template.Tasks))
	for i, task := range template.Tasks {
		resolvedTasks[i] = r.resolveTemplate(task, resolvedParameters)
	}

	instance := &contracts.TemplateInstance{
		TemplateID: templateID,
		Parameters: resolvedParameters,
		Tasks:      resolvedTasks,
	}

	return instance, nil
}

// resolveTemplate resolves ${param} placeholders in a template string
func (r *SqliteRepository) resolveTemplate(template string, parameters map[string]string) string {
	// Use regex to find ${param} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract parameter name (remove ${ and })
		paramName := match[2 : len(match)-1]

		if value, exists := parameters[paramName]; exists {
			return value
		}

		// Return original if parameter not found
		return match
	})
}

// generateTemplateID generates a unique template ID based on the name
func generateTemplateID(name string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	id := strings.ToLower(name)
	id = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")

	// Add timestamp suffix to ensure uniqueness
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", id, timestamp)
}
