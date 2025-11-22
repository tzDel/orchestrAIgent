package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GitClient struct {
	repositoryRoot string
}

func NewGitClient(repositoryRoot string) *GitClient {
	return &GitClient{repositoryRoot: repositoryRoot}
}

// executeGitCommand executes a git command in the repository root directory
// and returns the combined output (stdout and stderr) along with any error
func (gitClient *GitClient) executeGitCommand(ctx context.Context, args ...string) ([]byte, error) {
	gitCommand := exec.CommandContext(ctx, "git", args...)
	gitCommand.Dir = gitClient.repositoryRoot

	commandOutput, err := gitCommand.CombinedOutput()
	if err != nil {
		return commandOutput, fmt.Errorf("git command failed: %w (output: %s)", err, string(commandOutput))
	}

	return commandOutput, nil
}

// executeGitCommandWithOutput executes a git command and returns only stdout
// Used for commands where we need to parse the output (like branch --list)
func (gitClient *GitClient) executeGitCommandWithOutput(ctx context.Context, args ...string) ([]byte, error) {
	gitCommand := exec.CommandContext(ctx, "git", args...)
	gitCommand.Dir = gitClient.repositoryRoot

	commandOutput, err := gitCommand.Output()
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w", err)
	}

	return commandOutput, nil
}

func (gitClient *GitClient) CreateWorktree(ctx context.Context, worktreePath string, branchName string) error {
	_, err := gitClient.executeGitCommand(ctx, "worktree", "add", "-b", branchName, worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	return nil
}

func (gitClient *GitClient) RemoveWorktree(ctx context.Context, worktreePath string, force bool) error {
	args := []string{"worktree", "remove", worktreePath}
	if force {
		args = append(args, "--force")
	}

	_, err := gitClient.executeGitCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	return nil
}

func (gitClient *GitClient) BranchExists(ctx context.Context, branchName string) (bool, error) {
	commandOutput, err := gitClient.executeGitCommandWithOutput(ctx, "branch", "--list", branchName)
	if err != nil {
		return false, fmt.Errorf("failed to check branch existence: %w", err)
	}

	return strings.TrimSpace(string(commandOutput)) != "", nil
}

func (gitClient *GitClient) HasUncommittedChanges(ctx context.Context, worktreePath string) (bool, int, error) {
	commandOutput, err := gitClient.executeGitCommandWithOutput(ctx, "-C", worktreePath, "status", "--porcelain")
	if err != nil {
		return false, 0, fmt.Errorf("failed to check status: %w", err)
	}

	outputString := strings.TrimSpace(string(commandOutput))
	if outputString == "" {
		return false, 0, nil
	}

	lines := strings.Split(outputString, "\n")
	return true, len(lines), nil
}

func (gitClient *GitClient) HasUnpushedCommits(ctx context.Context, baseBranch string, agentBranch string) (int, error) {
	revRange := fmt.Sprintf("%s..%s", baseBranch, agentBranch)
	commandOutput, err := gitClient.executeGitCommandWithOutput(ctx, "rev-list", revRange, "--count")
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	outputString := strings.TrimSpace(string(commandOutput))
	count := 0
	_, parseErr := fmt.Sscanf(outputString, "%d", &count)
	if parseErr != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", parseErr)
	}

	return count, nil
}

func (gitClient *GitClient) DeleteBranch(ctx context.Context, branchName string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}

	_, err := gitClient.executeGitCommand(ctx, "branch", flag, branchName)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}
