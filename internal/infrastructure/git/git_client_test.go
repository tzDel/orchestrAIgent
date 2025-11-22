package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	temporaryDirectory := t.TempDir()

	gitInitCommand := exec.Command("git", "init")
	gitInitCommand.Dir = temporaryDirectory
	if err := gitInitCommand.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	gitConfigNameCommand := exec.Command("git", "config", "user.name", "Test User")
	gitConfigNameCommand.Dir = temporaryDirectory
	gitConfigNameCommand.Run()

	gitConfigEmailCommand := exec.Command("git", "config", "user.email", "test@example.com")
	gitConfigEmailCommand.Dir = temporaryDirectory
	gitConfigEmailCommand.Run()

	testFilePath := filepath.Join(temporaryDirectory, "README.md")
	if err := os.WriteFile(testFilePath, []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gitAddCommand := exec.Command("git", "add", "README.md")
	gitAddCommand.Dir = temporaryDirectory
	if err := gitAddCommand.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	gitCommitCommand := exec.Command("git", "commit", "-m", "Initial commit")
	gitCommitCommand.Dir = temporaryDirectory
	if err := gitCommitCommand.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(temporaryDirectory)
	}

	return temporaryDirectory, cleanup
}

type testRepoWithWorktree struct {
	repositoryRoot string
	gitClient      *GitClient
	ctx            context.Context
	worktreePath   string
	branchName     string
	cleanup        func()
}

func setupTestRepoWithWorktree(t *testing.T) *testRepoWithWorktree {
	t.Helper()

	repositoryRoot, cleanup := setupTestRepo(t)
	gitClient := NewGitClient(repositoryRoot)
	ctx := context.Background()
	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "test-agent")
	branchName := "agent-test"

	if err := gitClient.CreateWorktree(ctx, worktreePath, branchName); err != nil {
		cleanup()
		t.Fatalf("Failed to create worktree: %v", err)
	}

	return &testRepoWithWorktree{
		repositoryRoot: repositoryRoot,
		gitClient:      gitClient,
		ctx:            ctx,
		worktreePath:   worktreePath,
		branchName:     branchName,
		cleanup:        cleanup,
	}
}

func TestGitClient_CreateWorktree(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := NewGitClient(repositoryRoot)
	ctx := context.Background()
	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "test-agent")
	branchName := "agent-test"

	// act
	err := gitClient.CreateWorktree(ctx, worktreePath, branchName)

	// assert
	if err != nil {
		t.Fatalf("CreateWorktree() error: %v", err)
	}

	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("Worktree directory was not created")
	}

	exists, err := gitClient.BranchExists(ctx, branchName)
	if err != nil {
		t.Fatalf("BranchExists() error: %v", err)
	}
	if !exists {
		t.Error("Branch was not created")
	}
}

func TestGitClient_RemoveWorktree(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	// act
	err := setup.gitClient.RemoveWorktree(setup.ctx, setup.worktreePath, false)

	// assert
	if err != nil {
		t.Fatalf("RemoveWorktree() error: %v", err)
	}

	if _, err := os.Stat(setup.worktreePath); !os.IsNotExist(err) {
		t.Error("Worktree directory still exists after removal")
	}
}

func TestGitClient_RemoveWorktree_WithForce(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	newFilePath := filepath.Join(setup.worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	// act
	err := setup.gitClient.RemoveWorktree(setup.ctx, setup.worktreePath, true)

	// assert
	if err != nil {
		t.Fatalf("RemoveWorktree() with force error: %v", err)
	}

	if _, err := os.Stat(setup.worktreePath); !os.IsNotExist(err) {
		t.Error("Worktree directory still exists after force removal")
	}
}

func TestGitClient_BranchExists(t *testing.T) {
	// arrange
	repositoryRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	gitClient := NewGitClient(repositoryRoot)
	ctx := context.Background()

	// act
	exists, err := gitClient.BranchExists(ctx, "nonexistent")

	// assert
	if err != nil {
		t.Fatalf("BranchExists() error: %v", err)
	}
	if exists {
		t.Error("BranchExists() returned true for non-existent branch")
	}

	// arrange
	worktreePath := filepath.Join(repositoryRoot, ".worktrees", "test-agent")
	branchName := "agent-test"
	gitClient.CreateWorktree(ctx, worktreePath, branchName)

	// act
	exists, err = gitClient.BranchExists(ctx, branchName)

	// assert
	if err != nil {
		t.Fatalf("BranchExists() error: %v", err)
	}
	if !exists {
		t.Error("BranchExists() returned false for existing branch")
	}
}

func TestGitClient_HasUncommittedChanges_CleanWorktree(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	// act
	hasChanges, fileCount, err := setup.gitClient.HasUncommittedChanges(setup.ctx, setup.worktreePath)

	// assert
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error: %v", err)
	}
	if hasChanges {
		t.Error("HasUncommittedChanges() returned true for clean worktree")
	}
	if fileCount != 0 {
		t.Errorf("HasUncommittedChanges() fileCount = %d, want 0", fileCount)
	}
}

func TestGitClient_HasUncommittedChanges_ModifiedFiles(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	testFilePath := filepath.Join(setup.worktreePath, "README.md")
	os.WriteFile(testFilePath, []byte("# Modified"), 0644)

	// act
	hasChanges, fileCount, err := setup.gitClient.HasUncommittedChanges(setup.ctx, setup.worktreePath)

	// assert
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error: %v", err)
	}
	if !hasChanges {
		t.Error("HasUncommittedChanges() returned false for modified files")
	}
	if fileCount != 1 {
		t.Errorf("HasUncommittedChanges() fileCount = %d, want 1", fileCount)
	}
}

func TestGitClient_HasUncommittedChanges_UntrackedFiles(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	newFilePath := filepath.Join(setup.worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	// act
	hasChanges, fileCount, err := setup.gitClient.HasUncommittedChanges(setup.ctx, setup.worktreePath)

	// assert
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error: %v", err)
	}
	if !hasChanges {
		t.Error("HasUncommittedChanges() returned false for untracked files")
	}
	if fileCount != 1 {
		t.Errorf("HasUncommittedChanges() fileCount = %d, want 1", fileCount)
	}
}

func TestGitClient_HasUnpushedCommits_NoUnpushedCommits(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	// act
	count, err := setup.gitClient.HasUnpushedCommits(setup.ctx, "master", setup.branchName)

	// assert
	if err != nil {
		t.Fatalf("HasUnpushedCommits() error: %v", err)
	}
	if count != 0 {
		t.Errorf("HasUnpushedCommits() count = %d, want 0", count)
	}
}

func TestGitClient_HasUnpushedCommits_WithUnpushedCommits(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	newFilePath := filepath.Join(setup.worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	gitAddCommand := exec.Command("git", "add", "new-file.txt")
	gitAddCommand.Dir = setup.worktreePath
	gitAddCommand.Run()

	gitCommitCommand := exec.Command("git", "commit", "-m", "Add new file")
	gitCommitCommand.Dir = setup.worktreePath
	gitCommitCommand.Run()

	// act
	count, err := setup.gitClient.HasUnpushedCommits(setup.ctx, "master", setup.branchName)

	// assert
	if err != nil {
		t.Fatalf("HasUnpushedCommits() error: %v", err)
	}
	if count != 1 {
		t.Errorf("HasUnpushedCommits() count = %d, want 1", count)
	}
}

func TestGitClient_DeleteBranch_SafeDelete(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	setup.gitClient.RemoveWorktree(setup.ctx, setup.worktreePath, false)

	// act
	err := setup.gitClient.DeleteBranch(setup.ctx, setup.branchName, false)

	// assert
	if err != nil {
		t.Fatalf("DeleteBranch() error: %v", err)
	}

	exists, _ := setup.gitClient.BranchExists(setup.ctx, setup.branchName)
	if exists {
		t.Error("Branch still exists after deletion")
	}
}

func TestGitClient_DeleteBranch_ForceDelete(t *testing.T) {
	// arrange
	setup := setupTestRepoWithWorktree(t)
	defer setup.cleanup()

	newFilePath := filepath.Join(setup.worktreePath, "new-file.txt")
	os.WriteFile(newFilePath, []byte("new content"), 0644)

	gitAddCommand := exec.Command("git", "add", "new-file.txt")
	gitAddCommand.Dir = setup.worktreePath
	gitAddCommand.Run()

	gitCommitCommand := exec.Command("git", "commit", "-m", "Add new file")
	gitCommitCommand.Dir = setup.worktreePath
	gitCommitCommand.Run()

	setup.gitClient.RemoveWorktree(setup.ctx, setup.worktreePath, true)

	// act
	err := setup.gitClient.DeleteBranch(setup.ctx, setup.branchName, true)

	// assert
	if err != nil {
		t.Fatalf("DeleteBranch() error: %v", err)
	}

	exists, _ := setup.gitClient.BranchExists(setup.ctx, setup.branchName)
	if exists {
		t.Error("Branch still exists after force deletion")
	}
}
