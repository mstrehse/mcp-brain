package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"gopkg.in/yaml.v3"
)

// FileRepository handles file-based storage for task templates using YAML files
type FileRepository struct {
	baseDir string
}

// NewFileRepository creates a new file-based template repository
func NewFileRepository(baseDir string) (*FileRepository, error) {
	templatesDir := filepath.Join(baseDir, "task-templates")

	// Ensure the templates directory exists
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	return &FileRepository{
		baseDir: templatesDir,
	}, nil
}

// Close is a no-op for file-based storage
func (r *FileRepository) Close() error {
	return nil
}

// getTemplateFilePath returns the file path for a template
func (r *FileRepository) getTemplateFilePath(id string) string {
	return filepath.Join(r.baseDir, id+".yaml")
}

// CreateTemplate creates a new task template
func (r *FileRepository) CreateTemplate(template *contracts.TaskTemplate) error {
	if template.ID == "" {
		template.ID = generateFileTemplateID(template.Name)
	}

	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Initialize empty slices if nil
	if template.Parameters == nil {
		template.Parameters = make(map[string]contracts.Parameter)
	}
	if template.Tasks == nil {
		template.Tasks = []string{}
	}
	if template.Prerequisites == nil {
		template.Prerequisites = []string{}
	}

	filePath := r.getTemplateFilePath(template.ID)

	// Check if template already exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return fmt.Errorf("template with ID %s already exists", template.ID)
	}

	return r.saveTemplate(template)
}

// GetTemplate retrieves a template by ID
func (r *FileRepository) GetTemplate(id string) (*contracts.TaskTemplate, error) {
	filePath := r.getTemplateFilePath(id)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template contracts.TaskTemplate
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}

// ListTemplates lists all templates, optionally filtered by category
func (r *FileRepository) ListTemplates(category string) ([]*contracts.TaskTemplate, error) {
	files, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []*contracts.TaskTemplate

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// Extract template ID from filename
		templateID := strings.TrimSuffix(file.Name(), ".yaml")

		template, err := r.GetTemplate(templateID)
		if err != nil {
			// Skip templates that can't be loaded
			continue
		}

		// Filter by category if specified
		if category != "" && template.Category != category {
			continue
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// UpdateTemplate updates an existing template
func (r *FileRepository) UpdateTemplate(template *contracts.TaskTemplate) error {
	if template.ID == "" {
		return fmt.Errorf("template ID is required for update")
	}

	filePath := r.getTemplateFilePath(template.ID)

	// Check if template exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	// Update timestamp
	template.UpdatedAt = time.Now()

	return r.saveTemplate(template)
}

// DeleteTemplate deletes a template by ID
func (r *FileRepository) DeleteTemplate(id string) error {
	filePath := r.getTemplateFilePath(id)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template not found: %s", id)
		}
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// InstantiateTemplate creates a template instance with resolved parameters
func (r *FileRepository) InstantiateTemplate(templateID string, parameters map[string]string) (*contracts.TemplateInstance, error) {
	template, err := r.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Validate required parameters
	for paramName, param := range template.Parameters {
		if param.Required {
			if _, exists := parameters[paramName]; !exists {
				return nil, fmt.Errorf("required parameter '%s' is missing", paramName)
			}
		}
	}

	// Resolve template strings
	resolvedTasks := make([]string, len(template.Tasks))
	for i, task := range template.Tasks {
		resolvedTasks[i] = r.resolveTemplate(task, parameters)
	}

	return &contracts.TemplateInstance{
		TemplateID: templateID,
		Parameters: parameters,
		Tasks:      resolvedTasks,
	}, nil
}

// saveTemplate saves a template to disk
func (r *FileRepository) saveTemplate(template *contracts.TaskTemplate) error {
	filePath := r.getTemplateFilePath(template.ID)

	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// resolveTemplate resolves template parameters in a string
func (r *FileRepository) resolveTemplate(template string, parameters map[string]string) string {
	// Replace ${param} with actual values
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		paramName := match[2 : len(match)-1] // Remove ${ and }
		if value, exists := parameters[paramName]; exists {
			return value
		}
		return match // Return original if parameter not found
	})
}

// generateFileTemplateID generates a template ID from a name
func generateFileTemplateID(name string) string {
	// Convert to lowercase and replace non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	id := re.ReplaceAllString(strings.ToLower(name), "-")

	// Remove leading/trailing hyphens
	id = strings.Trim(id, "-")

	// Add timestamp suffix to ensure uniqueness
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", id, timestamp)
}
