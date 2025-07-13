package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskTemplateGetHandler creates a handler for getting a specific template
func NewTaskTemplateGetHandler(repo contracts.TaskTemplateRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
