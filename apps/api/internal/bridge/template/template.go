package template

import (
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/bridge/parser"
)

type AdapterTemplate struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	OpenAPISpec string        `json:"openapi_spec"`
	Mappings    []ToolMapping `json:"mappings"`
	Options     AdapterOptions `json:"options"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type ToolMapping struct {
	ToolName      string        `json:"tool_name"`
	OperationID   string        `json:"operation_id"`
	Description   string        `json:"description"`
	InputMappings []FieldMapping `json:"input_mappings"`
	OutputFilter  OutputFilter  `json:"output_filter"`
}

type FieldMapping struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required"`
}

type OutputFilter struct {
	Whitelist   []string `json:"whitelist,omitempty"`
	Blacklist   []string `json:"blacklist,omitempty"`
	MaxBytes    int64    `json:"max_bytes,omitempty"`
	UseSummary  bool     `json:"use_summary"`
}

type AdapterOptions struct {
	ResponseLimitBytes    int64 `json:"response_limit_bytes"`
	SummaryThresholdBytes  int64 `json:"summary_threshold_bytes"`
	TimeoutSeconds        int   `json:"timeout_seconds"`
}

type MCPTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema map[string]any    `json:"inputSchema"`
	Endpoint    ToolEndpoint      `json:"endpoint"`
}

type ToolEndpoint struct {
	Path   string            `json:"path"`
	Method string            `json:"method"`
	Parameters []EndpointParam `json:"parameters,omitempty"`
}

type EndpointParam struct {
	Name      string `json:"name"`
	In        string `json:"in"`
	Required  bool   `json:"required"`
}

type GeneratedAdapter struct {
	TemplateID string    `json:"template_id"`
	Tools      []MCPTool `json:"tools"`
	BaseURL    string    `json:"base_url"`
}

func (t *AdapterTemplate) GenerateTools(endpoints []parser.ParsedEndpoint) []MCPTool {
	var tools []MCPTool

	mappingMap := make(map[string]ToolMapping)
	for i := range t.Mappings {
		mappingMap[t.Mappings[i].OperationID] = t.Mappings[i]
	}

	for _, ep := range endpoints {
		mapping, hasMapping := mappingMap[ep.OperationID]

		toolName := ep.OperationID
		description := ep.Summary
		if ep.Description != "" {
			description = ep.Description
		}

		if hasMapping {
			if mapping.ToolName != "" {
				toolName = mapping.ToolName
			}
			if mapping.Description != "" {
				description = mapping.Description
			}
		}

		inputSchema := buildInputSchema(ep.Parameters, ep.RequestBody)

		tool := MCPTool{
			Name:        toolName,
			Description: description,
			InputSchema: inputSchema,
			Endpoint: ToolEndpoint{
				Path:   ep.Path,
				Method: ep.Method,
			},
		}

		for _, p := range ep.Parameters {
			tool.Endpoint.Parameters = append(tool.Endpoint.Parameters, EndpointParam{
				Name:     p.Name,
				In:       p.In,
				Required: p.Required,
			})
		}

		tools = append(tools, tool)
	}

	return tools
}

func buildInputSchema(params []parser.ParsedParameter, body *parser.ParsedRequestBody) map[string]any {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{},
		"required":   []string{},
	}

	required := []string{}

	for _, p := range params {
		prop := map[string]any{
			"type":        "string",
			"description": p.Description,
		}

		if p.Schema != nil {
			if p.Schema.Type != "" {
				prop["type"] = p.Schema.Type
			}
			if p.Schema.Format != "" {
				prop["format"] = p.Schema.Format
			}
			if len(p.Schema.Enum) > 0 {
				prop["enum"] = p.Schema.Enum
			}
		}

		schema["properties"].(map[string]any)[p.Name] = prop

		if p.Required {
			required = append(required, p.Name)
		}
	}

	if body != nil && body.Schema != nil {
		buildSchemaFromBody(schema, body.Schema)
	}

	schema["required"] = required
	return schema
}

func buildSchemaFromBody(schema map[string]any, bodySchema *parser.Schema) {
	if bodySchema == nil {
		return
	}

	if bodySchema.Type == "object" && bodySchema.Properties != nil {
		props := schema["properties"].(map[string]any)
		reqs := []string{}

		for name, propSchema := range bodySchema.Properties {
			prop := map[string]any{
				"type": "string",
			}

			if propSchema.Type != "" {
				prop["type"] = propSchema.Type
			}
			if propSchema.Description != "" {
				prop["description"] = propSchema.Description
			}

			props[name] = prop
		}

		if len(bodySchema.Required) > 0 {
			for _, req := range bodySchema.Required {
				reqs = append(reqs, req)
			}
		}

		schema["properties"] = props
		if len(reqs) > 0 {
			schema["required"] = reqs
		}
	}
}

func (t *AdapterTemplate) ApplyInputMapping(operationID string, input map[string]any) map[string]string {
	result := make(map[string]string)

	for _, m := range t.Mappings {
		if m.OperationID != operationID {
			continue
		}

		for _, im := range m.InputMappings {
			if val, ok := input[im.SourceField]; ok {
				result[im.TargetField] = fmt.Sprintf("%v", val)
			} else if im.Required && im.Default == "" {
				result[im.TargetField] = ""
			} else if im.Default != "" {
				result[im.TargetField] = im.Default
			}
		}
	}

	return result
}