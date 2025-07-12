package contracts

import "time"

// TaskTemplate represents a reusable task workflow template
type TaskTemplate struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Category      string               `json:"category"`
	Parameters    map[string]Parameter `json:"parameters"`
	Tasks         []string             `json:"tasks"`
	EstimatedTime string               `json:"estimated_time,omitempty"`
	Prerequisites []string             `json:"prerequisites,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// Parameter defines a template parameter
type Parameter struct {
	Type        string   `json:"type"` // string, enum, number, boolean
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Default     string   `json:"default,omitempty"`
	Values      []string `json:"values,omitempty"` // for enum type
}

// TemplateInstance represents an instantiated template with resolved parameters
type TemplateInstance struct {
	TemplateID string            `json:"template_id"`
	Parameters map[string]string `json:"parameters"`
	Tasks      []string          `json:"tasks"`
}

// TaskTemplateRepository defines the interface for template operations
type TaskTemplateRepository interface {
	// CreateTemplate creates a new task template
	CreateTemplate(template *TaskTemplate) error

	// GetTemplate retrieves a template by ID
	GetTemplate(id string) (*TaskTemplate, error)

	// ListTemplates lists all templates, optionally filtered by category
	ListTemplates(category string) ([]*TaskTemplate, error)

	// UpdateTemplate updates an existing template
	UpdateTemplate(template *TaskTemplate) error

	// DeleteTemplate deletes a template by ID
	DeleteTemplate(id string) error

	// InstantiateTemplate creates a template instance with resolved parameters
	InstantiateTemplate(templateID string, parameters map[string]string) (*TemplateInstance, error)

	// Close closes the repository and cleans up resources
	Close() error
}
