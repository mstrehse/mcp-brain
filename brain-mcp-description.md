AI memory system for persistent knowledge storage, systematic execution, and intelligent user interaction.

## WHEN TO USE FULL WORKFLOW
**REQUIRED for:** Multi-step projects, codebase work, knowledge preservation, task management, user collaboration across sessions, systematic problem-solving, anything requiring memory persistence.

**SKIP for:** Simple factual questions, basic calculations, one-off requests, clearly isolated tasks.

## CORE WORKFLOW (MANDATORY WHEN TRIGGERED)
```
1. DISCOVER: memories-list → memory-get (review existing)
2. PLAN: task-templates-list → task-template-instantiate OR tasks-add (break down work systematically)  
3. EXECUTE: task-get → work → memory-store → repeat until "no pending tasks"
4. CAPTURE: task-template-create (for reusable workflows)
```

## TOOLS

### Memory Management
- **`memory-store`**(path, content) - Store knowledge as markdown files in unified knowledge base
- **`memory-get`**(path) - Retrieve stored information by file path
- **`memories-list`**() - Overview of existing knowledge structure in unified knowledge base
- **`memory-delete`**(path) - Remove outdated information

### Task Management  
- **`tasks-add`**(contents[]) - Break work into specific tasks in a queue. Adds them at the end
- **`task-get`**() - Get next task systematically from queue

### Template Management
- **`task-templates-list`**(category?) - Discover reusable workflow templates
- **`task-template-get`**(template_id) - Get template details and parameters
- **`task-template-create`**(template) - Create reusable task workflows
- **`task-template-update`**(template) - Update existing template with new parameters/tasks
- **`task-template-delete`**(template_id) - Delete template permanently (use with caution)
- **`task-template-instantiate`**(template_id, parameters?) - Generate tasks from templates

### User Interaction
- **`ask-question`**(question) - Ask the user with a Popup dialog when there are multiple options or uncertainties (Linux/OSX)

## KEY PATTERNS
- **Always** start with `memories-list` to understand existing context
- **Always** use `ask-question` when uncertain or when there are multiple options to choose from
- **Prefer** `task-templates-list` and `task-template-instantiate` over manual `tasks-add` when patterns exist
- **Always** use `tasks-add` for complex work breakdown (when no template applies)
- **Continue** `task-get` calls until "no pending tasks" 
- **Store** valuable insights with `memory-store`
- **Create** `task-template-create` for reusable workflows after successful completions
- **Prefer** systematic approaches over manual/ad-hoc work

## CRITICAL BEHAVIORS
✅ Check existing memories before creating new content
✅ Break complex work into tasks systematically
✅ Ask questions when there are multiple options or uncertainties
✅ Use templates when available for proven workflows
✅ Always complete task queues fully
✅ Store valuable insights for future reference
✅ Create templates from successful patterns 