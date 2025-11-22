package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tzDel/orchestrAIgent/internal/adapters/mcp"
	"github.com/tzDel/orchestrAIgent/internal/application"
	"github.com/tzDel/orchestrAIgent/internal/infrastructure/git"
	"github.com/tzDel/orchestrAIgent/internal/infrastructure/persistence"
)

func main() {
	repositoryPath := parseRepositoryPath()
	server := initializeMCPServer(repositoryPath)
	startMCPServer(server, repositoryPath)
}

func parseRepositoryPath() string {
	var repositoryPath string
	flag.StringVar(&repositoryPath, "repo", "", "path to git repository (defaults to current directory)")
	flag.Parse()

	if repositoryPath == "" {
		return resolveCurrentWorkingDirectory()
	}

	return repositoryPath
}

func resolveCurrentWorkingDirectory() string {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to resolve current working directory: %v", err)
	}
	return currentWorkingDirectory
}

func initializeMCPServer(repositoryPath string) *mcp.MCPServer {
	gitOperations := git.NewGitClient(repositoryPath)
	agentRepository := persistence.NewInMemoryAgentRepository()
	createWorktreeUseCase := application.NewCreateWorktreeUseCase(gitOperations, agentRepository, repositoryPath)
	removeWorktreeUseCase := application.NewRemoveWorktreeUseCase(gitOperations, agentRepository, "main")

	server, err := mcp.NewMCPServer(createWorktreeUseCase, removeWorktreeUseCase)
	if err != nil {
		log.Fatalf("failed to initialize MCP server: %v", err)
	}

	return server
}

func startMCPServer(server *mcp.MCPServer, repositoryPath string) {
	fmt.Fprintf(os.Stderr, "Starting MCP server for repository: %s\n", repositoryPath)

	serverContext := context.Background()
	if err := server.Run(serverContext); err != nil {
		log.Fatalf("MCP server terminated with error: %v", err)
	}
}
