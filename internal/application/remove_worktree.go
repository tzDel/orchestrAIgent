package application

import (
	"context"
	"fmt"
	"time"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

type RemoveWorktreeRequest struct {
	AgentID string
	Force   bool
}

type RemoveWorktreeResponse struct {
	AgentID            string    `json:"agentId"`
	RemovedAt          time.Time `json:"removedAt,omitempty"`
	HasUnmergedChanges bool      `json:"hasUnmergedChanges"`
	UnmergedCommits    int       `json:"unmergedCommits"`
	UncommittedFiles   int       `json:"uncommittedFiles"`
	Warning            string    `json:"warning,omitempty"`
}

type RemoveWorktreeUseCase struct {
	gitOperations   domain.GitOperations
	agentRepository domain.AgentRepository
	baseBranch      string
}

func NewRemoveWorktreeUseCase(
	gitOperations domain.GitOperations,
	agentRepository domain.AgentRepository,
	baseBranch string,
) *RemoveWorktreeUseCase {
	return &RemoveWorktreeUseCase{
		gitOperations:   gitOperations,
		agentRepository: agentRepository,
		baseBranch:      baseBranch,
	}
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) Execute(
	ctx context.Context,
	request RemoveWorktreeRequest,
) (*RemoveWorktreeResponse, error) {
	agentID, err := removeWorktreeUseCase.validateAgentID(request.AgentID)
	if err != nil {
		return nil, err
	}
	agent, err := removeWorktreeUseCase.fetchAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if err := removeWorktreeUseCase.verifyAgentNotRemoved(agent); err != nil {
		return nil, err
	}

	response := &RemoveWorktreeResponse{
		AgentID: request.AgentID,
	}

	if !request.Force {
		err := removeWorktreeUseCase.checkForUnmergedWork(ctx, agent, response)
		if err != nil {
			return nil, err
		}
		if response.HasUnmergedChanges {
			return response, nil
		}
	}

	if err := removeWorktreeUseCase.removeWorktree(ctx, agent, request.Force); err != nil {
		return nil, err
	}
	removeWorktreeUseCase.deleteBranchIfPossible(ctx, agent)
	if err := removeWorktreeUseCase.markAgentRemoved(ctx, agent); err != nil {
		return nil, err
	}

	response.RemovedAt = time.Now()
	response.HasUnmergedChanges = false
	return response, nil
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) validateAgentID(agentIDString string) (domain.AgentID, error) {
	agentID, err := domain.NewAgentID(agentIDString)
	if err != nil {
		return domain.AgentID{}, fmt.Errorf("invalid agent ID: %w", err)
	}
	return agentID, nil
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) fetchAgent(ctx context.Context, agentID domain.AgentID) (*domain.Agent, error) {
	agent, err := removeWorktreeUseCase.agentRepository.FindByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}
	return agent, nil
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) verifyAgentNotRemoved(agent *domain.Agent) error {
	if agent.Status() == domain.StatusRemoved {
		return fmt.Errorf("agent already removed")
	}
	return nil
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) checkForUnmergedWork(
	ctx context.Context,
	agent *domain.Agent,
	response *RemoveWorktreeResponse,
) error {
	hasUncommitted, fileCount, err := removeWorktreeUseCase.gitOperations.HasUncommittedChanges(ctx, agent.WorktreePath())
	if err != nil {
		return fmt.Errorf("failed to check uncommitted changes: %w", err)
	}

	unpushedCount, err := removeWorktreeUseCase.gitOperations.HasUnpushedCommits(ctx, removeWorktreeUseCase.baseBranch, agent.BranchName())
	if err != nil {
		return fmt.Errorf("failed to check unpushed commits: %w", err)
	}

	response.UncommittedFiles = fileCount
	response.UnmergedCommits = unpushedCount

	if hasUncommitted || unpushedCount > 0 {
		removeWorktreeUseCase.setUnmergedChanges(response, unpushedCount, fileCount)
	}

	return nil
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) setUnmergedChanges(response *RemoveWorktreeResponse, unpushedCount int, fileCount int) {
	response.HasUnmergedChanges = true
	response.Warning = fmt.Sprintf(
		"Agent has %d unpushed commits and %d uncommitted files. Call with force=true to remove anyway.",
		unpushedCount,
		fileCount,
	)
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) removeWorktree(ctx context.Context, agent *domain.Agent, force bool) error {
	if err := removeWorktreeUseCase.gitOperations.RemoveWorktree(ctx, agent.WorktreePath(), force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) deleteBranchIfPossible(ctx context.Context, agent *domain.Agent) {
	removeWorktreeUseCase.gitOperations.DeleteBranch(ctx, agent.BranchName(), true)
}

func (removeWorktreeUseCase *RemoveWorktreeUseCase) markAgentRemoved(ctx context.Context, agent *domain.Agent) error {
	agent.MarkRemoved()
	if err := removeWorktreeUseCase.agentRepository.Save(ctx, agent); err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}
	return nil
}
