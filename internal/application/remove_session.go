package application

import (
	"context"
	"fmt"
	"time"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

type RemoveSessionRequest struct {
	SessionID string
	Force     bool
}

type RemoveSessionResponse struct {
	SessionID          string    `json:"sessionId"`
	RemovedAt          time.Time `json:"removedAt,omitempty"`
	HasUnmergedChanges bool      `json:"hasUnmergedChanges"`
	UnmergedCommits    int       `json:"unmergedCommits"`
	UncommittedFiles   int       `json:"uncommittedFiles"`
	Warning            string    `json:"warning,omitempty"`
}

type RemoveSessionUseCase struct {
	gitOperations     domain.GitOperations
	sessionRepository domain.SessionRepository
	baseBranch        string
}

func NewRemoveSessionUseCase(
	gitOperations domain.GitOperations,
	sessionRepository domain.SessionRepository,
	baseBranch string,
) *RemoveSessionUseCase {
	return &RemoveSessionUseCase{
		gitOperations:     gitOperations,
		sessionRepository: sessionRepository,
		baseBranch:        baseBranch,
	}
}

func (removeSessionUseCase *RemoveSessionUseCase) Execute(
	ctx context.Context,
	request RemoveSessionRequest,
) (*RemoveSessionResponse, error) {
	sessionID, err := removeSessionUseCase.validateSessionID(request.SessionID)
	if err != nil {
		return nil, err
	}
	session, err := removeSessionUseCase.fetchSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := removeSessionUseCase.verifySessionNotRemoved(session); err != nil {
		return nil, err
	}

	response := &RemoveSessionResponse{
		SessionID: request.SessionID,
	}

	if !request.Force {
		err := removeSessionUseCase.checkForUnmergedWork(ctx, session, response)
		if err != nil {
			return nil, err
		}
		if response.HasUnmergedChanges {
			return response, nil
		}
	}

	if err := removeSessionUseCase.removeSession(ctx, session, request.Force); err != nil {
		return nil, err
	}
	removeSessionUseCase.deleteBranchIfPossible(ctx, session)
	if err := removeSessionUseCase.markSessionRemoved(ctx, session); err != nil {
		return nil, err
	}

	response.RemovedAt = time.Now()
	response.HasUnmergedChanges = false
	return response, nil
}

func (removeSessionUseCase *RemoveSessionUseCase) validateSessionID(sessionIDString string) (domain.SessionID, error) {
	sessionID, err := domain.NewSessionID(sessionIDString)
	if err != nil {
		return domain.SessionID{}, fmt.Errorf("invalid session ID: %w", err)
	}
	return sessionID, nil
}

func (removeSessionUseCase *RemoveSessionUseCase) fetchSession(ctx context.Context, sessionID domain.SessionID) (*domain.Session, error) {
	session, err := removeSessionUseCase.sessionRepository.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	return session, nil
}

func (removeSessionUseCase *RemoveSessionUseCase) verifySessionNotRemoved(session *domain.Session) error {
	if session.Status() == domain.StatusRemoved {
		return fmt.Errorf("session already removed")
	}
	return nil
}

func (removeSessionUseCase *RemoveSessionUseCase) checkForUnmergedWork(
	ctx context.Context,
	session *domain.Session,
	response *RemoveSessionResponse,
) error {
	hasUncommitted, fileCount, err := removeSessionUseCase.gitOperations.HasUncommittedChanges(ctx, session.WorktreePath())
	if err != nil {
		return fmt.Errorf("failed to check uncommitted changes: %w", err)
	}

	unpushedCount, err := removeSessionUseCase.gitOperations.HasUnpushedCommits(ctx, removeSessionUseCase.baseBranch, session.BranchName())
	if err != nil {
		return fmt.Errorf("failed to check unpushed commits: %w", err)
	}

	response.UncommittedFiles = fileCount
	response.UnmergedCommits = unpushedCount

	if hasUncommitted || unpushedCount > 0 {
		removeSessionUseCase.setUnmergedChanges(response, unpushedCount, fileCount)
	}

	return nil
}

func (removeSessionUseCase *RemoveSessionUseCase) setUnmergedChanges(response *RemoveSessionResponse, unpushedCount int, fileCount int) {
	response.HasUnmergedChanges = true
	response.Warning = fmt.Sprintf(
		"Session has %d unpushed commits and %d uncommitted files. Call with force=true to remove anyway.",
		unpushedCount,
		fileCount,
	)
}

func (removeSessionUseCase *RemoveSessionUseCase) removeSession(ctx context.Context, session *domain.Session, force bool) error {
	if err := removeSessionUseCase.gitOperations.RemoveWorktree(ctx, session.WorktreePath(), force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

func (removeSessionUseCase *RemoveSessionUseCase) deleteBranchIfPossible(ctx context.Context, session *domain.Session) {
	removeSessionUseCase.gitOperations.DeleteBranch(ctx, session.BranchName(), true)
}

func (removeSessionUseCase *RemoveSessionUseCase) markSessionRemoved(ctx context.Context, session *domain.Session) error {
	session.MarkRemoved()
	if err := removeSessionUseCase.sessionRepository.Save(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}
