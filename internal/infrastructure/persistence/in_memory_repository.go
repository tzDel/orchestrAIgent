package persistence

import (
	"context"
	"fmt"
	"sync"

	"github.com/tzDel/orchestrAIgent/internal/domain"
)

type InMemoryAgentRepository struct {
	mutex  sync.RWMutex
	agents map[string]*domain.Agent
}

func NewInMemoryAgentRepository() *InMemoryAgentRepository {
	return &InMemoryAgentRepository{
		agents: make(map[string]*domain.Agent),
	}
}

func (repository *InMemoryAgentRepository) Save(ctx context.Context, agent *domain.Agent) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	repository.agents[agent.ID().String()] = agent
	return nil
}

func (repository *InMemoryAgentRepository) FindByID(ctx context.Context, agentID domain.AgentID) (*domain.Agent, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	agent, exists := repository.agents[agentID.String()]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID.String())
	}

	return agent, nil
}

func (repository *InMemoryAgentRepository) Exists(ctx context.Context, agentID domain.AgentID) (bool, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	_, exists := repository.agents[agentID.String()]
	return exists, nil
}
