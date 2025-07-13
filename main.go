package main

import (
	_ "embed"
	"flag"
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
	// Define command line flags
	brainDir := flag.String("brain-dir", "", "Directory to store brain data (defaults to ./.brain)")
	flag.Parse()

	// Determine the base directory
	var baseDir string
	if *brainDir != "" {
		baseDir = *brainDir
	} else {
		// Default to ./.brain (current working directory)
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			return
		}
		baseDir = filepath.Join(cwd, ".brain")
	}

	// Create repositories with proper dependency injection
	repositories, err := actions.NewRepositories(baseDir)
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

	// Add memory-store tool
	memoryStoreTool := mcp.NewTool("memory-store",
		mcp.WithDescription("Store information as a markdown file in the user's brain for a specific project. IMPORTANT: Before storing new information, always use 'memories-list' to check what already exists and 'memory-get' to review existing content to avoid duplication or conflicts. Use this to persist knowledge, notes, or context for later retrieval. Optimized for LLM workflows. Always use the full functionality of this tool and its parameters."),
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

	memoryGetTool := mcp.NewTool("memory-get",
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

	// Add memory-delete tool
	memoryDeleteTool := mcp.NewTool("memory-delete",
		mcp.WithDescription("Delete a markdown memory file in the user's brain for a specific project. CAUTION: Only use this tool when you're certain the information is no longer needed or when replacing outdated information. Always check the content with 'memory-get' before deleting to ensure you're not removing valuable knowledge. Use this to remove knowledge, notes, or context that is no longer needed. Optimized for LLM workflows. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("The name of the project (usually the folder name) to delete the memory from."),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path (can include subfolders) for the markdown file inside the project. Do not use absolute paths or '..'."),
		),
	)

	// Add ask-question tool
	askQuestionTool := mcp.NewTool("ask-question",
		mcp.WithDescription("Ask the user a question via a popup dialog. The user's answer is returned. Works on GNOME (Linux) and OSX. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("question",
			mcp.Required(),
			mcp.Description("The question to ask the user."),
		),
	)

	// Add memories-list tool
	memoriesListTool := mcp.NewTool("memories-list",
		mcp.WithDescription("Get a hierarchical structure of all memories (markdown files and folders) for a specific project. START HERE: Always use this tool first when working with a project to understand what knowledge already exists. These memories are an important source of information for the system and user. Use this to understand the available knowledge and its organization before making any assumptions about what needs to be done. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("The name of the project (usually the folder name) to list memories for."),
		),
	)

	// Add task management tools
	tasksAddTool := mcp.NewTool("tasks-add",
		mcp.WithDescription("Add multiple tasks to the queue for the current chat session. WORKFLOW PATTERN: When facing complex work, immediately break it down into specific tasks using this tool. Create a complete task list upfront, then use 'task-get' to retrieve and complete them one by one. This ensures systematic completion and prevents missing important steps. This is mandatory - tasks should always be created for future work. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("chat_session_id",
			mcp.Required(),
			mcp.Description("The ID of the current chat session."),
		),
		mcp.WithArray("contents",
			mcp.Required(),
			mcp.Description("Array of task descriptions to add."),
		),
	)

	taskGetTool := mcp.NewTool("task-get",
		mcp.WithDescription("Retrieve and remove the next pending task from the queue for the current chat session. SYSTEMATIC WORKFLOW: After completing each task, immediately call this tool to get the next task. This ensures you work through your task list systematically and don't miss any steps. Continue calling this tool until you get 'no pending tasks' - only then is your work complete. This is mandatory - always check for remaining tasks before considering work complete. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("chat_session_id",
			mcp.Required(),
			mcp.Description("The ID of the current chat session."),
		),
	)

	// Add template management tools
	taskTemplatesListTool := mcp.NewTool("task-templates-list",
		mcp.WithDescription("List all available task templates. DISCOVERY PATTERN: Use this tool to discover reusable workflows and task patterns. Templates provide structured approaches to common work like code reviews, bug fixes, research, and development tasks. Start with this tool to see what templates are available before creating manual task lists. Always use the full functionality of this tool and its parameters."),
	)

	taskTemplateGetTool := mcp.NewTool("task-template-get",
		mcp.WithDescription("Retrieve detailed information about a specific task template, including its parameters and task structure. INSPECTION PATTERN: Use this tool to understand what parameters a template requires and preview the tasks it will create. This helps you gather the right information before instantiating the template. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("The ID of the template to retrieve."),
		),
	)

	taskTemplateCreateTool := mcp.NewTool("task-template-create",
		mcp.WithDescription("Create a new reusable task template with parameters and task patterns. PATTERN CREATION: Use this tool to capture successful workflows as reusable templates. Define parameters using ${param} syntax in task descriptions for dynamic content. This builds institutional knowledge and accelerates future similar work. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("template",
			mcp.Required(),
			mcp.Description("JSON representation of the task template structure."),
		),
	)

	taskTemplateInstantiateTool := mcp.NewTool("task-template-instantiate",
		mcp.WithDescription("Create tasks from a template with specific parameters and add them to the current chat session. WORKFLOW ACCELERATION: Use this tool to quickly set up structured workflows from proven templates. The template parameters will be resolved and tasks added to your queue automatically. This is the preferred way to start complex work - templates over manual task creation. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("The ID of the template to instantiate."),
		),
		mcp.WithString("chat_session_id",
			mcp.Required(),
			mcp.Description("The ID of the current chat session."),
		),
		mcp.WithString("parameters",
			mcp.Description("JSON object containing parameter values for the template."),
		),
	)

	taskTemplateUpdateTool := mcp.NewTool("task-template-update",
		mcp.WithDescription("Update an existing task template with new parameters, tasks, or metadata. TEMPLATE MANAGEMENT: Use this tool to refine and improve existing templates based on experience. Always include the template ID in the template JSON to specify which template to update. This maintains template evolution and continuous improvement. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("template",
			mcp.Required(),
			mcp.Description("JSON representation of the updated task template structure including the ID."),
		),
	)

	taskTemplateDeleteTool := mcp.NewTool("task-template-delete",
		mcp.WithDescription("Delete a task template by ID. CAUTION: This permanently removes the template and cannot be undone. Use this tool to clean up obsolete or incorrect templates. Always verify the template ID before deletion. This helps maintain a clean template library. Always use the full functionality of this tool and its parameters."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("The ID of the template to delete."),
		),
	)

	// Create actions with dependency injection
	askQuestionAction := actions.NewAskQuestionAction()

	// Register tools with dependency-injected handlers
	s.AddTool(memoryStoreTool, actions.NewMemoryStoreHandler(repositories.Knowledge))
	s.AddTool(memoryGetTool, actions.NewMemoryGetHandler(repositories.Knowledge))
	s.AddTool(memoryDeleteTool, actions.NewMemoryDeleteHandler(repositories.Knowledge))
	s.AddTool(askQuestionTool, askQuestionAction.AskQuestion)
	s.AddTool(memoriesListTool, actions.NewMemoriesListHandler(repositories.Knowledge))
	s.AddTool(tasksAddTool, actions.NewTasksAddHandler(repositories.Task))
	s.AddTool(taskGetTool, actions.NewTaskGetHandler(repositories.Task))
	s.AddTool(taskTemplatesListTool, actions.NewTaskTemplatesListHandler(repositories.Template))
	s.AddTool(taskTemplateGetTool, actions.NewTaskTemplateGetHandler(repositories.Template))
	s.AddTool(taskTemplateCreateTool, actions.NewTaskTemplateCreateHandler(repositories.Template))
	s.AddTool(taskTemplateInstantiateTool, actions.NewTaskTemplateInstantiateHandler(repositories.Template, repositories.Task))
	s.AddTool(taskTemplateUpdateTool, actions.NewTaskTemplateUpdateHandler(repositories.Template))
	s.AddTool(taskTemplateDeleteTool, actions.NewTaskTemplateDeleteHandler(repositories.Template))

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
