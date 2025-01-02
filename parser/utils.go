package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/evisdrenova/axon-server/mcp"
)

type specVersion struct {
	Swagger string `json:"swagger"`
	OpenAPI string `json:"openapi"`
}

func DetectSpecVersion(specPath string) (string, error) {
	file, err := os.Open(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to open spec file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read spec file: %w", err)
	}

	var version specVersion
	if err := json.Unmarshal(content, &version); err != nil {
		// Try YAML format if JSON fails
		// Convert common YAML version indicators to JSON
		yamlContent := string(content)
		if strings.Contains(yamlContent, "swagger: \"2.0\"") {
			return "2.0", nil
		}
		if strings.Contains(yamlContent, "openapi: \"3.") {
			return "3.0", nil
		}
		return "", fmt.Errorf("failed to parse spec version: %w", err)
	}

	if version.Swagger != "" {
		return version.Swagger, nil
	}
	if version.OpenAPI != "" {
		return version.OpenAPI, nil
	}

	return "", fmt.Errorf("no version information found in spec")
}

func ParseSpecRouter(specPath string) ([]mcp.Tool, error) {
	var tools []mcp.Tool

	// Detect spec version
	version, err := DetectSpecVersion(specPath)
	if err != nil {
		return nil, fmt.Errorf("error detecting spec version: %v", err)
	}

	// TODO: we could probably make this better by checking the first field of the
	// spec either swagger or openapi instead of using the version
	if strings.HasPrefix(version, "2.") {
		// Swagger 2.0
		swaggerSpec, err := LoadSwaggerSpec(specPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load swagger spec")
		}

		tools, err = ConvertSwaggerToMCPTools(swaggerSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Swagger spec to MCP tools: %v", err)
		}
	} else if strings.HasPrefix(version, "3.") {
		openApiSpec, err := LoadOpenApiSpec(specPath)
		if err != nil {
			return nil, fmt.Errorf("error loading OpenAPI spec: %v", err)
		}

		tools, err = ConvertOpenAPIToMCPTools(openApiSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert OpenAPI spec to MCP tools: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported specification version: %s", version)
	}

	return tools, nil
}
