package actions

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// NewAddTasksHandler creates a handler for adding tasks with dependency injection
func NewAddTasksHandler(repo contracts.TaskRepository) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		chatSessionID, err := request.RequireString("chat_session_id")
		if err != nil {
			return mcp.NewToolResultError("Missing 'chat_session_id' parameter: " + err.Error()), nil
		}

		contents, err := request.RequireStringSlice("contents")
		if err != nil {
			return mcp.NewToolResultError("Missing 'contents' parameter: " + err.Error()), nil
		}

		if len(contents) == 0 {
			return mcp.NewToolResultError("Contents array cannot be empty"), nil
		}

		tasks, err := repo.AddTasks(chatSessionID, contents)
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
