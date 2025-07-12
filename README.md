# Brain MCP Server 🧠

A comprehensive memory and task management system designed for LLM workflows in code editors that support MCP (Model Control Protocol) like Cursor, Claude Desktop, and others.

## Features

- **🧠 Knowledge Management**: Store, retrieve, and organize markdown files by project
- **📋 Task Management**: Systematic task queue for complex workflow execution
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

### User Interaction

- **`ask-user`**: Ask users questions via popup dialogs (Linux/OSX)

### Data Storage

The server stores data in `~/.mcp-brain/` directory:

- **Knowledge files**: Organized by project as markdown files
- **SQLite database**: For task queues and metadata
- **Automatic cleanup**: Proper database connection handling

## License

This project is licensed under the GPL3 License - see the [LICENSE](LICENSE) file for details.

## Workflow Integration

This MCP server is designed to integrate with LLM workflows by providing:

1. **Systematic Task Execution**: Break complex work into manageable tasks
2. **Knowledge Persistence**: Build institutional memory across sessions
3. **Professional User Interaction**: Popup dialogs instead of chat-based questions
4. **Context Preservation**: Maintain understanding across interruptions

### Recommended Usage Pattern

1. **Discover**: Use `list-memories` to understand existing knowledge
2. **Plan**: Use `add-tasks` to break down complex work systematically
3. **Execute**: Use `get-task` to work through tasks one by one
4. **Store**: Use `store-memory` to preserve valuable insights

## Support

For issues, questions, or contributions, please open an issue on GitHub.

---

Built with ❤️ using Go and the MCP protocol framework. 