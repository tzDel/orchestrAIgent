package application

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

type mockGitOperations struct {
	createWorktreeFunc func(ctx context.Context, path string, branch string) error
	branchExistsFunc   func(ctx context.Context, branch string) (bool, error)
}

func (mock *mockGitOperations) CreateWorktree(ctx context.Context, path string, branch string) error {
	if mock.createWorktreeFunc != nil {
		return mock.createWorktreeFunc(ctx, path, branch)
	}
	return nil
}

func (mock *mockGitOperations) RemoveWorktree(ctx context.Context, path string) error {
	return nil
}

func (mock *mockGitOperations) BranchExists(ctx context.Context, branch string) (bool, error) {
	if mock.branchExistsFunc != nil {
		return mock.branchExistsFunc(ctx, branch)
	}
	return false, nil
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

func TestCreateWorktreeUseCase_Execute_Success(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	request := CreateWorktreeRequest{AgentID: "test-agent"}
	ctx := context.Background()

	// act
	response, err := createWorktreeUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if response.AgentID != "test-agent" {
		t.Errorf("AgentID = %q, want %q", response.AgentID, "test-agent")
	}

	if response.BranchName != "agent-test-agent" {
		t.Errorf("BranchName = %q, want %q", response.BranchName, "agent-test-agent")
	}

	if response.Status != "created" {
		t.Errorf("Status = %q, want %q", response.Status, "created")
	}
}

func TestCreateWorktreeUseCase_Execute_InvalidAgentID(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	request := CreateWorktreeRequest{AgentID: "Invalid_ID"}
	ctx := context.Background()

	// act
	_, err := createWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for invalid agent ID")
	}
}

func TestCreateWorktreeUseCase_Execute_AgentAlreadyExists(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")

	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)

	request := CreateWorktreeRequest{AgentID: "test-agent"}
	ctx := context.Background()

	// act
	_, err := createWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for existing agent")
	}
}

func TestCreateWorktreeUseCase_Execute_BranchAlreadyExists(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		branchExistsFunc: func(ctx context.Context, branch string) (bool, error) {
			return true, nil
		},
	}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	request := CreateWorktreeRequest{AgentID: "test-agent"}
	ctx := context.Background()

	// act
	_, err := createWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for existing branch")
	}
}

func TestCreateWorktreeUseCase_Execute_GitOperationFails(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		createWorktreeFunc: func(ctx context.Context, path string, branch string) error {
			return errors.New("git error")
		},
	}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	request := CreateWorktreeRequest{AgentID: "test-agent"}
	ctx := context.Background()

	// act
	_, err := createWorktreeUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error when git operation fails")
	}
}

func TestCreateWorktreeUseCase_ValidateAgentID_WithValidID_ReturnsAgentID(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	validIDString := "test-agent"

	// act
	agentID, err := createWorktreeUseCase.validateAgentID(validIDString)

	// assert
	if err != nil {
		t.Fatalf("validateAgentID() unexpected error: %v", err)
	}
	if agentID.String() != validIDString {
		t.Errorf("validateAgentID() returned %q, want %q", agentID.String(), validIDString)
	}
}

func TestCreateWorktreeUseCase_ValidateAgentID_WithInvalidID_ReturnsError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	invalidIDString := "Invalid_ID"

	// act
	_, err := createWorktreeUseCase.validateAgentID(invalidIDString)

	// assert
	if err == nil {
		t.Error("validateAgentID() expected error for invalid agent ID")
	}
}

func TestCreateWorktreeUseCase_EnsureAgentDoesNotExist_WhenAgentDoesNotExist_ReturnsNoError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	agentID, _ := domain.NewAgentID("test-agent")
	ctx := context.Background()

	// act
	err := createWorktreeUseCase.ensureAgentDoesNotExist(ctx, agentID)

	// assert
	if err != nil {
		t.Errorf("ensureAgentDoesNotExist() unexpected error: %v", err)
	}
}

func TestCreateWorktreeUseCase_EnsureAgentDoesNotExist_WhenAgentExists_ReturnsError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path")
	agentRepository.Save(context.Background(), agent)
	ctx := context.Background()

	// act
	err := createWorktreeUseCase.ensureAgentDoesNotExist(ctx, agentID)

	// assert
	if err == nil {
		t.Error("ensureAgentDoesNotExist() expected error when agent exists")
	}
}

func TestCreateWorktreeUseCase_EnsureBranchDoesNotExist_WhenBranchDoesNotExist_ReturnsNoError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		branchExistsFunc: func(ctx context.Context, branch string) (bool, error) {
			return false, nil
		},
	}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	branchName := "agent-test-agent"
	ctx := context.Background()

	// act
	err := createWorktreeUseCase.ensureBranchDoesNotExist(ctx, branchName)

	// assert
	if err != nil {
		t.Errorf("ensureBranchDoesNotExist() unexpected error: %v", err)
	}
}

func TestCreateWorktreeUseCase_EnsureBranchDoesNotExist_WhenBranchExists_ReturnsError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		branchExistsFunc: func(ctx context.Context, branch string) (bool, error) {
			return true, nil
		},
	}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	branchName := "agent-test-agent"
	ctx := context.Background()

	// act
	err := createWorktreeUseCase.ensureBranchDoesNotExist(ctx, branchName)

	// assert
	if err == nil {
		t.Error("ensureBranchDoesNotExist() expected error when branch exists")
	}
}

func TestCreateWorktreeUseCase_BuildWorktreePath_ReturnsCorrectPath(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	agentID, _ := domain.NewAgentID("test-agent")

	// act
	worktreePath := createWorktreeUseCase.buildWorktreePath(agentID)

	// assert
	expectedPath := filepath.Join("/repo/root", ".worktrees", "agent-test-agent")
	if worktreePath != expectedPath {
		t.Errorf("buildWorktreePath() returned %q, want %q", worktreePath, expectedPath)
	}
}

func TestCreateWorktreeUseCase_CreateWorktreeAndBranch_Success_ReturnsNoError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	worktreePath := "/repo/root/.worktrees/agent-test-agent"
	branchName := "agent-test-agent"
	ctx := context.Background()

	// act
	err := createWorktreeUseCase.createWorktreeAndBranch(ctx, worktreePath, branchName)

	// assert
	if err != nil {
		t.Errorf("createWorktreeAndBranch() unexpected error: %v", err)
	}
}

func TestCreateWorktreeUseCase_CreateWorktreeAndBranch_GitOperationFails_ReturnsError(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		createWorktreeFunc: func(ctx context.Context, path string, branch string) error {
			return errors.New("git error")
		},
	}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	worktreePath := "/repo/root/.worktrees/agent-test-agent"
	branchName := "agent-test-agent"
	ctx := context.Background()

	// act
	err := createWorktreeUseCase.createWorktreeAndBranch(ctx, worktreePath, branchName)

	// assert
	if err == nil {
		t.Error("createWorktreeAndBranch() expected error when git operation fails")
	}
}

func TestCreateWorktreeUseCase_CreateAndSaveAgent_Success_ReturnsAgent(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	agentID, _ := domain.NewAgentID("test-agent")
	worktreePath := "/repo/root/.worktrees/agent-test-agent"
	ctx := context.Background()

	// act
	agent, err := createWorktreeUseCase.createAndSaveAgent(ctx, agentID, worktreePath)

	// assert
	if err != nil {
		t.Fatalf("createAndSaveAgent() unexpected error: %v", err)
	}
	if agent.ID().String() != "test-agent" {
		t.Errorf("createAndSaveAgent() agent ID = %q, want %q", agent.ID().String(), "test-agent")
	}
	if agent.WorktreePath() != worktreePath {
		t.Errorf("createAndSaveAgent() worktree path = %q, want %q", agent.WorktreePath(), worktreePath)
	}
}

func TestCreateWorktreeUseCase_BuildResponse_ReturnsCorrectResponse(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	agentRepository := newMockAgentRepository()
	createWorktreeUseCase := NewCreateWorktreeUseCase(gitOperations, agentRepository, "/repo/root")
	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/repo/root/.worktrees/agent-test-agent")

	// act
	response := createWorktreeUseCase.buildResponse(agent)

	// assert
	if response.AgentID != "test-agent" {
		t.Errorf("buildResponse() AgentID = %q, want %q", response.AgentID, "test-agent")
	}
	if response.BranchName != "agent-test-agent" {
		t.Errorf("buildResponse() BranchName = %q, want %q", response.BranchName, "agent-test-agent")
	}
	if response.Status != "created" {
		t.Errorf("buildResponse() Status = %q, want %q", response.Status, "created")
	}
}
