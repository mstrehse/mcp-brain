package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewListTemplatesHandler creates a handler for listing task templates
func NewListTemplatesHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		category := request.GetString("category", "") // optional parameter

		templates, err := repo.ListTemplates(category)
		if err != nil {
			return mcp.NewToolResultError("Failed to list templates: " + err.Error()), nil
		}

		result := map[string]interface{}{
			"templates": templates,
			"count":     len(templates),
		}

		if category != "" {
			result["category"] = category
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal templates: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// NewGetTemplateHandler creates a handler for getting a specific template
func NewGetTemplateHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := request.RequireString("template_id")
		if err != nil {
			return mcp.NewToolResultError("Missing 'template_id' parameter: " + err.Error()), nil
		}

		template, err := repo.GetTemplate(templateID)
		if err != nil {
			return mcp.NewToolResultError("Failed to get template: " + err.Error()), nil
		}

		data, err := json.Marshal(template)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal template: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// NewCreateTemplateHandler creates a handler for creating new task templates
func NewCreateTemplateHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateJSON, err := request.RequireString("template")
		if err != nil {
			return mcp.NewToolResultError("Missing 'template' parameter: " + err.Error()), nil
		}

		var template contracts.TaskTemplate
		if err := json.Unmarshal([]byte(templateJSON), &template); err != nil {
			return mcp.NewToolResultError("Invalid template JSON: " + err.Error()), nil
		}

		if err := validateTemplate(&template); err != nil {
			return mcp.NewToolResultError("Template validation failed: " + err.Error()), nil
		}

		if err := repo.CreateTemplate(&template); err != nil {
			return mcp.NewToolResultError("Failed to create template: " + err.Error()), nil
		}

		result := map[string]interface{}{
			"message":     "Template created successfully",
			"template_id": template.ID,
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// NewInstantiateTemplateHandler creates a handler for instantiating templates
func NewInstantiateTemplateHandler(repo contracts.TaskTemplateRepository, taskRepo contracts.TaskRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := request.RequireString("template_id")
		if err != nil {
			return mcp.NewToolResultError("Missing 'template_id' parameter: " + err.Error()), nil
		}

		chatSessionID, err := request.RequireString("chat_session_id")
		if err != nil {
			return mcp.NewToolResultError("Missing 'chat_session_id' parameter: " + err.Error()), nil
		}

		parametersJSON := request.GetString("parameters", "")
		var parameters map[string]string
		if parametersJSON != "" {
			if err := json.Unmarshal([]byte(parametersJSON), &parameters); err != nil {
				return mcp.NewToolResultError("Invalid parameters JSON: " + err.Error()), nil
			}
		}

		// Get the template first to validate parameters
		template, err := repo.GetTemplate(templateID)
		if err != nil {
			return mcp.NewToolResultError("Failed to get template: " + err.Error()), nil
		}

		// Validate parameters
		if err := validateTemplateParameters(template, parameters); err != nil {
			return mcp.NewToolResultError("Parameter validation failed: " + err.Error()), nil
		}

		// Instantiate the template
		instance, err := repo.InstantiateTemplate(templateID, parameters)
		if err != nil {
			return mcp.NewToolResultError("Failed to instantiate template: " + err.Error()), nil
		}

		// Add the resolved tasks to the task queue
		addedTasks, err := taskRepo.AddTasks(chatSessionID, instance.Tasks)
		if err != nil {
			return mcp.NewToolResultError("Failed to add tasks from template: " + err.Error()), nil
		}

		result := map[string]interface{}{
			"message":     "Template instantiated successfully",
			"template_id": templateID,
			"tasks_added": len(addedTasks),
			"tasks":       addedTasks,
			"parameters":  parameters,
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// validateTemplate validates a task template
func validateTemplate(template *contracts.TaskTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Description == "" {
		return fmt.Errorf("template description is required")
	}
	if len(template.Tasks) == 0 {
		return fmt.Errorf("template must have at least one task")
	}

	// Validate parameters
	for paramName, param := range template.Parameters {
		if paramName == "" {
			return fmt.Errorf("parameter name cannot be empty")
		}
		if param.Type == "" {
			return fmt.Errorf("parameter '%s' must have a type", paramName)
		}
		if param.Type == "enum" && len(param.Values) == 0 {
			return fmt.Errorf("enum parameter '%s' must have values", paramName)
		}
	}

	return nil
}

// validateTemplateParameters validates that provided parameters match template requirements
func validateTemplateParameters(template *contracts.TaskTemplate, parameters map[string]string) error {
	// Check required parameters
	for paramName, param := range template.Parameters {
		if param.Required {
			if value, exists := parameters[paramName]; !exists || value == "" {
				return fmt.Errorf("required parameter '%s' is missing", paramName)
			}
		}
	}

	// Validate enum parameters
	for paramName, value := range parameters {
		if param, exists := template.Parameters[paramName]; exists {
			if param.Type == "enum" && len(param.Values) > 0 {
				valid := false
				for _, validValue := range param.Values {
					if value == validValue {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("parameter '%s' must be one of: %s", paramName, strings.Join(param.Values, ", "))
				}
			}
		}
	}

	return nil
}
