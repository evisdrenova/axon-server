package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/evisdrenova/axon-server/handlers/logger"
	"github.com/evisdrenova/axon-server/mcp"
	"github.com/evisdrenova/axon-server/server"
)

//TODO: finish swapping out internal paths for fully qualified path

// Creates a higher level funciton that encapsulates the logger and handler
// We use a logger here to test the handler
// We could return all of this to the Claude developer tools but i find that to be more annoying, so for now, we're just logger to an external file
func CreateOpenAPIMCPToolHandler(tool mcp.Tool) server.ToolHandlerFunc {
	// Try to create file logger
	logger, err := logger.CreateFileLogger()
	if err != nil {
		// Fall back to stderr logger if file creation fails
		logger = log.New(os.Stderr, "", log.LstdFlags)
		log.Printf("Falling back to stderr logging due to error: %v", err)
	}

	// Return the handler with the appropriate logger
	return createHandler(tool, logger)
}

// Handler that spins up an http server that claude actually calls as part of the MCP process
// this handler can really be anything! It doesn't have to be an http server, it can be a wasm module, or anything else!
func createHandler(tool mcp.Tool, logger *log.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client := &http.Client{}

		schema := tool.InputSchema.Properties
		endpoint, ok := schema["endpoint"].(map[string]interface{})
		if !ok || endpoint["const"] == nil {
			return mcp.NewToolResultError("Endpoint configuration not found in tool schema"), nil
		}

		method, ok := schema["method"].(map[string]interface{})
		if !ok || method["const"] == nil {
			return mcp.NewToolResultError("Method configuration not found in tool schema"), nil
		}

		endpointStr := endpoint["const"].(string)
		methodStr := method["const"].(string)

		for paramName, paramValue := range request.Params.Arguments {
			if paramName != "body" && paramName != "endpoint" && paramName != "method" {
				placeholder := fmt.Sprintf("{%s}", paramName)
				if strings.Contains(endpointStr, placeholder) {
					endpointStr = strings.ReplaceAll(endpointStr, placeholder, fmt.Sprint(paramValue))
				}
			}
		}

		var reqBody io.Reader
		var bodyJSON []byte
		var err error
		if bodyData, ok := request.Params.Arguments["body"]; ok {
			bodyJSON, err = json.MarshalIndent(bodyData, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal request body: %v", err)), nil
			}
			reqBody = bytes.NewBuffer(bodyJSON)
		}

		req, err := http.NewRequestWithContext(ctx, methodStr, endpointStr, reqBody)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create request: %v", err)), nil
		}

		if reqBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Log request details
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			logger.Printf("Error dumping request: %v", err)
		} else {
			logger.Printf("REQUEST:\n%s\n", string(reqDump))
			if bodyJSON != nil {
				logger.Printf("REQUEST BODY:\n%s\n", string(bodyJSON))
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("REQUEST ERROR: %v\n", err)
			return mcp.NewToolResultError(fmt.Sprintf("Request failed: %v", err)), nil
		}
		defer resp.Body.Close()

		// Log response details
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			logger.Printf("Error dumping response: %v", err)
		} else {
			logger.Printf("RESPONSE:\n%s\n", string(respDump))
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read response: %v", err)), nil
		}

		if resp.StatusCode >= 400 {
			logger.Printf("ERROR RESPONSE: Status %d - %s\n", resp.StatusCode, string(respBody))
			return mcp.NewToolResultError(fmt.Sprintf("Request failed with status %d: %s", resp.StatusCode, string(respBody))), nil
		}

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
			logger.Printf("RESPONSE BODY (JSON):\n%s\n", prettyJSON.String())
			return mcp.NewToolResultText(prettyJSON.String()), nil
		}

		logger.Printf("RESPONSE BODY (Raw):\n%s\n", string(respBody))
		return mcp.NewToolResultText(string(respBody)), nil
	}
}
