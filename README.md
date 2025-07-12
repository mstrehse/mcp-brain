# Brain MCP Server 🧠

A comprehensive memory and task management system designed for LLM workflows in code editors that support MCP (Model Context Protocol) like Cursor, Claude Desktop, and others.

## ⚠️ Development Status

**This project is under active development and not yet ready for production use.**

- 🚧 APIs may change without notice
- 🧪 Some features are experimental
- 📋 Documentation may be incomplete or outdated
- 🐛 Bugs and breaking changes are expected

Please use with caution and expect potential issues. Contributions and feedback are welcome!

## Features

- **🧠 Knowledge Management**: Store, retrieve, and organize markdown files by project
- **📋 Task Management**: Systematic task queue for complex workflow execution
- **🎯 Template System**: Create and use reusable workflow templates with parameters
- **💬 User Interaction**: Professional popup dialogs for user questions (Linux/OSX)
- **🔄 Persistent Storage**: SQLite-based storage in user's home directory

## Installation

### Prerequisites

- Go 1.24.2 or higher

### Install

```bash
# Install Go (if not already installed)
# Visit https://golang.org/dl/ for installation instructions

# Install the Brain MCP server
go install github.com/mstrehse/mcp-brain@latest
```

## Configuration

### Cursor IDE

Add this to your Cursor settings (`.cursor/mcp_servers.json` or through Settings > MCP):

```json
{
  "mcpServers": {
    "brain": {
      "command": "mcp-brain",
      "args": []
    }
  }
}
```

### Claude Desktop

Add this to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS or `%APPDATA%/Claude/claude_desktop_config.json` on Windows):

```json
{
  "mcpServers": {
    "brain": {
      "command": "mcp-brain",
      "args": []
    }
  }
}
```

### Other MCP-Compatible Editors

For any editor that supports MCP, configure it to run the `mcp-brain` command. The server communicates via stdin/stdout using the MCP protocol.

## Usage

Once configured, the Brain MCP server provides the following tools in your LLM conversations:

### Knowledge Management

- **`store-memory`**: Store information as markdown files organized by project
- **`get-memory`**: Retrieve previously stored knowledge
- **`list-memories`**: Get hierarchical structure of all memories for a project
- **`delete-memory`**: Remove outdated or incorrect information

### Task Management

- **`add-tasks`**: Add multiple tasks to the queue for systematic execution
- **`get-task`**: Retrieve and remove the next pending task from the queue

### Template Management

- **`list-task-templates`**: Discover available reusable workflow templates
- **`get-task-template`**: Get detailed information about a specific template
- **`create-task-template`**: Create new reusable task workflow templates
- **`instantiate-task-template`**: Generate tasks from templates with specific parameters

### User Interaction

- **`ask-user`**: Ask users questions via popup dialogs (Linux/OSX)

## License

This project is licensed under the GPL3 License - see the [LICENSE](LICENSE) file for details.

## Workflow Integration

This MCP server is designed to integrate with LLM workflows by providing:

1. **Systematic Task Execution**: Break complex work into manageable tasks
2. **Knowledge Persistence**: Build institutional memory across sessions
3. **Template-Based Workflows**: Reusable task patterns with parameter substitution
4. **User Interaction**: Popup dialogs instead of chat-based questions
5. **Context Preservation**: Maintain understanding across interruptions

### Recommended Usage Pattern

1. **Discover**: Use `list-memories` to understand existing knowledge
2. **Plan**: Use `list-task-templates` and `instantiate-task-template` for proven workflows, or `add-tasks` for new work patterns
3. **Execute**: Use `get-task` to work through tasks one by one
4. **Store**: Use `store-memory` to preserve valuable insights
5. **Capture**: Use `create-task-template` to save successful workflows for reuse

## Support

For issues, questions, or contributions, please open an issue on GitHub.

---

Built with ❤️ using Go and the MCP protocol framework. 