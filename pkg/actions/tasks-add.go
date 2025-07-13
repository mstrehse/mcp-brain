package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewTasksAddHandler creates a handler for adding tasks with dependency injection
func NewTasksAddHandler(repo contracts.TaskRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contents, err := request.RequireStringSlice("contents")
		if err != nil {
			return mcp.NewToolResultError("Missing 'contents' parameter: " + err.Error()), nil
		}

		if len(contents) == 0 {
			return mcp.NewToolResultError("Contents array cannot be empty"), nil
		}

		tasks, err := repo.AddTasks(contents)
		if err != nil {
			return mcp.NewToolResultError("Failed to add tasks: " + err.Error()), nil
		}

		result := map[string]interface{}{
			"tasks_added": len(tasks),
			"tasks":       tasks,
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal tasks result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
