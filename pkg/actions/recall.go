package actions

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewReadKnowledgeHandler creates a handler for reading knowledge with dependency injection
func NewReadKnowledgeHandler(repo contracts.KnowledgeRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		project, err := request.RequireString("project")
		if err != nil {
			return mcp.NewToolResultError("Missing 'project' parameter: " + err.Error()), nil
		}
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError("Missing 'path' parameter: " + err.Error()), nil
		}
		content, err := repo.Read(project, path)
		if err != nil {
			return mcp.NewToolResultError("Failed to read file: " + err.Error()), nil
		}
		return mcp.NewToolResultText(content), nil
	}
}
