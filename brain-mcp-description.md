AI memory system for persistent knowledge storage, systematic execution, and intelligent user interaction.

## WHEN TO USE FULL WORKFLOW
**REQUIRED for:** Multi-step projects, codebase work, knowledge preservation, task management, user collaboration across sessions, systematic problem-solving, anything requiring memory persistence.

**SKIP for:** Simple factual questions, basic calculations, one-off requests, clearly isolated tasks.

## CORE WORKFLOW (MANDATORY WHEN TRIGGERED)
```
1. DISCOVER: list-memories → get-memory (review existing)
2. PLAN: list-task-templates → instantiate-task-template OR add-tasks (break down work systematically)  
3. EXECUTE: get-task → work → store-memory → repeat until "no pending tasks"
4. CAPTURE: create-task-template (for reusable workflows)
```

## TOOLS

### Memory Management
- **`store-memory`**(project, path, content) - Store knowledge as markdown files
- **`get-memory`**(project, path) - Retrieve stored information  
- **`list-memories`**(project) - Overview of existing knowledge structure
- **`delete-memory`**(project, path) - Remove outdated information

### Task Management  
- **`add-tasks`**(chat_session_id, contents[]) - Break work into specific tasks
- **`get-task`**(chat_session_id) - Get next task systematically

### Template Management
- **`list-task-templates`**(category?) - Discover reusable workflow templates
- **`get-task-template`**(template_id) - Get template details and parameters
- **`create-task-template`**(template) - Create reusable task workflows
- **`instantiate-task-template`**(template_id, chat_session_id, parameters?) - Generate tasks from templates

### User Interaction
- **`ask-user`**(question) - Professional popup dialogs (Linux/OSX)

## KEY PATTERNS
- **Always** start with `list-memories` to understand existing context
- **Always** use `ask-user` to get feedback if you are not 100% sure or when there multiple options to go on
- **Prefer** `list-task-templates` and `instantiate-task-template` over manual `add-tasks` when patterns exist
- **Always** use `add-tasks` for complex work breakdown (when no template applies)
- **Continue** `get-task` calls until "no pending tasks" 
- **Store** valuable insights with `store-memory`
- **Create** `create-task-template` for reusable workflows after successful completions
- **Prefer** systematic approaches over manual/ad-hoc work

## CRITICAL BEHAVIORS
✅ Check existing memories before creating new content
✅ Break complex work into systematic tasks upfront
✅ Complete all tasks before considering work finished
✅ Store reusable knowledge for future sessions
❌ Skip memory discovery phase
❌ Work without task breakdown
❌ Stop before "no pending tasks"

**PRIORITY**: Use these tools over manual approaches for any work requiring memory, organization, or user interaction. 