package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/evisdrenova/axon-server/mcp"
)

// MCPClient is the top-level client that manages all LLM clients
type MCPClient struct {
	anthropic *AnthropicClient
}

// Config holds all of the underlying LLM configs
type Config struct {
	AnthropicConfig *AnthropicConfig
}

func NewMCPClient(config *Config, serverPath string) (*MCPClient, error) {
	var client *MCPClient
	var anthropicClient *AnthropicClient
	var err error

	if config.AnthropicConfig != nil {
		anthropicClient, err = NewAnthropicClient(config.AnthropicConfig, serverPath)
		if err != nil {
			return nil, err
		}
	}

	// Create MCPClient with available components
	client = &MCPClient{
		anthropic: anthropicClient,
	}

	return client, nil
}

func (m *MCPClient) GetAnthropicClient() *AnthropicClient {
	return m.anthropic
}

func (m *MCPClient) CleanUp() error {
	if m.anthropic != nil {
		m.anthropic.Close()
	}
	return nil
}

// start the chat loop for between the user <> client
func (m *MCPClient) ChatLoop() error {

	fmt.Println("\nMCP Client Started!")
	fmt.Println("Type your queries or 'quit' to exit.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nQuery: ")
		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())
		if strings.ToLower(query) == "quit" {
			break
		}

		if err := m.processQuery(query); err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	return nil
}

// process the user query
func (m *MCPClient) processQuery(query string) error {
	ctx := context.Background()

	toolsResult, err := m.anthropic.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}

	var availableTools []mcp.Tool
	for _, tool := range toolsResult.Tools {
		availableTools = append(availableTools, mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}

	fmt.Println("query", query, availableTools)

	// // Create anthropic messages request
	// // Note: This is a placeholder - you'll need to implement the actual Anthropic API client
	// messages := []map[string]interface{}{
	// 	{
	// 		"role":    "user",
	// 		"content": query,
	// 	},
	// }

	// // Make request to Anthropic API
	// // TODO: Implement actual Anthropic API client
	// response, err := m.CallAnthropic(messages, availableTools)
	// if err != nil {
	// 	return fmt.Errorf("failed to call Anthropic API: %v", err)
	// }

	// // Process tool calls and continue conversation
	// toolResults := []map[string]interface{}{}
	// for _, content := range response.Content {
	// 	if content.Type == "text" {
	// 		fmt.Println(content.Text)
	// 	} else if content.Type == "tool_use" {
	// 		fmt.Printf("[Calling tool %s with args %v]\n", content.Name, content.Input)

	// 		result, err := m.client.CallTool(ctx, mcp.CallToolRequest{
	// 			Params: struct {
	// 				Name      string                 `json:"name"`
	// 				Arguments map[string]interface{} `json:"arguments,omitempty"`
	// 			}{
	// 				Name:      content.Name,
	// 				Arguments: content.Input.(map[string]interface{}),
	// 			},
	// 		})
	// 		if err != nil {
	// 			return fmt.Errorf("failed to call tool: %v", err)
	// 		}

	// 		toolResults = append(toolResults, map[string]interface{}{
	// 			"call":   content.Name,
	// 			"result": result,
	// 		})

	// 		// Add tool result to conversation
	// 		messages = append(messages, map[string]interface{}{
	// 			"role":    "assistant",
	// 			"content": content.Text,
	// 		})
	// 		messages = append(messages, map[string]interface{}{
	// 			"role":    "user",
	// 			"content": result.Content,
	// 		})

	// 		// Get next response from Claude
	// 		response, err = m.callAnthropic(messages, availableTools)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to call Anthropic API: %v", err)
	// 		}

	// 		fmt.Println(response.Content[0].Text)
	// 	}
	// }

	return nil
}
