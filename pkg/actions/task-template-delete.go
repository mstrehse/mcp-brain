package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskTemplateDeleteHandler creates a handler for deleting task templates
func NewTaskTemplateDeleteHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := request.RequireString("template_id")
		if err != nil {
			return mcp.NewToolResultError("Missing 'template_id' parameter: " + err.Error()), nil
		}

		// Check if template exists first
		_, err = repo.GetTemplate(templateID)
		if err != nil {
			return mcp.NewToolResultError("Template not found: " + err.Error()), nil
		}

		if err := repo.DeleteTemplate(templateID); err != nil {
			return mcp.NewToolResultError("Failed to delete template: " + err.Error()), nil
		}

		result := map[string]interface{}{
			"message":     "Template deleted successfully",
			"template_id": templateID,
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
