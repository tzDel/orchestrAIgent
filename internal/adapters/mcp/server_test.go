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
	useCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)

	// act
	server, err := NewMCPServer(useCase)

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
	useCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)

	server, err := NewMCPServer(useCase)
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
	useCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)

	server, err := NewMCPServer(useCase)
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
	useCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repositoryRoot)

	server, err := NewMCPServer(useCase)
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
