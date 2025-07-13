package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTaskGetHandler creates a handler for getting tasks with dependency injection
func NewTaskGetHandler(repo contracts.TaskRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		task, err := repo.GetTask()
		if err != nil {
			return mcp.NewToolResultError("Failed to get task: " + err.Error()), nil
		}

		data, err := json.Marshal(task)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal task result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
