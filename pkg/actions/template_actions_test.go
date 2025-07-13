package actions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/task"
	"github.com/mstrehse/mcp-brain/pkg/repositories/template"
)

// Helper function to create a valid test template
func createTestTemplate() *contracts.TaskTemplate {
	return &contracts.TaskTemplate{
		ID:          "test-template-1",
		Name:        "Test Template",
		Description: "A test template for validation",
		Parameters: map[string]contracts.Parameter{
			"project_name": {
				Type:        "string",
				Description: "Name of the project",
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
			"Setup project ${project_name}",
			"Set priority to ${priority}",
			"Initialize repository",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestTaskTemplateCreateHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := template.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTaskTemplateCreateHandler(repo)

	t.Run("successful create", func(t *testing.T) {
		template := createTestTemplate()
		templateJSON, _ := json.Marshal(template)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-create",
				Arguments: map[string]interface{}{
					"template": string(templateJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON
		var createResult map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &createResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Check expected fields
		if message, ok := createResult["message"].(string); !ok || message != "Template created successfully" {
			t.Errorf("Expected success message, got: %v", createResult["message"])
		}

		if templateID, ok := createResult["template_id"].(string); !ok || templateID != template.ID {
			t.Errorf("Expected template_id %s, got: %v", template.ID, createResult["template_id"])
		}
	})

	t.Run("missing template parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-template-create",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing template")
		}
	})

	t.Run("invalid template JSON", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-create",
				Arguments: map[string]interface{}{
					"template": "invalid json",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for invalid JSON")
		}
	})

	t.Run("template validation failure", func(t *testing.T) {
		invalidTemplate := &contracts.TaskTemplate{
			ID:          "invalid-template",
			Name:        "", // Missing required name
			Description: "Test",
			Tasks:       []string{"task 1"},
		}
		templateJSON, _ := json.Marshal(invalidTemplate)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-create",
				Arguments: map[string]interface{}{
					"template": string(templateJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for validation failure")
		}
	})
}

func TestTaskTemplateGetHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := template.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTaskTemplateGetHandler(repo)

	// Setup test data
	testTemplate := createTestTemplate()
	if err := repo.CreateTemplate(testTemplate); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful get", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-get",
				Arguments: map[string]interface{}{
					"template_id": testTemplate.ID,
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON and contains the template
		var retrievedTemplate contracts.TaskTemplate
		if err := json.Unmarshal([]byte(textContent.Text), &retrievedTemplate); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		if retrievedTemplate.ID != testTemplate.ID {
			t.Errorf("Expected template ID %s, got %s", testTemplate.ID, retrievedTemplate.ID)
		}

		if retrievedTemplate.Name != testTemplate.Name {
			t.Errorf("Expected template name %s, got %s", testTemplate.Name, retrievedTemplate.Name)
		}
	})

	t.Run("missing template_id parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-template-get",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing template_id")
		}
	})

	t.Run("non-existent template", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-get",
				Arguments: map[string]interface{}{
					"template_id": "non-existent-template",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for non-existent template")
		}
	})
}

func TestTaskTemplatesListHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := template.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTaskTemplatesListHandler(repo)

	// Setup test data
	template1 := createTestTemplate()
	template1.ID = "template-1"

	template2 := createTestTemplate()
	template2.ID = "template-2"

	if err := repo.CreateTemplate(template1); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}
	if err := repo.CreateTemplate(template2); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("list all templates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-templates-list",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON
		var listResult map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &listResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Check count
		if count, ok := listResult["count"].(float64); !ok || int(count) != 2 {
			t.Errorf("Expected count 2, got: %v", listResult["count"])
		}
	})
}

func TestTaskTemplateDeleteHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := template.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTaskTemplateDeleteHandler(repo)

	// Setup test data
	testTemplate := createTestTemplate()
	if err := repo.CreateTemplate(testTemplate); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-delete",
				Arguments: map[string]interface{}{
					"template_id": testTemplate.ID,
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON
		var deleteResult map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &deleteResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Check expected fields
		if message, ok := deleteResult["message"].(string); !ok || message != "Template deleted successfully" {
			t.Errorf("Expected success message, got: %v", deleteResult["message"])
		}

		// Verify template was actually deleted
		_, err = repo.GetTemplate(testTemplate.ID)
		if err == nil {
			t.Error("Expected template to be deleted")
		}
	})

	t.Run("missing template_id parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-template-delete",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing template_id")
		}
	})

	t.Run("non-existent template", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-delete",
				Arguments: map[string]interface{}{
					"template_id": "non-existent-template",
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for non-existent template")
		}
	})
}

func TestTaskTemplateUpdateHandler(t *testing.T) {
	baseDir := t.TempDir()
	repo, err := template.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer func() { _ = repo.Close() }()

	handler := NewTaskTemplateUpdateHandler(repo)

	// Setup test data
	testTemplate := createTestTemplate()
	if err := repo.CreateTemplate(testTemplate); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		updatedTemplate := createTestTemplate()
		updatedTemplate.Name = "Updated Test Template"
		updatedTemplate.Description = "An updated test template"
		templateJSON, _ := json.Marshal(updatedTemplate)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-update",
				Arguments: map[string]interface{}{
					"template": string(templateJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON
		var updateResult map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &updateResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Check expected fields
		if message, ok := updateResult["message"].(string); !ok || message != "Template updated successfully" {
			t.Errorf("Expected success message, got: %v", updateResult["message"])
		}

		// Verify template was actually updated
		retrievedTemplate, err := repo.GetTemplate(testTemplate.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated template: %v", err)
		}

		if retrievedTemplate.Name != "Updated Test Template" {
			t.Errorf("Expected updated name 'Updated Test Template', got %s", retrievedTemplate.Name)
		}
	})

	t.Run("missing template parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-template-update",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing template")
		}
	})

	t.Run("template without ID", func(t *testing.T) {
		templateWithoutID := createTestTemplate()
		templateWithoutID.ID = "" // Remove ID
		templateJSON, _ := json.Marshal(templateWithoutID)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-update",
				Arguments: map[string]interface{}{
					"template": string(templateJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for template without ID")
		}
	})

	t.Run("non-existent template", func(t *testing.T) {
		nonExistentTemplate := createTestTemplate()
		nonExistentTemplate.ID = "non-existent-template"
		templateJSON, _ := json.Marshal(nonExistentTemplate)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-update",
				Arguments: map[string]interface{}{
					"template": string(templateJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for non-existent template")
		}
	})
}

func TestTaskTemplateInstantiateHandler(t *testing.T) {
	baseDir := t.TempDir()
	templateRepo, err := template.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create template repository: %v", err)
	}
	defer func() { _ = templateRepo.Close() }()

	taskRepo, err := task.NewFileRepository(baseDir)
	if err != nil {
		t.Fatalf("Failed to create task repository: %v", err)
	}
	defer func() { _ = taskRepo.Close() }()

	handler := NewTaskTemplateInstantiateHandler(templateRepo, taskRepo)

	// Setup test data
	testTemplate := createTestTemplate()
	if err := templateRepo.CreateTemplate(testTemplate); err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	t.Run("successful instantiate with parameters", func(t *testing.T) {
		parameters := map[string]string{
			"project_name": "MyAwesomeProject",
			"priority":     "high",
		}
		paramJSON, _ := json.Marshal(parameters)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-instantiate",
				Arguments: map[string]interface{}{
					"template_id": testTemplate.ID,
					"parameters":  string(paramJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Errorf("Handler returned error: %s", textContent.Text)
			} else {
				t.Error("Handler returned error but couldn't read error message")
			}
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected result content")
		}

		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content")
		}

		// Verify the result is valid JSON
		var instantiateResult map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &instantiateResult); err != nil {
			t.Errorf("Result is not valid JSON: %v", err)
		}

		// Check expected fields
		if message, ok := instantiateResult["message"].(string); !ok || message != "Template instantiated successfully" {
			t.Errorf("Expected success message, got: %v", instantiateResult["message"])
		}

		if tasksAdded, ok := instantiateResult["tasks_added"].(float64); !ok || int(tasksAdded) != len(testTemplate.Tasks) {
			t.Errorf("Expected tasks_added %d, got: %v", len(testTemplate.Tasks), instantiateResult["tasks_added"])
		}
	})

	t.Run("missing template_id parameter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "task-template-instantiate",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing template_id")
		}
	})

	t.Run("missing required parameter", func(t *testing.T) {
		parameters := map[string]string{
			// Missing required "project_name" parameter
			"priority": "medium",
		}
		paramJSON, _ := json.Marshal(parameters)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "task-template-instantiate",
				Arguments: map[string]interface{}{
					"template_id": testTemplate.ID,
					"parameters":  string(paramJSON),
				},
			},
		}

		result, err := handler(context.Background(), request)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing required parameter")
		}
	})
}
