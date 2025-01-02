package parser

import (
	"fmt"
	"net/url"

	"github.com/evisdrenova/axon-server/mcp"
	"github.com/getkin/kin-openapi/openapi3"
)

// Converts an OpenAPI spec to an array of MCP Tools that a Host can recognize
func ConvertOpenAPIToMCPTools(spec *openapi3.T) ([]mcp.Tool, error) {
	var tools []mcp.Tool

	// Extract the base url
	// TODO: update this to handle multiple servers
	baseUrl := spec.Servers[0].URL

	// Iterate through all paths
	for path, method := range spec.Paths.Map() {
		if method == nil {
			continue
		}

		// construct the entire path
		fullPath := baseUrl + path

		if method.Get != nil {
			tool, err := ConvertOperationToMCPTool(method.Get, "GET", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert GET operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		if method.Post != nil {
			tool, err := ConvertOperationToMCPTool(method.Post, "POST", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert POST operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		if method.Put != nil {
			tool, err := ConvertOperationToMCPTool(method.Put, "PUT", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert PUT operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		if method.Delete != nil {
			tool, err := ConvertOperationToMCPTool(method.Delete, "DELETE", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert DELETE operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		if method.Patch != nil {
			tool, err := ConvertOperationToMCPTool(method.Patch, "PATCH", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert PATCH operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}
	}

	return tools, nil
}

func ConvertOperationToMCPTool(
	operation *openapi3.Operation,
	method string,
	path string,
) (*mcp.Tool, error) {
	if operation.OperationID == "" {
		return nil, nil
	}

	// Create properties map for the tool schema
	properties := make(map[string]interface{})
	required := []string{}

	// Add endpoint information
	properties["endpoint"] = map[string]interface{}{
		"type":  "string",
		"const": path,
	}
	properties["method"] = map[string]interface{}{
		"type":  "string",
		"const": method,
	}

	// Handle path parameters
	for _, param := range operation.Parameters {
		if param.Value == nil {
			continue
		}

		schema := ConvertSchemaToMap(param.Value.Schema.Value)
		if param.Value.Description != "" {
			schema["description"] = param.Value.Description
		}

		properties[param.Value.Name] = schema
		if param.Value.Required {
			required = append(required, param.Value.Name)
		}
	}

	// Handle request body
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		if jsonContent := operation.RequestBody.Value.Content.Get("application/json"); jsonContent != nil {
			bodySchema := ConvertSchemaToMap(jsonContent.Schema.Value)
			properties["body"] = bodySchema
			if operation.RequestBody.Value.Required {
				required = append(required, "body")
			}
		}
	}

	// Build description
	description := operation.Summary
	if description == "" {
		description = operation.Description
	}
	if description == "" {
		description = fmt.Sprintf("%s %s", method, path)
	}

	return &mcp.Tool{
		Name:        operation.OperationID,
		Description: description,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: properties,
			Required:   required,
		},
	}, nil
}

// Gets the schema type for the Open API spec
func GetSchemaType(types *openapi3.Types) string {
	if types == nil || len(*types) == 0 {
		return "object" // default to object type
	}
	return (*types)[0] // take the first type
}

func ConvertSchemaToMap(schema *openapi3.Schema) map[string]interface{} {
	if schema == nil {
		return map[string]interface{}{"type": "object"}
	}

	result := make(map[string]interface{})

	// Handle basic properties
	// schema.Type is a string slice
	if len(GetSchemaType(schema.Type)) > 0 {
		result["type"] = "object"
	}
	if schema.Format != "" {
		result["format"] = schema.Format
	}
	if schema.Description != "" {
		result["description"] = schema.Description
	}

	// Handle array type
	if schema.Items != nil && schema.Items.Value != nil {
		result["items"] = ConvertSchemaToMap(schema.Items.Value)
	}

	// Handle object properties
	if schema.Properties != nil {
		props := make(map[string]interface{})
		for name, prop := range schema.Properties {
			props[name] = ConvertSchemaToMap(prop.Value)
		}
		result["properties"] = props
		if len(schema.Required) > 0 {
			result["required"] = schema.Required
		}
	}

	// Handle constraints
	if schema.Min != nil {
		result["minimum"] = *schema.Min
	}
	if schema.Max != nil {
		result["maximum"] = *schema.Max
	}
	if schema.MinLength != 0 {
		result["minLength"] = schema.MinLength
	}
	if schema.MaxLength != nil {
		result["maxLength"] = *schema.MaxLength
	}
	if schema.Pattern != "" {
		result["pattern"] = schema.Pattern
	}
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	return result
}

// Checks if the arg passed in is a URL
func IsURL(url string) bool {
	return len(url) > 8 && (url[:7] == "http://" || url[:8] == "https://")
}

// Load in a OpenApi spec. Will return an error if the spec is not valid.
func LoadOpenApiSpec(specPath string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	var doc *openapi3.T
	var err error

	if IsURL(specPath) {
		doc, err = loader.LoadFromURI(&url.URL{Path: specPath})
	} else {
		doc, err = loader.LoadFromFile(specPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	err = doc.Validate(loader.Context)
	if err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	return doc, nil
}
