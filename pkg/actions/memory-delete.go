package actions

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewMemoryDeleteHandler creates a handler for deleting knowledge with dependency injection
func NewMemoryDeleteHandler(repo contracts.KnowledgeRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError("Missing 'path' parameter: " + err.Error()), nil
		}
		if err := repo.Delete(path); err != nil {
			return mcp.NewToolResultError("Failed to delete file: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Memory deleted successfully."), nil
	}
}
