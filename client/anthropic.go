package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/evisdrenova/axon-server/mcp"
)

type AnthropicConfig struct {
	APIKey   string
	Endpoint string
}

// AnthropicClient implements MCPClientInterface and other Anthropic-specific methods
type AnthropicClient struct {
	config *AnthropicConfig
	client *StdioMCPClient
}

type MessageRequest struct {
	Model     string                   `json:"model"`
	MaxTokens int                      `json:"max_tokens"`
	Messages  []map[string]interface{} `json:"messages"`
	Tools     []map[string]interface{} `json:"tools,omitempty"`
}

type AnthropicResponse struct {
	Content []struct {
		Type  string      `json:"type"`
		Text  string      `json:"text,omitempty"`
		Name  string      `json:"name,omitempty"`
		Input interface{} `json:"input,omitempty"`
	} `json:"content"`
}

func NewAnthropicClient(config *AnthropicConfig, serverPath string) (*AnthropicClient, error) {

	mcpClient, err := NewStdioMCPClient(serverPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &AnthropicClient{
		config: config,
		client: mcpClient,
	}, nil
}

func (c *AnthropicClient) ConnectToServer(serverPath string) error {
	// If serverPath is not an absolute path, resolve it relative to current directory
	if !filepath.IsAbs(serverPath) {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
		serverPath = filepath.Join(currentDir, serverPath)
	}

	var err error

	// Check if the file exists and is executable
	info, err := os.Stat(serverPath)
	if err != nil {
		return fmt.Errorf("server binary not found: %v", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("server binary is not executable: %s", serverPath)
	}

	mcpClient, err := NewStdioMCPClient(serverPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Initialize the client
	ctx := context.Background()
	params := mcp.InitializeRequest{}.Params
	params.Capabilities = mcp.ClientCapabilities{}
	params.ClientInfo = mcp.Implementation{
		Name:    "mcpcli",
		Version: "1.0.0",
	}
	params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION

	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: params,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize client: %v", err)
	}

	// List available tools
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}

	var toolNames []string
	for _, tool := range toolsResult.Tools {
		toolNames = append(toolNames, tool.Name)
	}
	fmt.Printf("\nConnected to server with tools: %v\n", toolNames)

	return nil

}

func (c *AnthropicClient) Close() error {
	return c.client.Close()
}
