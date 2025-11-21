package persistence

import (
	"context"
	"testing"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

func TestInMemoryRepository_SaveAndFind(t *testing.T) {
	// arrange
	repository := NewInMemoryAgentRepository()
	ctx := context.Background()
	agentID, _ := domain.NewAgentID("test-agent")
	agent, _ := domain.NewAgent(agentID, "/path/to/worktree")

	// act
	err := repository.Save(ctx, agent)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	found, err := repository.FindByID(ctx, agentID)

	// assert
	if err != nil {
		t.Fatalf("FindByID() error: %v", err)
	}

	if found.ID().String() != agent.ID().String() {
		t.Errorf("FindByID() returned wrong agent")
	}
}

func TestInMemoryRepository_FindByID_NotFound(t *testing.T) {
	// arrange
	repository := NewInMemoryAgentRepository()
	ctx := context.Background()
	agentID, _ := domain.NewAgentID("nonexistent")

	// act
	_, err := repository.FindByID(ctx, agentID)

	// assert
	if err == nil {
		t.Error("FindByID() expected error for non-existent agent")
	}
}

func TestInMemoryRepository_Exists(t *testing.T) {
	// arrange
	repository := NewInMemoryAgentRepository()
	ctx := context.Background()
	agentID, _ := domain.NewAgentID("test-agent")

	// act
	exists, err := repository.Exists(ctx, agentID)

	// assert
	if err != nil {
		t.Fatalf("Exists() error: %v", err)
	}
	if exists {
		t.Error("Exists() returned true for non-existent agent")
	}

	// arrange
	agent, _ := domain.NewAgent(agentID, "/path")
	repository.Save(ctx, agent)

	// act
	exists, err = repository.Exists(ctx, agentID)

	// assert
	if err != nil {
		t.Fatalf("Exists() error: %v", err)
	}
	if !exists {
		t.Error("Exists() returned false for existing agent")
	}
}
