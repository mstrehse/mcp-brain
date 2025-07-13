package template

import (
	"os"
	"testing"
	"time"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

func TestFileRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_template_repo")
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

	// Create a test template
	template := &contracts.TaskTemplate{
		Name:        "Test Template",
		Description: "A template for testing",
		Category:    "testing",
		Parameters: map[string]contracts.Parameter{
			"project_name": {
				Type:        "string",
				Description: "Name of the project",
				Required:    true,
			},
			"language": {
				Type:        "enum",
				Description: "Programming language",
				Required:    false,
				Default:     "go",
				Values:      []string{"go", "python", "javascript"},
			},
		},
		Tasks: []string{
			"Create project structure for ${project_name}",
			"Initialize ${language} project",
			"Set up basic configuration",
		},
		Prerequisites: []string{
			"Ensure development environment is set up",
		},
	}

	// Test CreateTemplate
	err = repo.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Verify template has ID and timestamps
	if template.ID == "" {
		t.Error("Template ID should be generated")
	}
	if template.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if template.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Test GetTemplate
	retrieved, err := repo.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrieved.Name != template.Name {
		t.Errorf("Expected name %s, got %s", template.Name, retrieved.Name)
	}
	if retrieved.Description != template.Description {
		t.Errorf("Expected description %s, got %s", template.Description, retrieved.Description)
	}
	if retrieved.Category != template.Category {
		t.Errorf("Expected category %s, got %s", template.Category, retrieved.Category)
	}
	if len(retrieved.Parameters) != len(template.Parameters) {
		t.Errorf("Expected %d parameters, got %d", len(template.Parameters), len(retrieved.Parameters))
	}
	if len(retrieved.Tasks) != len(template.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(template.Tasks), len(retrieved.Tasks))
	}
}

func TestFileRepositoryListTemplates(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_template_repo")
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

	// Create templates in different categories
	templates := []*contracts.TaskTemplate{
		{
			Name:        "Development Template",
			Description: "For development tasks",
			Category:    "development",
			Parameters:  map[string]contracts.Parameter{},
			Tasks:       []string{"Dev task 1", "Dev task 2"},
		},
		{
			Name:        "Testing Template",
			Description: "For testing tasks",
			Category:    "testing",
			Parameters:  map[string]contracts.Parameter{},
			Tasks:       []string{"Test task 1", "Test task 2"},
		},
		{
			Name:        "Another Development Template",
			Description: "Another dev template",
			Category:    "development",
			Parameters:  map[string]contracts.Parameter{},
			Tasks:       []string{"Dev task 3"},
		},
	}

	// Create all templates
	for _, template := range templates {
		err = repo.CreateTemplate(template)
		if err != nil {
			t.Fatalf("Failed to create template %s: %v", template.Name, err)
		}
	}

	// Test listing all templates
	allTemplates, err := repo.ListTemplates("")
	if err != nil {
		t.Fatalf("Failed to list all templates: %v", err)
	}
	if len(allTemplates) != 3 {
		t.Errorf("Expected 3 templates, got %d", len(allTemplates))
	}

	// Test listing templates by category
	devTemplates, err := repo.ListTemplates("development")
	if err != nil {
		t.Fatalf("Failed to list development templates: %v", err)
	}
	if len(devTemplates) != 2 {
		t.Errorf("Expected 2 development templates, got %d", len(devTemplates))
	}

	testingTemplates, err := repo.ListTemplates("testing")
	if err != nil {
		t.Fatalf("Failed to list testing templates: %v", err)
	}
	if len(testingTemplates) != 1 {
		t.Errorf("Expected 1 testing template, got %d", len(testingTemplates))
	}

	// Test listing non-existent category
	nonExistentTemplates, err := repo.ListTemplates("nonexistent")
	if err != nil {
		t.Fatalf("Failed to list non-existent category templates: %v", err)
	}
	if len(nonExistentTemplates) != 0 {
		t.Errorf("Expected 0 templates for non-existent category, got %d", len(nonExistentTemplates))
	}
}

func TestFileRepositoryUpdateTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_template_repo")
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

	// Create a template
	template := &contracts.TaskTemplate{
		Name:        "Original Template",
		Description: "Original description",
		Category:    "original",
		Parameters:  map[string]contracts.Parameter{},
		Tasks:       []string{"Original task"},
	}

	err = repo.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	originalUpdatedAt := template.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp changes

	// Update the template
	template.Name = "Updated Template"
	template.Description = "Updated description"
	template.Category = "updated"
	template.Tasks = []string{"Updated task 1", "Updated task 2"}

	err = repo.UpdateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Verify UpdatedAt timestamp was updated
	if !template.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt timestamp should be updated")
	}

	// Retrieve and verify updates
	updated, err := repo.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to get updated template: %v", err)
	}

	if updated.Name != "Updated Template" {
		t.Errorf("Expected name 'Updated Template', got %s", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got %s", updated.Description)
	}
	if updated.Category != "updated" {
		t.Errorf("Expected category 'updated', got %s", updated.Category)
	}
	if len(updated.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(updated.Tasks))
	}
}

func TestFileRepositoryDeleteTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_template_repo")
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

	// Create a template
	template := &contracts.TaskTemplate{
		Name:        "Template to Delete",
		Description: "This template will be deleted",
		Category:    "temporary",
		Parameters:  map[string]contracts.Parameter{},
		Tasks:       []string{"Temporary task"},
	}

	err = repo.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Verify template exists
	_, err = repo.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("Template should exist before deletion: %v", err)
	}

	// Delete the template
	err = repo.DeleteTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	// Verify template no longer exists
	_, err = repo.GetTemplate(template.ID)
	if err == nil {
		t.Fatal("Template should not exist after deletion")
	}
}

func TestFileRepositoryInstantiateTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_template_repo")
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

	// Create a template with parameters
	template := &contracts.TaskTemplate{
		Name:        "Project Template",
		Description: "Creates a new project",
		Category:    "development",
		Parameters: map[string]contracts.Parameter{
			"project_name": {
				Type:        "string",
				Description: "Name of the project",
				Required:    true,
			},
			"language": {
				Type:        "enum",
				Description: "Programming language",
				Required:    false,
				Default:     "go",
				Values:      []string{"go", "python", "javascript"},
			},
		},
		Tasks: []string{
			"Create directory structure for ${project_name}",
			"Initialize ${language} project in ${project_name}",
			"Set up configuration for ${project_name}",
		},
	}

	err = repo.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Test instantiation with required parameters
	parameters := map[string]string{
		"project_name": "my-awesome-project",
		"language":     "python",
	}

	instance, err := repo.InstantiateTemplate(template.ID, parameters)
	if err != nil {
		t.Fatalf("Failed to instantiate template: %v", err)
	}

	if instance.TemplateID != template.ID {
		t.Errorf("Expected template ID %s, got %s", template.ID, instance.TemplateID)
	}

	expectedTasks := []string{
		"Create directory structure for my-awesome-project",
		"Initialize python project in my-awesome-project",
		"Set up configuration for my-awesome-project",
	}

	if len(instance.Tasks) != len(expectedTasks) {
		t.Errorf("Expected %d tasks, got %d", len(expectedTasks), len(instance.Tasks))
	}

	for i, expectedTask := range expectedTasks {
		if instance.Tasks[i] != expectedTask {
			t.Errorf("Expected task %d to be '%s', got '%s'", i, expectedTask, instance.Tasks[i])
		}
	}

	// Test instantiation with missing required parameter
	invalidParameters := map[string]string{
		"language": "go",
		// Missing required "project_name"
	}

	_, err = repo.InstantiateTemplate(template.ID, invalidParameters)
	if err == nil {
		t.Fatal("Expected error for missing required parameter")
	}
}
