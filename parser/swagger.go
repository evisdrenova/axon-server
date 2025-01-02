package parser

import (
	"fmt"

	"github.com/evisdrenova/axon-server/mcp"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// ConvertSwaggerToMCPTools converts a Swagger 2.0 spec to an array of MCP Tools
func ConvertSwaggerToMCPTools(swaggerDoc *spec.Swagger) ([]mcp.Tool, error) {
	var tools []mcp.Tool

	// Extract base URL from the swagger doc
	baseURL := swaggerDoc.BasePath
	if swaggerDoc.Host != "" {
		scheme := "https"
		if len(swaggerDoc.Schemes) > 0 {
			scheme = swaggerDoc.Schemes[0]
		}
		baseURL = fmt.Sprintf("%s://%s%s", scheme, swaggerDoc.Host, baseURL)
	}

	// Iterate through all paths
	for path, pathItem := range swaggerDoc.Paths.Paths {
		fullPath := baseURL + path

		// Handle GET operations
		if pathItem.Get != nil {
			tool, err := convertOperationToMCPTool(pathItem.Get, "GET", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert GET operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		// Handle POST operations
		if pathItem.Post != nil {
			tool, err := convertOperationToMCPTool(pathItem.Post, "POST", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert POST operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		// Handle PUT operations
		if pathItem.Put != nil {
			tool, err := convertOperationToMCPTool(pathItem.Put, "PUT", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert PUT operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		// Handle DELETE operations
		if pathItem.Delete != nil {
			tool, err := convertOperationToMCPTool(pathItem.Delete, "DELETE", fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to convert DELETE operation for %s: %w", fullPath, err)
			}
			if tool != nil {
				tools = append(tools, *tool)
			}
		}

		// Handle PATCH operations
		if pathItem.Patch != nil {
			tool, err := convertOperationToMCPTool(pathItem.Patch, "PATCH", fullPath)
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

func convertOperationToMCPTool(
	operation *spec.Operation,
	method string,
	path string,
) (*mcp.Tool, error) {
	if operation.ID == "" {
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

	// Handle parameters
	for _, param := range operation.Parameters {
		if param.Name == "" {
			continue
		}

		schema := convertSchemaToMap(param.Schema)
		if param.Description != "" {
			schema["description"] = param.Description
		}

		properties[param.Name] = schema
		if param.Required {
			required = append(required, param.Name)
		}

		// Handle body parameter specifically
		if param.In == "body" && param.Schema != nil {
			properties["body"] = convertSchemaToMap(param.Schema)
			if param.Required {
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
		Name:        operation.ID,
		Description: description,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: properties,
			Required:   required,
		},
	}, nil
}

func convertSchemaToMap(schema *spec.Schema) map[string]interface{} {
	if schema == nil {
		return map[string]interface{}{"type": "object"}
	}

	result := make(map[string]interface{})

	// Handle basic properties
	if schema.Type.Contains("array") {
		result["type"] = "array"
		if schema.Items != nil && schema.Items.Schema != nil {
			result["items"] = convertSchemaToMap(schema.Items.Schema)
		}
	} else if len(schema.Type) > 0 {
		result["type"] = schema.Type[0]
	}

	if schema.Format != "" {
		result["format"] = schema.Format
	}
	if schema.Description != "" {
		result["description"] = schema.Description
	}

	// Handle object properties
	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for name, prop := range schema.Properties {
			props[name] = convertSchemaToMap(&prop)
		}
		result["properties"] = props
	}

	// Handle required fields
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	// Handle constraints
	if schema.Minimum != nil {
		result["minimum"] = *schema.Minimum
	}
	if schema.Maximum != nil {
		result["maximum"] = *schema.Maximum
	}
	if schema.MinLength != nil {
		result["minLength"] = *schema.MinLength
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

// LoadSwaggerSpec loads a Swagger 2.0 specification from a file or URL
func LoadSwaggerSpec(specPath string) (*spec.Swagger, error) {
	var doc *loads.Document
	var err error

	if isURL(specPath) {
		doc, err = loads.Spec(specPath)
	} else {
		doc, err = loads.Spec(specPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load Swagger spec: %w", err)
	}

	// Create a new validator
	validator := validate.NewSpecValidator(doc.Schema(), strfmt.Default)

	// Validate the document
	result, _ := validator.Validate(doc)
	if result.HasErrors() {
		return nil, fmt.Errorf("invalid Swagger spec: %v", result.Errors)
	}

	return doc.Spec(), nil
}

func isURL(str string) bool {
	return len(str) > 8 && (str[:7] == "http://" || str[:8] == "https://")
}
