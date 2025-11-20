package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tzDel/agent-manager-mcp/internal/application"
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

type MCPServer struct {
	mcpServer             *mcpsdk.Server
	createWorktreeUseCase *application.CreateWorktreeUseCase
}

func NewMCPServer(createWorktreeUseCase *application.CreateWorktreeUseCase) (*MCPServer, error) {
	impl := &mcpsdk.Implementation{
		Name:    "agent-manager-mcp",
		Version: "0.1.0",
	}

	mcpServer := mcpsdk.NewServer(impl, nil)

	server := &MCPServer{
		mcpServer:             mcpServer,
		createWorktreeUseCase: createWorktreeUseCase,
	}

	mcpsdk.AddTool(
		mcpServer,
		&mcpsdk.Tool{
			Name:        "create_worktree",
			Description: "Creates an isolated git worktree for a specific agent with its own branch",
		},
		server.handleCreateWorktree,
	)

	return server, nil
}

func (s *MCPServer) handleCreateWorktree(
	ctx context.Context,
	req *mcpsdk.CallToolRequest,
	args CreateWorktreeArgs,
) (*mcpsdk.CallToolResult, any, error) {
	// req is required by MCP SDK tool handler signature but unused here
	// All necessary input comes from args; req would be used for request metadata (tracing, audit logs, etc.)
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

func (s *MCPServer) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcpsdk.StdioTransport{})
}
