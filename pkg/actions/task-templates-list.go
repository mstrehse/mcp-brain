package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskTemplatesListHandler creates a handler for listing task templates
func NewTaskTemplatesListHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
