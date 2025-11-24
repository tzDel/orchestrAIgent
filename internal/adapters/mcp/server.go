package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tzDel/orchestrAIgent/internal/application"
)

type CreateWorktreeArgs struct {
	SessionID string `json:"sessionId" jsonschema:"required" jsonschema_description:"The unique identifier for the session"`
}

type CreateWorktreeOutput struct {
	SessionID    string `json:"sessionId"`
	WorktreePath string `json:"worktreePath"`
	BranchName   string `json:"branchName"`
	Status       string `json:"status"`
}

type RemoveSessionArgs struct {
	SessionID string `json:"sessionId" jsonschema:"required" jsonschema_description:"Session identifier"`
	Force     bool   `json:"force" jsonschema_description:"Skip safety checks and force removal"`
}

type RemoveSessionOutput struct {
	SessionID          string `json:"sessionId"`
	RemovedAt          string `json:"removedAt,omitempty"`
	HasUnmergedChanges bool   `json:"hasUnmergedChanges"`
	UnmergedCommits    int    `json:"unmergedCommits"`
	UncommittedFiles   int    `json:"uncommittedFiles"`
	Warning            string `json:"warning,omitempty"`
}

type MCPServer struct {
	mcpServer             *mcpsdk.Server
	createWorktreeUseCase *application.CreateWorktreeUseCase
	removeSessionUseCase  *application.RemoveSessionUseCase
}

func NewMCPServer(
	createWorktreeUseCase *application.CreateWorktreeUseCase,
	removeSessionUseCase *application.RemoveSessionUseCase,
) (*MCPServer, error) {
	impl := &mcpsdk.Implementation{
		Name:    "orchestrAIgent",
		Version: "0.1.0",
	}

	mcpServer := mcpsdk.NewServer(impl, nil)

	server := &MCPServer{
		mcpServer:             mcpServer,
		createWorktreeUseCase: createWorktreeUseCase,
		removeSessionUseCase:  removeSessionUseCase,
	}

	mcpsdk.AddTool(
		mcpServer,
		&mcpsdk.Tool{
			Name:        "create_worktree",
			Description: "Creates an isolated git worktree for a specific session with its own branch",
		},
		server.handleCreateWorktree,
	)

	mcpsdk.AddTool(
		mcpServer,
		&mcpsdk.Tool{
			Name:        "remove_session",
			Description: "Removes an session's worktree and branch. Checks for unmerged changes unless force=true.",
		},
		server.handleRemoveSession,
	)

	return server, nil
}

func (s *MCPServer) handleCreateWorktree(
	ctx context.Context,
	req *mcpsdk.CallToolRequest,
	args CreateWorktreeArgs,
) (*mcpsdk.CallToolResult, any, error) {
	request := application.CreateWorktreeRequest{
		SessionID: args.SessionID,
	}

	response, err := s.createWorktreeUseCase.Execute(ctx, request)
	if err != nil {
		message := fmt.Sprintf("Failed to create worktree: %v", err)
		return newErrorResult(message), nil, err
	}

	output := CreateWorktreeOutput{
		SessionID:    response.SessionID,
		WorktreePath: response.WorktreePath,
		BranchName:   response.BranchName,
		Status:       response.Status,
	}

	message := fmt.Sprintf("Successfully created worktree for session '%s' at '%s' on branch '%s'", response.SessionID, response.WorktreePath, response.BranchName)
	return newSuccessResult(message), output, nil
}

func (s *MCPServer) handleRemoveSession(
	ctx context.Context,
	req *mcpsdk.CallToolRequest,
	args RemoveSessionArgs,
) (*mcpsdk.CallToolResult, any, error) {
	request := application.RemoveSessionRequest{
		SessionID: args.SessionID,
		Force:     args.Force,
	}

	response, err := s.removeSessionUseCase.Execute(ctx, request)
	if err != nil {
		message := fmt.Sprintf("Failed to remove worktree: %v", err)
		return newErrorResult(message), nil, err
	}

	output := RemoveSessionOutput{
		SessionID:          response.SessionID,
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
			"WARNING: Session '%s' has unmerged changes\n\nUncommitted files: %d\nUnpushed commits: %d\n\n%s",
			response.SessionID,
			response.UncommittedFiles,
			response.UnmergedCommits,
			response.Warning,
		)
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{newTextContent(message)},
			IsError: false,
		}, output, nil
	}

	message := fmt.Sprintf("Successfully removed worktree for session '%s'", response.SessionID)
	return newSuccessResult(message), output, nil
}

func (s *MCPServer) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcpsdk.StdioTransport{})
}
