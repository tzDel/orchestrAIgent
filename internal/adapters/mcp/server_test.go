package mcp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tzDel/orchestrAIgent/internal/application"
	"github.com/tzDel/orchestrAIgent/internal/domain"
	"github.com/tzDel/orchestrAIgent/internal/infrastructure/git"
	"github.com/tzDel/orchestrAIgent/internal/infrastructure/persistence"
)

func initializeGitRepo(repositoryPath string) error {
	gitInitCommand := exec.Command("git", "init")
	gitInitCommand.Dir = repositoryPath
	return gitInitCommand.Run()
}

func configureGitUser(repositoryPath string) error {
	gitConfigNameCommand := exec.Command("git", "config", "user.name", "Test User")
	gitConfigNameCommand.Dir = repositoryPath
	if err := gitConfigNameCommand.Run(); err != nil {
		return err
	}

	gitConfigEmailCommand := exec.Command("git", "config", "user.email", "test@example.com")
	gitConfigEmailCommand.Dir = repositoryPath
	return gitConfigEmailCommand.Run()
}

func createAndCommitFile(repositoryPath, filename, content string) error {
	filePath := filepath.Join(repositoryPath, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	gitAddCommand := exec.Command("git", "add", filename)
	gitAddCommand.Dir = repositoryPath
	if err := gitAddCommand.Run(); err != nil {
		return err
	}

	gitCommitCommand := exec.Command("git", "commit", "-m", "Initial commit")
	gitCommitCommand.Dir = repositoryPath
	return gitCommitCommand.Run()
}

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	temporaryDirectory := t.TempDir()

	if err := initializeGitRepo(temporaryDirectory); err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	if err := configureGitUser(temporaryDirectory); err != nil {
		t.Fatalf("failed to configure git user: %v", err)
	}

	if err := createAndCommitFile(temporaryDirectory, "README.md", "# Test Repo"); err != nil {
		t.Fatalf("failed to create and commit file: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(temporaryDirectory); err != nil {
			t.Logf("failed to remove temporary directory: %v", err)
		}
	}

	return temporaryDirectory, cleanup
}

func TestNewMCPServer_CreatesServerWithToolsRegistered(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	// act
	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)

	// assert
	if err != nil {
		t.Fatalf("expected no error creating MCP server, got: %v", err)
	}
	if server == nil {
		t.Fatal("expected server to be non-nil")
	}
	if server.mcpServer == nil {
		t.Fatal("expected internal MCP server to be initialized")
	}
}

func TestCreateWorktreeToolHandler_ValidInput_ReturnsSuccess(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := CreateWorktreeArgs{
		AgentID: "copilot",
	}

	// act
	result, output, err := server.handleCreateWorktree(ctx, nil, args)

	// assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.IsError {
		t.Error("expected IsError to be false")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content to be non-empty")
	}

	response, ok := output.(CreateWorktreeOutput)
	if !ok {
		t.Fatalf("expected output to be CreateWorktreeOutput, got: %T", output)
	}
	if response.AgentID != "copilot" {
		t.Errorf("expected agent ID 'copilot', got: %s", response.AgentID)
	}
	if response.BranchName != "agent-copilot" {
		t.Errorf("expected branch name 'agent-copilot', got: %s", response.BranchName)
	}
	if response.Status != "created" {
		t.Errorf("expected status 'created', got: %s", response.Status)
	}
}

func TestCreateWorktreeToolHandler_InvalidAgentID_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := CreateWorktreeArgs{
		AgentID: "invalid agent id",
	}

	// act
	result, _, err := server.handleCreateWorktree(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for invalid agent ID")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestCreateWorktreeToolHandler_DuplicateAgent_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := CreateWorktreeArgs{
		AgentID: "copilot",
	}

	agentID, _ := domain.NewAgentID("copilot")
	agent, _ := domain.NewAgent(agentID, filepath.Join(repositoryRoot, ".worktrees", "copilot"))
	_ = agentRepository.Save(ctx, agent)

	// act
	result, _, err := server.handleCreateWorktree(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for duplicate agent")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestRemoveWorktreeToolHandler_CleanWorktree_ReturnsSuccess(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()

	createArgs := CreateWorktreeArgs{AgentID: "test-agent"}
	_, _, _ = server.handleCreateWorktree(ctx, nil, createArgs)

	removeArgs := RemoveWorktreeArgs{AgentID: "test-agent", Force: false}

	// act
	result, output, err := server.handleRemoveWorktree(ctx, nil, removeArgs)

	// assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.IsError {
		t.Error("expected IsError to be false")
	}

	response, ok := output.(RemoveWorktreeOutput)
	if !ok {
		t.Fatalf("expected output to be RemoveWorktreeOutput, got: %T", output)
	}
	if response.AgentID != "test-agent" {
		t.Errorf("expected agent ID 'test-agent', got: %s", response.AgentID)
	}
	if response.HasUnmergedChanges {
		t.Error("expected HasUnmergedChanges to be false")
	}
	if response.RemovedAt == "" {
		t.Error("expected RemovedAt to be set")
	}
}

func TestRemoveWorktreeToolHandler_WithUncommittedChanges_ReturnsWarning(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()

	createArgs := CreateWorktreeArgs{AgentID: "test-agent"}
	createResult, _, _ := server.handleCreateWorktree(ctx, nil, createArgs)
	if createResult.IsError {
		t.Fatalf("failed to create worktree: %v", createResult.Content)
	}

	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "agent-test-agent")
	newFilePath := filepath.Join(worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	removeArgs := RemoveWorktreeArgs{AgentID: "test-agent", Force: false}

	// act
	result, output, err := server.handleRemoveWorktree(ctx, nil, removeArgs)

	// assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.IsError {
		t.Error("expected IsError to be false even with warning")
	}

	response, ok := output.(RemoveWorktreeOutput)
	if !ok {
		t.Fatalf("expected output to be RemoveWorktreeOutput, got: %T", output)
	}
	if !response.HasUnmergedChanges {
		t.Error("expected HasUnmergedChanges to be true")
	}
	if response.UncommittedFiles != 1 {
		t.Errorf("expected UncommittedFiles = 1, got %d", response.UncommittedFiles)
	}
	if response.Warning == "" {
		t.Error("expected warning message")
	}
	if response.RemovedAt != "" {
		t.Error("expected RemovedAt to be empty (not removed)")
	}
}

func TestRemoveWorktreeToolHandler_ForceRemoveWithChanges_ReturnsSuccess(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()

	createArgs := CreateWorktreeArgs{AgentID: "test-agent"}
	createResult, _, _ := server.handleCreateWorktree(ctx, nil, createArgs)
	if createResult.IsError {
		t.Fatalf("failed to create worktree: %v", createResult.Content)
	}

	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "agent-test-agent")
	newFilePath := filepath.Join(worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	removeArgs := RemoveWorktreeArgs{AgentID: "test-agent", Force: true}

	// act
	result, output, err := server.handleRemoveWorktree(ctx, nil, removeArgs)

	// assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.IsError {
		t.Error("expected IsError to be false")
	}

	response, ok := output.(RemoveWorktreeOutput)
	if !ok {
		t.Fatalf("expected output to be RemoveWorktreeOutput, got: %T", output)
	}
	if response.HasUnmergedChanges {
		t.Error("expected HasUnmergedChanges to be false when force=true")
	}
	if response.RemovedAt == "" {
		t.Error("expected RemovedAt to be set")
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("expected worktree directory to be removed")
	}
}

func TestRemoveWorktreeToolHandler_InvalidAgentID_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := RemoveWorktreeArgs{AgentID: "invalid agent id", Force: false}

	// act
	result, _, err := server.handleRemoveWorktree(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for invalid agent ID")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestRemoveWorktreeToolHandler_NonexistentAgent_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitClient, agentRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := RemoveWorktreeArgs{AgentID: "nonexistent", Force: false}

	// act
	result, _, err := server.handleRemoveWorktree(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for non-existent agent")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}
