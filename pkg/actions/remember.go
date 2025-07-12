package actions

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewCreateKnowledgeHandler creates a handler for storing knowledge with dependency injection
func NewCreateKnowledgeHandler(repo contracts.KnowledgeRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		project, err := request.RequireString("project")
		if err != nil {
			return mcp.NewToolResultError("Missing 'project' parameter: " + err.Error()), nil
		}
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError("Missing 'path' parameter: " + err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("Missing 'content' parameter: " + err.Error()), nil
		}
		if err := repo.Write(project, path, content); err != nil {
			return mcp.NewToolResultError("Failed to write file: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Memory stored successfully."), nil
	}
}
