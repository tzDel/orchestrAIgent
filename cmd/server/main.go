package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tzDel/agent-manager-mcp/internal/adapters/mcp"
	"github.com/tzDel/agent-manager-mcp/internal/application"
	"github.com/tzDel/agent-manager-mcp/internal/infrastructure/git"
	"github.com/tzDel/agent-manager-mcp/internal/infrastructure/persistence"
)

func main() {
	var repoRoot string
	flag.StringVar(&repoRoot, "repo", "", "path to git repository (defaults to current directory)")
	flag.Parse()

	if repoRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current directory: %v", err)
		}
		repoRoot = cwd
	}

	gitClient := git.NewClient(repoRoot)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitClient, agentRepository, repoRoot)

	server, err := mcp.NewMCPServer(createWorktreeUseCase)
	if err != nil {
		log.Fatalf("failed to create MCP server: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting MCP server for repository: %s\n", repoRoot)

	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
