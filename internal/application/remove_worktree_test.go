package application

import (
	"context"
	"errors"
	"testing"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

func TestRemoveWorktreeUseCase_Execute_AgentNotFound(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")
	request := RemoveWorktreeRequest{AgentID: "nonexistent", Force: false}
	ctx := context.Background()

	// act
	_, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for non-existent agent")
	}
}

func TestRemoveWorktreeUseCase_Execute_AgentAlreadyRemoved(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agent.MarkRemoved()
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: false}
	ctx := context.Background()

	// act
	_, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for already removed agent")
	}
}

func TestRemoveWorktreeUseCase_Execute_UncommittedChangesWithoutForce(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return true, 3, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
			return 0, nil
		},
	}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: false}
	ctx := context.Background()

	// act
	response, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if !response.HasUnmergedChanges {
		t.Error("Execute() expected HasUnmergedChanges to be true")
	}
	if response.UncommittedFiles != 3 {
		t.Errorf("Execute() UncommittedFiles = %d, want 3", response.UncommittedFiles)
	}
	if response.Warning == "" {
		t.Error("Execute() expected warning message")
	}
	if !response.RemovedAt.IsZero() {
		t.Error("Execute() expected RemovedAt to be zero (not removed)")
	}
}

func TestRemoveWorktreeUseCase_Execute_UnpushedCommitsWithoutForce(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
			return 5, nil
		},
	}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: false}
	ctx := context.Background()

	// act
	response, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if !response.HasUnmergedChanges {
		t.Error("Execute() expected HasUnmergedChanges to be true")
	}
	if response.UnmergedCommits != 5 {
		t.Errorf("Execute() UnmergedCommits = %d, want 5", response.UnmergedCommits)
	}
	if response.Warning == "" {
		t.Error("Execute() expected warning message")
	}
}

func TestRemoveWorktreeUseCase_Execute_CleanWorktree(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
			return 0, nil
		},
	}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: false}
	ctx := context.Background()

	// act
	response, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if response.HasUnmergedChanges {
		t.Error("Execute() expected HasUnmergedChanges to be false")
	}
	if response.RemovedAt.IsZero() {
		t.Error("Execute() expected RemovedAt to be set")
	}

	savedAgent, _ := agentRepository.FindByID(ctx, agentID)
	if savedAgent.Status() != domain.StatusRemoved {
		t.Errorf("Execute() agent status = %v, want %v", savedAgent.Status(), domain.StatusRemoved)
	}
}

func TestRemoveWorktreeUseCase_Execute_ForceRemoveWithChanges(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return true, 3, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
			return 2, nil
		},
	}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: true}
	ctx := context.Background()

	// act
	response, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if response.HasUnmergedChanges {
		t.Error("Execute() expected HasUnmergedChanges to be false when force=true")
	}
	if response.RemovedAt.IsZero() {
		t.Error("Execute() expected RemovedAt to be set")
	}

	savedAgent, _ := agentRepository.FindByID(ctx, agentID)
	if savedAgent.Status() != domain.StatusRemoved {
		t.Errorf("Execute() agent status = %v, want %v", savedAgent.Status(), domain.StatusRemoved)
	}
}

func TestRemoveWorktreeUseCase_Execute_InvalidAgentID(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")
	request := RemoveWorktreeRequest{AgentID: "Invalid_ID", Force: false}
	ctx := context.Background()

	// act
	_, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for invalid agent ID")
	}
}

func TestRemoveWorktreeUseCase_Execute_GitOperationFails(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
			return 0, nil
		},
		removeWorktreeFunc: func(ctx context.Context, path string, force bool) error {
			return errors.New("git error")
		},
	}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: false}
	ctx := context.Background()

	// act
	_, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error when git operation fails")
	}
}

func TestRemoveWorktreeUseCase_Execute_BranchDeleteFailsContinues(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
			return 0, nil
		},
		deleteBranchFunc: func(ctx context.Context, branchName string, force bool) error {
			return errors.New("branch delete error")
		},
	}
	agentRepository := newMockAgentRepository()
	removeWorktreeUseCase := NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := RemoveWorktreeRequest{AgentID: "test-agent", Force: false}
	ctx := context.Background()

	// act
	response, err := removeWorktreeUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if response.RemovedAt.IsZero() {
		t.Error("Execute() expected RemovedAt to be set even if branch delete fails")
	}

	savedAgent, _ := agentRepository.FindByID(ctx, agentID)
	if savedAgent.Status() != domain.StatusRemoved {
		t.Errorf("Execute() agent status = %v, want %v", savedAgent.Status(), domain.StatusRemoved)
	}
}
