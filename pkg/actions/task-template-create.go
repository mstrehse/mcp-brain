package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskTemplateCreateHandler creates a handler for creating new task templates
func NewTaskTemplateCreateHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
