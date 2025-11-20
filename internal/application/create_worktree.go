package application

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/tzDel/agent-manager-mcp/internal/domain"
)

type CreateWorktreeRequest struct {
	AgentID string
}

type CreateWorktreeResponse struct {
	AgentID      string
	WorktreePath string
	BranchName   string
	Status       string
}

type CreateWorktreeUseCase struct {
	gitOperations     domain.GitOperations
	agentRepository   domain.AgentRepository
	repositoryRoot    string
	worktreeDirectory string
}

func NewCreateWorktreeUseCase(
	gitOperations domain.GitOperations,
	agentRepository domain.AgentRepository,
	repositoryRoot string,
) *CreateWorktreeUseCase {
	return &CreateWorktreeUseCase{
		gitOperations:     gitOperations,
		agentRepository:   agentRepository,
		repositoryRoot:    repositoryRoot,
		worktreeDirectory: filepath.Join(repositoryRoot, ".worktrees"),
	}
}

func (createWorktreeUseCase *CreateWorktreeUseCase) Execute(ctx context.Context, request CreateWorktreeRequest) (*CreateWorktreeResponse, error) {
	agentID, err := createWorktreeUseCase.validateAgentID(request.AgentID)
	if err != nil {
		return nil, err
	}

	if err := createWorktreeUseCase.ensureAgentDoesNotExist(ctx, agentID); err != nil {
		return nil, err
	}

	if err := createWorktreeUseCase.ensureBranchDoesNotExist(ctx, agentID.BranchName()); err != nil {
		return nil, err
	}

	worktreePath := createWorktreeUseCase.buildWorktreePath(agentID)

	if err := createWorktreeUseCase.createWorktreeAndBranch(ctx, worktreePath, agentID.BranchName()); err != nil {
		return nil, err
	}

	agent, err := createWorktreeUseCase.createAndSaveAgent(ctx, agentID, worktreePath)
	if err != nil {
		return nil, err
	}

	return createWorktreeUseCase.buildResponse(agent), nil
}

func (createWorktreeUseCase *CreateWorktreeUseCase) validateAgentID(agentIDString string) (domain.AgentID, error) {
	agentID, err := domain.NewAgentID(agentIDString)
	if err != nil {
		return domain.AgentID{}, fmt.Errorf("invalid agent ID: %w", err)
	}
	return agentID, nil
}

func (createWorktreeUseCase *CreateWorktreeUseCase) ensureAgentDoesNotExist(ctx context.Context, agentID domain.AgentID) error {
	exists, err := createWorktreeUseCase.agentRepository.Exists(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to check agent existence: %w", err)
	}
	if exists {
		return fmt.Errorf("agent already exists: %s", agentID.String())
	}
	return nil
}

func (createWorktreeUseCase *CreateWorktreeUseCase) ensureBranchDoesNotExist(ctx context.Context, branchName string) error {
	branchExists, err := createWorktreeUseCase.gitOperations.BranchExists(ctx, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if branchExists {
		return fmt.Errorf("branch already exists: %s", branchName)
	}
	return nil
}

func (createWorktreeUseCase *CreateWorktreeUseCase) buildWorktreePath(agentID domain.AgentID) string {
	return filepath.Join(createWorktreeUseCase.worktreeDirectory, agentID.WorktreeDirName())
}

func (createWorktreeUseCase *CreateWorktreeUseCase) createWorktreeAndBranch(ctx context.Context, worktreePath string, branchName string) error {
	if err := createWorktreeUseCase.gitOperations.CreateWorktree(ctx, worktreePath, branchName); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	return nil
}

func (createWorktreeUseCase *CreateWorktreeUseCase) createAndSaveAgent(ctx context.Context, agentID domain.AgentID, worktreePath string) (*domain.Agent, error) {
	agent, err := domain.NewAgent(agentID, worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	if err := createWorktreeUseCase.agentRepository.Save(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to save agent: %w", err)
	}

	return agent, nil
}

func (createWorktreeUseCase *CreateWorktreeUseCase) buildResponse(agent *domain.Agent) *CreateWorktreeResponse {
	return &CreateWorktreeResponse{
		AgentID:      agent.ID().String(),
		WorktreePath: agent.WorktreePath(),
		BranchName:   agent.BranchName(),
		Status:       string(agent.Status()),
	}
}
