package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mstrehse/mcp-brain/pkg/actions"
)

//go:embed brain-mcp-description.md
var serverDescription string

func main() {
	// Initialize the knowledge repository with user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	// Create repositories with proper dependency injection
	repositories, err := actions.NewRepositories(filepath.Join(homeDir, ".mcp-brain"))
	if err != nil {
		fmt.Printf("Error initializing repositories: %v\n", err)
		return
	}

	// Ensure database is closed when program exits
	defer func() {
		if err := repositories.Close(); err != nil {
			fmt.Printf("Error closing repositories: %v\n", err)
		}
	}()

	// Create a new MCP server with embedded description
	s := server.NewMCPServer(
		serverDescription,
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add store-memory tool
	storeMemoryTool := mcp.NewTool("store-memory",
		mcp.WithDescription("Store information as a markdown file in the user's brain for a specific project. IMPORTANT: Before storing new information, always use 'list-memories' to check what already exists and 'get-memory' to review existing content to avoid duplication or conflicts. Use this to persist knowledge, notes, or context for later retrieval. Optimized for LLM workflows. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("The name of the project (usually the folder name) to store the memory under."),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path (can include subfolders) for the markdown file inside the project. Do not use absolute paths or '..'."),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The markdown content to store."),
		),
	)

	getMemoryTool := mcp.NewTool("get-memory",
		mcp.WithDescription("Retrieve information from a markdown file in the user's brain for a specific project. CRITICAL: Always use this tool to check for existing knowledge before making assumptions or creating new content. This prevents duplication and ensures you have the complete context. Use this to recall previously stored knowledge or notes. Optimized for LLM workflows. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("The name of the project (usually the folder name) to retrieve the memory from."),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path (can include subfolders) for the markdown file inside the project. Do not use absolute paths or '..'."),
		),
	)

	// Add delete-memory tool
	deleteMemoryTool := mcp.NewTool("delete-memory",
		mcp.WithDescription("Delete a markdown memory file in the user's brain for a specific project. CAUTION: Only use this tool when you're certain the information is no longer needed or when replacing outdated information. Always check the content with 'get-memory' before deleting to ensure you're not removing valuable knowledge. Use this to remove knowledge, notes, or context that is no longer needed. Optimized for LLM workflows. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("The name of the project (usually the folder name) to delete the memory from."),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path (can include subfolders) for the markdown file inside the project. Do not use absolute paths or '..'."),
		),
	)

	// Add ask-user tool
	askUserTool := mcp.NewTool("ask-user",
		mcp.WithDescription("Ask the user a question via a popup dialog. The user's answer is returned. Works on GNOME (Linux) and OSX. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("question",
			mcp.Required(),
			mcp.Description("The question to ask the user."),
		),
	)

	// Add list-memories tool
	listMemoriesTool := mcp.NewTool("list-memories",
		mcp.WithDescription("Get a hierarchical structure of all memories (markdown files and folders) for a specific project. START HERE: Always use this tool first when working with a project to understand what knowledge already exists. These memories are an important source of information for the system and user. Use this to understand the available knowledge and its organization before making any assumptions about what needs to be done. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("The name of the project (usually the folder name) to list memories for."),
		),
	)

	// Add task management tools
	addTasksTool := mcp.NewTool("add-tasks",
		mcp.WithDescription("Add multiple tasks to the queue for the current chat session. WORKFLOW PATTERN: When facing complex work, immediately break it down into specific tasks using this tool. Create a complete task list upfront, then use 'get-task' to retrieve and complete them one by one. This ensures systematic completion and prevents missing important steps. This is mandatory - tasks should always be created for future work. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("chat_session_id",
			mcp.Required(),
			mcp.Description("The ID of the current chat session."),
		),
		mcp.WithArray("contents",
			mcp.Required(),
			mcp.Description("Array of task descriptions to add."),
		),
	)

	getTaskTool := mcp.NewTool("get-task",
		mcp.WithDescription("Retrieve and remove the next pending task from the queue for the current chat session. SYSTEMATIC WORKFLOW: After completing each task, immediately call this tool to get the next task. This ensures you work through your task list systematically and don't miss any steps. Continue calling this tool until you get 'no pending tasks' - only then is your work complete. This is mandatory - always check for remaining tasks before considering work complete. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("chat_session_id",
			mcp.Required(),
			mcp.Description("The ID of the current chat session."),
		),
	)

	// Create actions with dependency injection
	askAction := actions.NewAskAction()

	// Register tools with dependency-injected handlers
	s.AddTool(storeMemoryTool, actions.NewCreateKnowledgeHandler(repositories.Knowledge))
	s.AddTool(getMemoryTool, actions.NewReadKnowledgeHandler(repositories.Knowledge))
	s.AddTool(deleteMemoryTool, actions.NewDeleteKnowledgeHandler(repositories.Knowledge))
	s.AddTool(askUserTool, askAction.AskUser)
	s.AddTool(listMemoriesTool, actions.NewListKnowledgeHandler(repositories.Knowledge))
	s.AddTool(addTasksTool, actions.NewAddTasksHandler(repositories.Task))
	s.AddTool(getTaskTool, actions.NewGetTaskHandler(repositories.Task))

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
