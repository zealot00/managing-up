package gateway

import (
	"encoding/json"
	"fmt"
	"strings"
)

type OpenAPIImporter struct{}

type AdapterTemplate struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	BaseURL       string            `json:"base_url"`
	Auth          AuthConfig        `json:"auth"`
	Endpoints     []EndpointTemplate `json:"endpoints"`
	InputMapping  map[string]string `json:"input_mapping"`
	OutputRules   []OutputRule      `json:"output_rules"`
}

type AuthConfig struct {
	Type    string `json:"type"`
	Token   string `json:"token,omitempty"`
	APIKey  string `json:"api_key,omitempty"`
	Header  string `json:"header,omitempty"`
}

type EndpointTemplate struct {
	Path        string          `json:"path"`
	Method      string          `json:"method"`
	Description string          `json:"description"`
	InputFields []FieldTemplate `json:"input_fields"`
	OutputFields []string       `json:"output_fields"`
}

type FieldTemplate struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

type OutputRule struct {
	Type      string   `json:"type"`
	Fields    []string `json:"fields,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
}

func NewOpenAPIImporter() *OpenAPIImporter {
	return &OpenAPIImporter{}
}

func (i *OpenAPIImporter) Import(specJSON []byte) (*AdapterTemplate, error) {
	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	info, _ := spec["info"].(map[string]interface{})
	servers, _ := spec["servers"].([]interface{})
	paths, _ := spec["paths"].(map[string]interface{})

	template := &AdapterTemplate{
		ID:        fmt.Sprintf("adapter_%d", len(spec)),
		Name:      getString(info, "title", "Untitled API"),
		BaseURL:   getStringFromServers(servers),
		Endpoints: []EndpointTemplate{},
	}

	for path, pathItem := range paths {
		if methods, ok := pathItem.(map[string]interface{}); ok {
			for method, operation := range methods {
				method = strings.ToUpper(method)
				if method == "GET" || method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
					template.Endpoints = append(template.Endpoints, i.parseOperation(path, method, operation))
				}
			}
		}
	}

	return template, nil
}

func (i *OpenAPIImporter) parseOperation(path, method string, operation interface{}) EndpointTemplate {
	op, _ := operation.(map[string]interface{})
	return EndpointTemplate{
		Path:        path,
		Method:      method,
		Description: getString(op, "summary", ""),
		InputFields: i.extractInputFields(op),
		OutputFields: []string{"data"},
	}
}

func (i *OpenAPIImporter) extractInputFields(op map[string]interface{}) []FieldTemplate {
	var fields []FieldTemplate

	params, _ := op["parameters"].([]interface{})
	for _, p := range params {
		if param, ok := p.(map[string]interface{}); ok {
			fields = append(fields, FieldTemplate{
				Name:     getString(param, "name", ""),
				In:       getString(param, "in", "query"),
				Required: getBool(param, "required", false),
				Type:     getString(param, "type", "string"),
			})
		}
	}

	return fields
}

func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultVal
}

func getStringFromServers(servers []interface{}) string {
	if len(servers) > 0 {
		if s, ok := servers[0].(map[string]interface{}); ok {
			return getString(s, "url", "http://localhost")
		}
	}
	return "http://localhost"
}
