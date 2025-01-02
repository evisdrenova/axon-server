package cmd

import (
	"fmt"
	"log"
	"os"

	handlers "github.com/evisdrenova/axon-server/handlers"
	"github.com/evisdrenova/axon-server/parser"
	"github.com/evisdrenova/axon-server/server"
)

func main() {
	s := server.NewMCPServer(
		"Axon",
		"0.0.1",
	)

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-api-spec>\n", os.Args[0])
		os.Exit(1)
	}

	specPath := os.Args[1]

	// Parse the spec
	tools, err := parser.ParseSpecRouter(specPath)
	if err != nil {
		log.Fatalf("Unable to convert spec: %v", err)
	}

	// Register tools with the server
	for _, tool := range tools {
		s.AddTool(tool, handlers.CreateOpenAPIMCPToolHandler(tool))
	}

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
