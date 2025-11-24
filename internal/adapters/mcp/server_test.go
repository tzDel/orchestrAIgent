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
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	// act
	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)

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
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := CreateWorktreeArgs{
		SessionID: "copilot",
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
	if response.SessionID != "copilot" {
		t.Errorf("expected session ID 'copilot', got: %s", response.SessionID)
	}
	if response.BranchName != "session-copilot" {
		t.Errorf("expected branch name 'session-copilot', got: %s", response.BranchName)
	}
	if response.Status != "created" {
		t.Errorf("expected status 'created', got: %s", response.Status)
	}
}

func TestCreateWorktreeToolHandler_InvalidSessionID_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := CreateWorktreeArgs{
		SessionID: "invalid session id",
	}

	// act
	result, _, err := server.handleCreateWorktree(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for invalid session ID")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestCreateWorktreeToolHandler_DuplicateSession_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := CreateWorktreeArgs{
		SessionID: "copilot",
	}

	sessionID, _ := domain.NewSessionID("copilot")
	session, _ := domain.NewSession(sessionID, filepath.Join(repositoryRoot, ".worktrees", "copilot"))
	_ = sessionRepository.Save(ctx, session)

	// act
	result, _, err := server.handleCreateWorktree(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for duplicate session")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestRemoveSessionToolHandler_CleanWorktree_ReturnsSuccess(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()

	createArgs := CreateWorktreeArgs{SessionID: "test-session"}
	_, _, _ = server.handleCreateWorktree(ctx, nil, createArgs)

	removeArgs := RemoveSessionArgs{SessionID: "test-session", Force: false}

	// act
	result, output, err := server.handleRemoveSession(ctx, nil, removeArgs)

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

	response, ok := output.(RemoveSessionOutput)
	if !ok {
		t.Fatalf("expected output to be RemoveSessionOutput, got: %T", output)
	}
	if response.SessionID != "test-session" {
		t.Errorf("expected session ID 'test-session', got: %s", response.SessionID)
	}
	if response.HasUnmergedChanges {
		t.Error("expected HasUnmergedChanges to be false")
	}
	if response.RemovedAt == "" {
		t.Error("expected RemovedAt to be set")
	}
}

func TestRemoveSessionToolHandler_WithUncommittedChanges_ReturnsWarning(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()

	createArgs := CreateWorktreeArgs{SessionID: "test-session"}
	createResult, _, _ := server.handleCreateWorktree(ctx, nil, createArgs)
	if createResult.IsError {
		t.Fatalf("failed to create worktree: %v", createResult.Content)
	}

	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "session-test-session")
	newFilePath := filepath.Join(worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	removeArgs := RemoveSessionArgs{SessionID: "test-session", Force: false}

	// act
	result, output, err := server.handleRemoveSession(ctx, nil, removeArgs)

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

	response, ok := output.(RemoveSessionOutput)
	if !ok {
		t.Fatalf("expected output to be RemoveSessionOutput, got: %T", output)
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

func TestRemoveSessionToolHandler_ForceRemoveWithChanges_ReturnsSuccess(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()

	createArgs := CreateWorktreeArgs{SessionID: "test-session"}
	createResult, _, _ := server.handleCreateWorktree(ctx, nil, createArgs)
	if createResult.IsError {
		t.Fatalf("failed to create worktree: %v", createResult.Content)
	}

	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "session-test-session")
	newFilePath := filepath.Join(worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	removeArgs := RemoveSessionArgs{SessionID: "test-session", Force: true}

	// act
	result, output, err := server.handleRemoveSession(ctx, nil, removeArgs)

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

	response, ok := output.(RemoveSessionOutput)
	if !ok {
		t.Fatalf("expected output to be RemoveSessionOutput, got: %T", output)
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

func TestRemoveSessionToolHandler_InvalidSessionID_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := RemoveSessionArgs{SessionID: "invalid session id", Force: false}

	// act
	result, _, err := server.handleRemoveSession(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for invalid session ID")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestRemoveSessionToolHandler_NonexistentSession_ReturnsError(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := git.NewGitClient(repositoryRoot)
	sessionRepository := persistence.NewInMemorySessionRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, sessionRepository, repositoryRoot)
	removeSessionUseCase := application.NewRemoveSessionUseCase(gitClient, sessionRepository, "master")

	server, err := NewMCPServer(createWorktreeUseCase, removeSessionUseCase)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	ctx := context.Background()
	args := RemoveSessionArgs{SessionID: "nonexistent", Force: false}

	// act
	result, _, err := server.handleRemoveSession(ctx, nil, args)

	// assert
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
	if result != nil && !result.IsError {
		t.Error("expected IsError to be true")
	}
}
