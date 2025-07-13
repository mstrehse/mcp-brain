package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewMemoriesListHandler creates a handler for listing knowledge with dependency injection
func NewMemoriesListHandler(repo contracts.KnowledgeRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dirStructure, err := repo.List()
		if err != nil {
			return mcp.NewToolResultError("Failed to list memories: " + err.Error()), nil
		}

		data, err := json.Marshal(dirStructure)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal result: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}
