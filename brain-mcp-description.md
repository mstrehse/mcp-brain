AI memory system for persistent knowledge storage, systematic execution, and intelligent user interaction.

## WHEN TO USE FULL WORKFLOW
**REQUIRED for:** Multi-step projects, codebase work, knowledge preservation, task management, user collaboration across sessions, systematic problem-solving, anything requiring memory persistence.

**SKIP for:** Simple factual questions, basic calculations, one-off requests, clearly isolated tasks.

## CORE WORKFLOW (MANDATORY WHEN TRIGGERED)
```
1. DISCOVER: list-memories → get-memory (review existing)
2. PLAN: add-tasks (break down work systematically)  
3. EXECUTE: get-task → work → store-memory → repeat until "no pending tasks"
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

### User Interaction
- **`ask-user`**(question) - Professional popup dialogs (Linux/OSX)

## KEY PATTERNS
- **Always** start with `list-memories` to understand existing context
- **Always** use `ask-upser` to get feedback if you are not 100% sure or when there multiple options to go on
- **Always** use `add-tasks` for complex work breakdown
- **Continue** `get-task` calls until "no pending tasks" 
- **Store** valuable insights with `store-memory`
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