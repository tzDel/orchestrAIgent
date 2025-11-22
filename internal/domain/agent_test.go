package domain

import "testing"

func TestNewAgent(t *testing.T) {
	// arrange
	agentID, _ := NewAgentID("test-agent")
	worktreePath := "/path/to/worktree"

	// act
	agent, err := NewAgent(agentID, worktreePath)

	// assert
	if err != nil {
		t.Fatalf("NewAgent() unexpected error: %v", err)
	}

	if agent.ID().String() != "test-agent" {
		t.Errorf("ID() = %q, want %q", agent.ID().String(), "test-agent")
	}

	if agent.Status() != StatusCreated {
		t.Errorf("Status() = %q, want %q", agent.Status(), StatusCreated)
	}

	if agent.WorktreePath() != worktreePath {
		t.Errorf("WorktreePath() = %q, want %q", agent.WorktreePath(), worktreePath)
	}

	if agent.BranchName() != "agent-test-agent" {
		t.Errorf("BranchName() = %q, want %q", agent.BranchName(), "agent-test-agent")
	}
}

func TestNewAgent_InvalidWorktreePath(t *testing.T) {
	// arrange
	agentID, _ := NewAgentID("test-agent")

	// act
	_, err := NewAgent(agentID, "")

	// assert
	if err == nil {
		t.Error("NewAgent() with empty path expected error, got nil")
	}
}

func TestAgent_MarkMerged(t *testing.T) {
	// arrange
	agentID, _ := NewAgentID("test-agent")
	agent, _ := NewAgent(agentID, "/path")

	// act
	agent.MarkMerged()

	// assert
	if agent.Status() != StatusMerged {
		t.Errorf("Status after MarkMerged() = %q, want %q", agent.Status(), StatusMerged)
	}
}

func TestAgent_MarkFailed(t *testing.T) {
	// arrange
	agentID, _ := NewAgentID("test-agent")
	agent, _ := NewAgent(agentID, "/path")

	// act
	agent.MarkFailed()

	// assert
	if agent.Status() != StatusFailed {
		t.Errorf("Status after MarkFailed() = %q, want %q", agent.Status(), StatusFailed)
	}
}

func TestAgent_MarkRemoved(t *testing.T) {
	// arrange
	agentID, _ := NewAgentID("test-agent")
	agent, _ := NewAgent(agentID, "/path")

	// act
	agent.MarkRemoved()

	// assert
	if agent.Status() != StatusRemoved {
		t.Errorf("Status after MarkRemoved() = %q, want %q", agent.Status(), StatusRemoved)
	}
}
