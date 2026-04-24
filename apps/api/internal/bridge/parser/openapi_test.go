package parser

import (
	"strings"
	"testing"
)

func TestParseOpenAPI3(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/users/{id}": {
				"get": {
					"operationId": "getUser",
					"summary": "Get a user",
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": {"type": "string"}
						}
					],
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {"type": "object"}
								}
							}
						}
					}
				}
			}
		}
	}`

	data := []byte(spec)
	parsed, err := ParseOpenAPI(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed.Version != "3.0.0" {
		t.Errorf("expected version 3.0.0, got %s", parsed.Version)
	}

	if parsed.Info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", parsed.Info.Title)
	}

	endpoints := parsed.ParseEndpoints()
	if len(endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(endpoints))
	}

	if endpoints[0].OperationID != "getUser" {
		t.Errorf("expected operationId 'getUser', got %s", endpoints[0].OperationID)
	}

	if endpoints[0].Method != "GET" {
		t.Errorf("expected method 'GET', got %s", endpoints[0].Method)
	}

	if endpoints[0].Path != "/users/{id}" {
		t.Errorf("expected path '/users/{id}', got %s", endpoints[0].Path)
	}
}

func TestParseOpenAPIWithYAML(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: YAML API
  version: 2.0.0
paths:
  /items:
    post:
      operationId: createItem
      summary: Create an item
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
      responses:
        "201":
          description: Created
`

	data := []byte(spec)
	parsed, err := ParseOpenAPI(data)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	if parsed.Version != "3.0.0" {
		t.Errorf("expected version 3.0.0, got %s", parsed.Version)
	}

	if parsed.Info.Title != "YAML API" {
		t.Errorf("expected title 'YAML API', got %s", parsed.Info.Title)
	}

	endpoints := parsed.ParseEndpoints()
	if len(endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(endpoints))
	}

	if endpoints[0].OperationID != "createItem" {
		t.Errorf("expected operationId 'createItem', got %s", endpoints[0].OperationID)
	}
}

func TestIsYAML(t *testing.T) {
	jsonData := []byte(`{"openapi": "3.0.0"}`)
	if isYAML(jsonData) {
		t.Error("JSON data should not be detected as YAML")
	}

	yamlData := []byte("openapi: 3.0.0\ninfo:\n  title: Test")
	if !isYAML(yamlData) {
		t.Error("YAML data should be detected as YAML")
	}
}

func TestGetBaseURL(t *testing.T) {
	spec := &OpenAPISpec{
		Servers: []Server{
			{URL: "https://api.example.com/v1"},
		},
	}

	baseURL := spec.GetBaseURL()
	if baseURL != "https://api.example.com/v1" {
		t.Errorf("expected 'https://api.example.com/v1', got %s", baseURL)
	}
}

func TestResolveSchema(t *testing.T) {
	spec := &OpenAPISpec{
		Components: Components{
			Schemas: map[string]Schema{
				"User": {
					Type: "object",
					Properties: map[string]Schema{
						"name": {Type: "string"},
						"email": {Type: "string"},
					},
				},
			},
		},
	}

	schema := spec.ResolveSchema("#/components/schemas/User")
	if schema == nil {
		t.Fatal("expected schema to be resolved")
	}

	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(schema.Properties))
	}

	schema = spec.ResolveSchema("#/components/schemas/NonExistent")
	if schema != nil {
		t.Error("expected nil for non-existent schema")
	}
}

func TestMergeParameters(t *testing.T) {
	spec := &OpenAPISpec{}

	pathParams := []Parameter{
		{Name: "id", In: "path", Required: true},
	}

	opParams := []Parameter{
		{Name: "id", In: "path", Required: true},
		{Name: "filter", In: "query", Required: false},
	}

	merged := spec.mergeParameters(pathParams, opParams)
	if len(merged) != 2 {
		t.Errorf("expected 2 merged parameters, got %d", len(merged))
	}
}

func TestParseEndpoints(t *testing.T) {
	spec := &OpenAPISpec{
		Paths: map[string]PathItem{
			"/posts": {
				Get: &Operation{
					OperationID: "listPosts",
					Summary:    "List posts",
				},
				Post: &Operation{
					OperationID: "createPost",
					Summary:    "Create post",
				},
			},
			"/posts/{id}": {
				Get: &Operation{
					OperationID: "getPost",
					Summary:    "Get post",
					Parameters: []Parameter{
						{Name: "id", In: "path", Required: true},
					},
				},
				Delete: &Operation{
					OperationID: "deletePost",
					Summary:    "Delete post",
					Parameters: []Parameter{
						{Name: "id", In: "path", Required: true},
					},
				},
			},
		},
	}

	endpoints := spec.ParseEndpoints()
	if len(endpoints) != 4 {
		t.Errorf("expected 4 endpoints, got %d", len(endpoints))
	}

	methods := make(map[string]bool)
	for _, ep := range endpoints {
		methods[ep.Method] = true
	}

	if !methods["GET"] || !methods["POST"] || !methods["DELETE"] {
		t.Error("expected GET, POST, DELETE methods")
	}
}

func TestValidateSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    *OpenAPISpec
		wantErr bool
	}{
		{
			name:    "missing version",
			spec:    &OpenAPISpec{},
			wantErr: true,
		},
		{
			name: "unsupported version",
			spec: &OpenAPISpec{
				Version: "4.0.0",
				Paths:   map[string]PathItem{},
			},
			wantErr: true,
		},
		{
			name: "no paths",
			spec: &OpenAPISpec{
				Version: "3.0.0",
				Paths:   nil,
			},
			wantErr: true,
		},
		{
			name: "valid spec",
			spec: &OpenAPISpec{
				Version: "3.0.0",
				Paths: map[string]PathItem{
					"/test": {Get: &Operation{OperationID: "test"}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenAPISpec_String(t *testing.T) {
	spec := &OpenAPISpec{
		Version: "3.0.0",
		Info: SpecInfo{
			Title:       "Test",
			Description: "A test API",
			Version:     "1.0.0",
		},
	}

	str := spec.GetVersion()
	if !strings.Contains(str, "3.0") {
		t.Errorf("GetVersion() = %s, want contains 3.0", str)
	}
}