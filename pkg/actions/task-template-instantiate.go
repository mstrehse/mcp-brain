package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskTemplateInstantiateHandler creates a handler for instantiating templates
func NewTaskTemplateInstantiateHandler(repo contracts.TaskTemplateRepository, taskRepo contracts.TaskRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := request.RequireString("template_id")
		if err != nil {
			return mcp.NewToolResultError("Missing 'template_id' parameter: " + err.Error()), nil
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
		addedTasks, err := taskRepo.AddTasks(instance.Tasks)
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
