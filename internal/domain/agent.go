package domain

import (
	"errors"
	"time"
)

type AgentStatus string

const (
	StatusCreated AgentStatus = "created"
	StatusMerged  AgentStatus = "merged"
	StatusFailed  AgentStatus = "failed"
	StatusRemoved AgentStatus = "removed"
)

type Agent struct {
	id           AgentID
	status       AgentStatus
	worktreePath string
	branchName   string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewAgent(agentID AgentID, worktreePath string) (*Agent, error) {
	if worktreePath == "" {
		return nil, errors.New("worktree path cannot be empty")
	}

	now := time.Now()
	return &Agent{
		id:           agentID,
		status:       StatusCreated,
		worktreePath: worktreePath,
		branchName:   agentID.BranchName(),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func (agent *Agent) ID() AgentID {
	return agent.id
}

func (agent *Agent) Status() AgentStatus {
	return agent.status
}

func (agent *Agent) WorktreePath() string {
	return agent.worktreePath
}

func (agent *Agent) BranchName() string {
	return agent.branchName
}

func (agent *Agent) MarkMerged() {
	agent.status = StatusMerged
	agent.updatedAt = time.Now()
}

func (agent *Agent) MarkFailed() {
	agent.status = StatusFailed
	agent.updatedAt = time.Now()
}

func (agent *Agent) MarkRemoved() {
	agent.status = StatusRemoved
	agent.updatedAt = time.Now()
}
