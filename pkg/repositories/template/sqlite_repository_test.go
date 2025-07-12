package template

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
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

func createTestTemplate() *contracts.TaskTemplate {
	return &contracts.TaskTemplate{
		Name:        "Test Template",
		Description: "A test template for unit testing",
		Category:    "testing",
		Parameters: map[string]contracts.Parameter{
			"feature_name": {
				Type:        "string",
				Description: "Name of the feature",
				Required:    true,
			},
			"priority": {
				Type:        "enum",
				Description: "Priority level",
				Required:    false,
				Default:     "medium",
				Values:      []string{"low", "medium", "high"},
			},
		},
		Tasks: []string{
			"Create feature branch for ${feature_name}",
			"Implement ${feature_name} with ${priority} priority",
			"Test ${feature_name} functionality",
		},
		EstimatedTime: "2-3 hours",
		Prerequisites: []string{"Git repository", "Development environment"},
	}
}

func TestSqliteRepository_CreateGetTemplate(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	template := createTestTemplate()

	// Create template
	if err := repo.CreateTemplate(template); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// Verify ID was generated
	if template.ID == "" {
		t.Error("Expected template ID to be generated")
	}

	// Verify timestamps were set
	if template.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if template.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Get template
	retrieved, err := repo.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	// Verify template data
	if retrieved.ID != template.ID {
		t.Errorf("Expected ID %s, got %s", template.ID, retrieved.ID)
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
	if retrieved.EstimatedTime != template.EstimatedTime {
		t.Errorf("Expected estimated_time %s, got %s", template.EstimatedTime, retrieved.EstimatedTime)
	}

	// Verify parameters
	if len(retrieved.Parameters) != len(template.Parameters) {
		t.Errorf("Expected %d parameters, got %d", len(template.Parameters), len(retrieved.Parameters))
	}

	featureParam := retrieved.Parameters["feature_name"]
	if featureParam.Type != "string" {
		t.Errorf("Expected parameter type 'string', got %s", featureParam.Type)
	}
	if !featureParam.Required {
		t.Error("Expected feature_name parameter to be required")
	}

	priorityParam := retrieved.Parameters["priority"]
	if priorityParam.Type != "enum" {
		t.Errorf("Expected parameter type 'enum', got %s", priorityParam.Type)
	}
	if priorityParam.Default != "medium" {
		t.Errorf("Expected default value 'medium', got %s", priorityParam.Default)
	}

	// Verify tasks
	if len(retrieved.Tasks) != len(template.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(template.Tasks), len(retrieved.Tasks))
	}

	// Verify prerequisites
	if len(retrieved.Prerequisites) != len(template.Prerequisites) {
		t.Errorf("Expected %d prerequisites, got %d", len(template.Prerequisites), len(retrieved.Prerequisites))
	}
}

func TestSqliteRepository_GetTemplate_NotFound(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	_, err := repo.GetTemplate("nonexistent-id")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestSqliteRepository_ListTemplates(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	// Create multiple templates
	template1 := createTestTemplate()
	template1.Name = "Template 1"
	template1.Category = "development"

	template2 := createTestTemplate()
	template2.Name = "Template 2"
	template2.Category = "testing"

	template3 := createTestTemplate()
	template3.Name = "Template 3"
	template3.Category = "development"

	if err := repo.CreateTemplate(template1); err != nil {
		t.Fatalf("Failed to create template1: %v", err)
	}
	if err := repo.CreateTemplate(template2); err != nil {
		t.Fatalf("Failed to create template2: %v", err)
	}
	if err := repo.CreateTemplate(template3); err != nil {
		t.Fatalf("Failed to create template3: %v", err)
	}

	// List all templates
	allTemplates, err := repo.ListTemplates("")
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(allTemplates) != 3 {
		t.Errorf("Expected 3 templates, got %d", len(allTemplates))
	}

	// List templates by category
	devTemplates, err := repo.ListTemplates("development")
	if err != nil {
		t.Fatalf("ListTemplates with category failed: %v", err)
	}
	if len(devTemplates) != 2 {
		t.Errorf("Expected 2 development templates, got %d", len(devTemplates))
	}

	testTemplates, err := repo.ListTemplates("testing")
	if err != nil {
		t.Fatalf("ListTemplates with category failed: %v", err)
	}
	if len(testTemplates) != 1 {
		t.Errorf("Expected 1 testing template, got %d", len(testTemplates))
	}
}

func TestSqliteRepository_UpdateTemplate(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	template := createTestTemplate()
	if err := repo.CreateTemplate(template); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	originalUpdatedAt := template.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update template
	template.Name = "Updated Template"
	template.Description = "Updated description"
	template.Category = "updated"

	if err := repo.UpdateTemplate(template); err != nil {
		t.Fatalf("UpdateTemplate failed: %v", err)
	}

	// Verify UpdatedAt was changed
	if !template.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Retrieve and verify changes
	updated, err := repo.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
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
}

func TestSqliteRepository_UpdateTemplate_NotFound(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	template := createTestTemplate()
	template.ID = "nonexistent-id"

	err := repo.UpdateTemplate(template)
	if err == nil {
		t.Error("Expected error when updating non-existent template")
	}
}

func TestSqliteRepository_DeleteTemplate(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	template := createTestTemplate()
	if err := repo.CreateTemplate(template); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// Delete template
	if err := repo.DeleteTemplate(template.ID); err != nil {
		t.Fatalf("DeleteTemplate failed: %v", err)
	}

	// Verify template is gone
	_, err := repo.GetTemplate(template.ID)
	if err == nil {
		t.Error("Expected error when getting deleted template")
	}
}

func TestSqliteRepository_DeleteTemplate_NotFound(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	err := repo.DeleteTemplate("nonexistent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent template")
	}
}

func TestSqliteRepository_InstantiateTemplate(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	template := createTestTemplate()
	if err := repo.CreateTemplate(template); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// Instantiate template with parameters
	parameters := map[string]string{
		"feature_name": "user-authentication",
		"priority":     "high",
	}

	instance, err := repo.InstantiateTemplate(template.ID, parameters)
	if err != nil {
		t.Fatalf("InstantiateTemplate failed: %v", err)
	}

	// Verify instance
	if instance.TemplateID != template.ID {
		t.Errorf("Expected template ID %s, got %s", template.ID, instance.TemplateID)
	}

	if len(instance.Tasks) != len(template.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(template.Tasks), len(instance.Tasks))
	}

	// Verify parameter substitution
	expectedTasks := []string{
		"Create feature branch for user-authentication",
		"Implement user-authentication with high priority",
		"Test user-authentication functionality",
	}

	for i, task := range instance.Tasks {
		if task != expectedTasks[i] {
			t.Errorf("Expected task %d to be %s, got %s", i, expectedTasks[i], task)
		}
	}

	// Verify parameters include defaults
	if instance.Parameters["feature_name"] != "user-authentication" {
		t.Errorf("Expected feature_name to be 'user-authentication', got %s", instance.Parameters["feature_name"])
	}
	if instance.Parameters["priority"] != "high" {
		t.Errorf("Expected priority to be 'high', got %s", instance.Parameters["priority"])
	}
}

func TestSqliteRepository_InstantiateTemplate_WithDefaults(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	template := createTestTemplate()
	if err := repo.CreateTemplate(template); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// Instantiate template with only required parameters
	parameters := map[string]string{
		"feature_name": "user-registration",
	}

	instance, err := repo.InstantiateTemplate(template.ID, parameters)
	if err != nil {
		t.Fatalf("InstantiateTemplate failed: %v", err)
	}

	// Verify default value was used
	if instance.Parameters["priority"] != "medium" {
		t.Errorf("Expected priority default 'medium', got %s", instance.Parameters["priority"])
	}

	// Verify task contains default value
	expectedTask := "Implement user-registration with medium priority"
	if instance.Tasks[1] != expectedTask {
		t.Errorf("Expected task to be %s, got %s", expectedTask, instance.Tasks[1])
	}
}

func TestSqliteRepository_InstantiateTemplate_NotFound(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	_, err := repo.InstantiateTemplate("nonexistent-id", map[string]string{})
	if err == nil {
		t.Error("Expected error when instantiating non-existent template")
	}
}

func TestSqliteRepository_DatabasePersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persistence.db")

	// Create repo and add template
	repo1, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	template := createTestTemplate()
	if err := repo1.CreateTemplate(template); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	templateID := template.ID
	if err := repo1.Close(); err != nil {
		t.Fatalf("Failed to close repository: %v", err)
	}

	// Create new repo instance with same database
	repo2, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	// Template should still exist
	persistedTemplate, err := repo2.GetTemplate(templateID)
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if persistedTemplate.Name != template.Name {
		t.Errorf("Expected name %s, got %s", template.Name, persistedTemplate.Name)
	}
}

func TestGenerateTemplateID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Only check the prefix since timestamp is added
	}{
		{
			name:     "Simple name",
			input:    "Test Template",
			expected: "test-template",
		},
		{
			name:     "With special characters",
			input:    "Code Review & Testing!",
			expected: "code-review-testing",
		},
		{
			name:     "Already lowercase",
			input:    "simple-name",
			expected: "simple-name",
		},
		{
			name:     "With numbers",
			input:    "Version 2.0 Template",
			expected: "version-2-0-template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateTemplateID(tt.input)
			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("Expected ID to start with %s, got %s", tt.expected, result)
			}
		})
	}
}
