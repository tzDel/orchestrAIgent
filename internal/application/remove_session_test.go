package application

import (
	"context"
	"errors"
	"testing"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

func TestRemoveSessionUseCase_Execute_SessionNotFound(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")
	request := RemoveSessionRequest{SessionID: "nonexistent", Force: false}
	ctx := context.Background()

	// act
	_, err := removeSessionUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for non-existent session")
	}
}

func TestRemoveSessionUseCase_Execute_SessionAlreadyRemoved(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	session.MarkRemoved()
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: false}
	ctx := context.Background()

	// act
	_, err := removeSessionUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for already removed session")
	}
}

func TestRemoveSessionUseCase_Execute_UncommittedChangesWithoutForce(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return true, 3, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, sessionBranch string) (int, error) {
			return 0, nil
		},
	}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: false}
	ctx := context.Background()

	// act
	response, err := removeSessionUseCase.Execute(ctx, request)

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

func TestRemoveSessionUseCase_Execute_UnpushedCommitsWithoutForce(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, sessionBranch string) (int, error) {
			return 5, nil
		},
	}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: false}
	ctx := context.Background()

	// act
	response, err := removeSessionUseCase.Execute(ctx, request)

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

func TestRemoveSessionUseCase_Execute_CleanWorktree(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, sessionBranch string) (int, error) {
			return 0, nil
		},
	}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: false}
	ctx := context.Background()

	// act
	response, err := removeSessionUseCase.Execute(ctx, request)

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

	savedSession, _ := sessionRepository.FindByID(ctx, sessionID)
	if savedSession.Status() != domain.StatusRemoved {
		t.Errorf("Execute() session status = %v, want %v", savedSession.Status(), domain.StatusRemoved)
	}
}

func TestRemoveSessionUseCase_Execute_ForceRemoveWithChanges(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return true, 3, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, sessionBranch string) (int, error) {
			return 2, nil
		},
	}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: true}
	ctx := context.Background()

	// act
	response, err := removeSessionUseCase.Execute(ctx, request)

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

	savedSession, _ := sessionRepository.FindByID(ctx, sessionID)
	if savedSession.Status() != domain.StatusRemoved {
		t.Errorf("Execute() session status = %v, want %v", savedSession.Status(), domain.StatusRemoved)
	}
}

func TestRemoveSessionUseCase_Execute_InvalidSessionID(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")
	request := RemoveSessionRequest{SessionID: "Invalid_ID", Force: false}
	ctx := context.Background()

	// act
	_, err := removeSessionUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error for invalid session ID")
	}
}

func TestRemoveSessionUseCase_Execute_GitOperationFails(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, sessionBranch string) (int, error) {
			return 0, nil
		},
		removeWorktreeFunc: func(ctx context.Context, path string, force bool) error {
			return errors.New("git error")
		},
	}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: false}
	ctx := context.Background()

	// act
	_, err := removeSessionUseCase.Execute(ctx, request)

	// assert
	if err == nil {
		t.Error("Execute() expected error when git operation fails")
	}
}

func TestRemoveSessionUseCase_Execute_BranchDeleteFailsContinues(t *testing.T) {
	// arrange
	gitOperations := &mockGitOperations{
		hasUncommittedChangesFunc: func(ctx context.Context, worktreePath string) (bool, int, error) {
			return false, 0, nil
		},
		hasUnpushedCommitsFunc: func(ctx context.Context, baseBranch string, sessionBranch string) (int, error) {
			return 0, nil
		},
		deleteBranchFunc: func(ctx context.Context, branchName string, force bool) error {
			return errors.New("branch delete error")
		},
	}
	sessionRepository := newMockSessionRepository()
	removeSessionUseCase := NewRemoveSessionUseCase(gitOperations, sessionRepository, "main")

	sessionID, _ := domain.NewSessionID("test-session")
	session, _ := domain.NewSession(sessionID, "/path")
	sessionRepository.Save(context.Background(), session)

	request := RemoveSessionRequest{SessionID: "test-session", Force: false}
	ctx := context.Background()

	// act
	response, err := removeSessionUseCase.Execute(ctx, request)

	// assert
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if response.RemovedAt.IsZero() {
		t.Error("Execute() expected RemovedAt to be set even if branch delete fails")
	}

	savedSession, _ := sessionRepository.FindByID(ctx, sessionID)
	if savedSession.Status() != domain.StatusRemoved {
		t.Errorf("Execute() session status = %v, want %v", savedSession.Status(), domain.StatusRemoved)
	}
}
