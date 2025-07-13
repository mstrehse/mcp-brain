package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskTemplateUpdateHandler creates a handler for updating existing task templates
func NewTaskTemplateUpdateHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateJSON, err := request.RequireString("template")
		if err != nil {
			return mcp.NewToolResultError("Missing 'template' parameter: " + err.Error()), nil
		}

		var template contracts.TaskTemplate
		if err := json.Unmarshal([]byte(templateJSON), &template); err != nil {
			return mcp.NewToolResultError("Invalid template JSON: " + err.Error()), nil
		}

		if template.ID == "" {
			return mcp.NewToolResultError("Template ID is required for update"), nil
		}

		// Check if template exists
		existing, err := repo.GetTemplate(template.ID)
		if err != nil {
			return mcp.NewToolResultError("Template not found: " + err.Error()), nil
		}

		// Preserve creation timestamp
		template.CreatedAt = existing.CreatedAt

		if err := validateTemplate(&template); err != nil {
			return mcp.NewToolResultError("Template validation failed: " + err.Error()), nil
		}

		if err := repo.UpdateTemplate(&template); err != nil {
			return mcp.NewToolResultError("Failed to update template: " + err.Error()), nil
		}

		result := map[string]interface{}{
			"message":     "Template updated successfully",
			"template_id": template.ID,
			"updated_at":  template.UpdatedAt,
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
