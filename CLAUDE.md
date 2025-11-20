# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an MCP (Model Context Protocol) server that manages isolated git worktrees for AI coding agents. It allows each agent (Copilot, Claude, GPT, Gemini, etc.) to work in their own worktree, run tests in isolation, and merge approved changes back to the main branch.


## Architecture

### Core Concepts

1. **Agent Isolation**: Each agent gets its own git worktree and branch (pattern: `agent-{agentId}`)
2. **Worktree Management**: Manages creation, testing, and merging of agent-specific worktrees
3. **Test Execution**: Runs configurable test commands in isolated agent worktrees
4. **Merge Control**: Merges agent branches back to base branch with conflict handling

### Technology Stack

- **Language**: Go
- **Architecture**: Clean Architecture with domain, application, and infrastructure layers
- **Git Integration**: Shell out to `git` commands for worktrees, commits, merges
- **Interface**: MCP server exposing tools + optional REST/HTTP endpoints for debugging
- **Configuration**: File-based config (YAML/JSON) for repo path, base branch, test command
- **State**: In-memory repository with plans for SQLite or file-based storage

### Package Structure

```
internal/
├── domain/               # Core business logic (zero external dependencies)
│   ├── agent.go         # Agent aggregate root with lifecycle management
│   ├── agent_id.go      # AgentID value object with validation
│   └── ports.go         # Domain interfaces (GitOperations, AgentRepository)
│
├── application/          # Use cases and orchestration
│   └── create_worktree.go  # CreateWorktreeUseCase implementation
│
├── adapters/             # Inbound adapters (external protocols → application)
│   └── mcp/
│       ├── server.go         # MCP server adapter
│       ├── result.go         # MCP result helpers
│       └── server_test.go    # MCP server tests
│
└── infrastructure/       # Outbound adapters (application → external services)
    ├── git/
    │   └── git_client.go      # GitClient implementing GitOperations
    └── persistence/
        └── in_memory_repository.go  # InMemoryAgentRepository

cmd/
└── server/
    └── main.go          # Entry point and test harness
```

**Layer Dependencies** (strictly enforced):
- Domain: No dependencies (pure business logic)
- Application: Depends only on domain interfaces
- Adapters: Translate external protocols to application use cases (inbound)
- Infrastructure: Implements domain interfaces, depends on external libraries (outbound)
- Flow: Adapters/Infrastructure → Application → Domain

## Key Design Decisions

1. **In-Memory State**: Current implementation uses in-memory storage; server restart loses all state
2. **Manual Cleanup**: Worktrees and branches must be removed manually (no automatic cleanup)
3. **Single Repository**: One configured repository per server instance
4. **Local Merges**: Merges stay local; manual `git push` required to sync with remote
5. **Simple Status Model**: Three agent status values (`created`, `merged`, `failed`)
6. **Clean Architecture**: Strict layer separation (domain → application → infrastructure)
7. **Value Objects**: AgentID enforces validation rules at domain level

## Working Directory (CRITICAL)

**Your workspace is the active working directory. Do not leave it.**

- Inspect `<env>` to determine your working directory and the current branch
- If you are in a worktree (branch: `agent-manager-mcp/*`): edit all files here
- If you are in the main repository (branch: `main`): make changes here directly
- Do NOT infer or rely on parent paths from the directory structure
- Do NOT switch directories with `cd` to other locations unless explicitly required

❌ WRONG: `cd /inferred/path && command`
✅ RIGHT: `command` (executed in the current directory)

## Implementation Rules (CRITICAL)

### Testing Code

**Test Structure (MANDATORY)**
- Use descriptive test names that clearly describe the scenario being tested
- Every test function MUST contain explicit comment blocks:
  ```go
  // arrange
  // ... setup code

  // act
  // ... execution code

  // assert
  // ... verification code
  ```
- Follow table-driven test patterns for multiple scenarios
- Use test helpers to reduce duplication in setup/teardown

**Test Naming Convention**
- Format: `Test{Component}_{Scenario}_{ExpectedBehavior}`
- Examples:
  - `TestAgentID_CreateWithInvalidFormat_ReturnsError`
  - `TestCreateWorktreeUseCase_WhenAgentAlreadyExists_ReturnsError`
  - `TestGitClient_CreateWorktree_CreatesDirectoryAndBranch`

### Code Quality

**Non-Deterministic Solutions PROHIBITED**
- NO timeouts, delays, sleep (e.g., `setTimeout`, `sleep`) in application logic or test code.
  - This restriction does not apply to operational safeguards like wrapping long-running terminal commands
    with a timeout to prevent the CLI from hanging during manual workflows.
- NO retry loops, polling (especially `setInterval` for state sync!)
- NO timing-based solutions
- These approaches are unreliable, hard to maintain, and behave inconsistently across different environments

**Error Handling (MANDATORY)**
- NEVER use empty catch blocks
- Always log with context
- Provide actionable information

**Naming Conventions (MANDATORY)**
- Use descriptive variable names, not abbreviations
- Examples:
  - ✅ `agentRepository`, `worktreePath`, `testCommand`
  - ❌ `repo`, `path`, `cmd`
- Exception: Well-established conventions in limited scopes (e.g., `i` in loops, `err` for errors, `ctx` for context)
- Function names should clearly describe their action and intent
- Type names should be self-explanatory without needing comments

### Comment Style (MANDATORY)
- Do not use comments to narrate what changed or what is new.
- Prefer self-documenting code; only add comments when strictly necessary to explain WHY (intent/rationale), not WHAT.
- Keep any necessary comments concise and local to the logic they justify.

## Development Workflow

1. Make changes
2. Run test suite
3. Only commit when all checks pass

## Testing Requirements

### TDD (MANDATORY)
Always write tests first, before implementing features:
1. **Red**: Write a failing test that describes the desired behavior
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve the implementation while keeping tests green

## Specification Writing Guidelines

### Technical Specs (MANDATORY)
When creating specs for implementation agents:
- **Focus**: Technical implementation details, architecture, code examples
- **Requirements**: Clear dependencies, APIs, integration points
- **Structure**: Components → Implementation → Configuration → Phases
- **Omit**: Resource constraints, obvious details, verbose explanations
- **Include**: Platform-specific APIs, code snippets, data flows, dependencies

**CRITICAL Rules:**
- Test failures are NEVER unrelated - fix immediately
- NEVER skip tests (no `.skip()`, `xit()`)
- Fix performance test failures (they indicate real issues)
- After every code change, the responsible agent must rerun the full validation suite and report "tests green" before handing the work back. Only proceed with known failing tests when the user explicitly permits leaving the suite red for that task.

## Plan Files

- Store all plan MD files in the `./docs/plans/` directory, not at the repository root
- This keeps the root clean and organizes planning documents
- If you create plans research the codebase or requested details first before making a plan for the implementation
- Don't make plans for making plans, rather do the planning ahead and then implement