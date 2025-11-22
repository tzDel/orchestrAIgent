package application

import (
	"context"
	"errors"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

type mockGitOperations struct {
	createWorktreeFunc        func(ctx context.Context, path string, branch string) error
	removeWorktreeFunc        func(ctx context.Context, path string, force bool) error
	branchExistsFunc          func(ctx context.Context, branch string) (bool, error)
	hasUncommittedChangesFunc func(ctx context.Context, worktreePath string) (bool, int, error)
	hasUnpushedCommitsFunc    func(ctx context.Context, baseBranch string, agentBranch string) (int, error)
	deleteBranchFunc          func(ctx context.Context, branchName string, force bool) error
}

func (mock *mockGitOperations) CreateWorktree(ctx context.Context, path string, branch string) error {
	if mock.createWorktreeFunc != nil {
		return mock.createWorktreeFunc(ctx, path, branch)
	}
	return nil
}

func (mock *mockGitOperations) RemoveWorktree(ctx context.Context, path string, force bool) error {
	if mock.removeWorktreeFunc != nil {
		return mock.removeWorktreeFunc(ctx, path, force)
	}
	return nil
}

func (mock *mockGitOperations) BranchExists(ctx context.Context, branch string) (bool, error) {
	if mock.branchExistsFunc != nil {
		return mock.branchExistsFunc(ctx, branch)
	}
	return false, nil
}

func (mock *mockGitOperations) HasUncommittedChanges(ctx context.Context, worktreePath string) (bool, int, error) {
	if mock.hasUncommittedChangesFunc != nil {
		return mock.hasUncommittedChangesFunc(ctx, worktreePath)
	}
	return false, 0, nil
}

func (mock *mockGitOperations) HasUnpushedCommits(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
	if mock.hasUnpushedCommitsFunc != nil {
		return mock.hasUnpushedCommitsFunc(ctx, baseBranch, agentBranch)
	}
	return 0, nil
}

func (mock *mockGitOperations) DeleteBranch(ctx context.Context, branchName string, force bool) error {
	if mock.deleteBranchFunc != nil {
		return mock.deleteBranchFunc(ctx, branchName, force)
	}
	return nil
}

type mockAgentRepository struct {
	agents map[string]*domain.Agent
}

func newMockAgentRepository() *mockAgentRepository {
	return &mockAgentRepository{
		agents: make(map[string]*domain.Agent),
	}
}

func (mock *mockAgentRepository) Save(ctx context.Context, agent *domain.Agent) error {
	mock.agents[agent.ID().String()] = agent
	return nil
}

func (mock *mockAgentRepository) FindByID(ctx context.Context, agentID domain.AgentID) (*domain.Agent, error) {
	agent, exists := mock.agents[agentID.String()]
	if !exists {
		return nil, errors.New("not found")
	}
	return agent, nil
}

func (mock *mockAgentRepository) Exists(ctx context.Context, agentID domain.AgentID) (bool, error) {
	_, exists := mock.agents[agentID.String()]
	return exists, nil
}
