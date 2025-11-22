package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tzDel/orchestrAIgent/internal/application"
)

type CreateWorktreeArgs struct {
	AgentID string `json:"agentId" jsonschema:"required" jsonschema_description:"The unique identifier for the agent"`
}

type CreateWorktreeOutput struct {
	AgentID      string `json:"agentId"`
	WorktreePath string `json:"worktreePath"`
	BranchName   string `json:"branchName"`
	Status       string `json:"status"`
}

type RemoveWorktreeArgs struct {
	AgentID string `json:"agentId" jsonschema:"required" jsonschema_description:"Agent identifier"`
	Force   bool   `json:"force" jsonschema_description:"Skip safety checks and force removal"`
}

type RemoveWorktreeOutput struct {
	AgentID            string `json:"agentId"`
	RemovedAt          string `json:"removedAt,omitempty"`
	HasUnmergedChanges bool   `json:"hasUnmergedChanges"`
	UnmergedCommits    int    `json:"unmergedCommits"`
	UncommittedFiles   int    `json:"uncommittedFiles"`
	Warning            string `json:"warning,omitempty"`
}

type MCPServer struct {
	mcpServer             *mcpsdk.Server
	createWorktreeUseCase *application.CreateWorktreeUseCase
	removeWorktreeUseCase *application.RemoveWorktreeUseCase
}

func NewMCPServer(
	createWorktreeUseCase *application.CreateWorktreeUseCase,
	removeWorktreeUseCase *application.RemoveWorktreeUseCase,
) (*MCPServer, error) {
	impl := &mcpsdk.Implementation{
		Name:    "orchestrAIgent",
		Version: "0.1.0",
	}

	mcpServer := mcpsdk.NewServer(impl, nil)

	server := &MCPServer{
		mcpServer:             mcpServer,
		createWorktreeUseCase: createWorktreeUseCase,
		removeWorktreeUseCase: removeWorktreeUseCase,
	}

	mcpsdk.AddTool(
		mcpServer,
		&mcpsdk.Tool{
			Name:        "create_worktree",
			Description: "Creates an isolated git worktree for a specific agent with its own branch",
		},
		server.handleCreateWorktree,
	)

	mcpsdk.AddTool(
		mcpServer,
		&mcpsdk.Tool{
			Name:        "remove_worktree",
			Description: "Removes an agent's worktree and branch. Checks for unmerged changes unless force=true.",
		},
		server.handleRemoveWorktree,
	)

	return server, nil
}

func (s *MCPServer) handleCreateWorktree(
	ctx context.Context,
	req *mcpsdk.CallToolRequest,
	args CreateWorktreeArgs,
) (*mcpsdk.CallToolResult, any, error) {
	request := application.CreateWorktreeRequest{
		AgentID: args.AgentID,
	}

	response, err := s.createWorktreeUseCase.Execute(ctx, request)
	if err != nil {
		message := fmt.Sprintf("Failed to create worktree: %v", err)
		return newErrorResult(message), nil, err
	}

	output := CreateWorktreeOutput{
		AgentID:      response.AgentID,
		WorktreePath: response.WorktreePath,
		BranchName:   response.BranchName,
		Status:       response.Status,
	}

	message := fmt.Sprintf("Successfully created worktree for agent '%s' at '%s' on branch '%s'", response.AgentID, response.WorktreePath, response.BranchName)
	return newSuccessResult(message), output, nil
}

func (s *MCPServer) handleRemoveWorktree(
	ctx context.Context,
	req *mcpsdk.CallToolRequest,
	args RemoveWorktreeArgs,
) (*mcpsdk.CallToolResult, any, error) {
	request := application.RemoveWorktreeRequest{
		AgentID: args.AgentID,
		Force:   args.Force,
	}

	response, err := s.removeWorktreeUseCase.Execute(ctx, request)
	if err != nil {
		message := fmt.Sprintf("Failed to remove worktree: %v", err)
		return newErrorResult(message), nil, err
	}

	output := RemoveWorktreeOutput{
		AgentID:            response.AgentID,
		HasUnmergedChanges: response.HasUnmergedChanges,
		UnmergedCommits:    response.UnmergedCommits,
		UncommittedFiles:   response.UncommittedFiles,
		Warning:            response.Warning,
	}

	if !response.RemovedAt.IsZero() {
		output.RemovedAt = response.RemovedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	if response.HasUnmergedChanges {
		message := fmt.Sprintf(
			"WARNING: Agent '%s' has unmerged changes\n\nUncommitted files: %d\nUnpushed commits: %d\n\n%s",
			response.AgentID,
			response.UncommittedFiles,
			response.UnmergedCommits,
			response.Warning,
		)
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{newTextContent(message)},
			IsError: false,
		}, output, nil
	}

	message := fmt.Sprintf("Successfully removed worktree for agent '%s'", response.AgentID)
	return newSuccessResult(message), output, nil
}

func (s *MCPServer) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcpsdk.StdioTransport{})
}
