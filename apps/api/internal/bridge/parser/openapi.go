package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type OpenAPISpec struct {
	Version     string                 `json:"openapi" yaml:"openapi"`
	Version2    string                 `json:"swagger" yaml:"swagger"`
	Info        SpecInfo               `json:"info" yaml:"info"`
	Servers     []Server               `json:"servers" yaml:"servers"`
	Paths       map[string]PathItem    `json:"paths" yaml:"paths"`
	Components  Components             `json:"components" yaml:"components"`
	BaseURL     string                 `json:"url" yaml:"url"`
}

type SpecInfo struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`
}

type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`
}

type PathItem struct {
	Ref        string     `json:"$ref" yaml:"$ref"`
	Summary    string     `json:"summary" yaml:"summary"`
	Description string    `json:"description" yaml:"description"`
	Get        *Operation `json:"get" yaml:"get"`
	Post       *Operation `json:"post" yaml:"post"`
	Put        *Operation `json:"put" yaml:"put"`
	Delete     *Operation `json:"delete" yaml:"delete"`
	Patch      *Operation `json:"patch" yaml:"patch"`
	Options    *Operation `json:"options" yaml:"options"`
	Head       *Operation `json:"head" yaml:"head"`
	Trace      *Operation `json:"trace" yaml:"trace"`
	Parameters []Parameter `json:"parameters" yaml:"parameters"`
}

type Operation struct {
	OperationID     string                 `json:"operationId" yaml:"operationId"`
	Summary         string                 `json:"summary" yaml:"summary"`
	Description     string                 `json:"description" yaml:"description"`
	Parameters      []Parameter            `json:"parameters" yaml:"parameters"`
	RequestBody     *RequestBody           `json:"requestBody" yaml:"requestBody"`
	Responses       map[string]Response    `json:"responses" yaml:"responses"`
	Deprecated      bool                   `json:"deprecated" yaml:"deprecated"`
	Tags            []string               `json:"tags" yaml:"tags"`
}

type Parameter struct {
	Name        string             `json:"name" yaml:"name"`
	In          string             `json:"in" yaml:"in"`
	Description string             `json:"description" yaml:"description"`
	Required    bool               `json:"required" yaml:"required"`
	Deprecated  bool               `json:"deprecated" yaml:"deprecated"`
	Schema      *Schema            `json:"schema" yaml:"schema"`
}

type RequestBody struct {
	Description string                 `json:"description" yaml:"description"`
	Required    bool                   `json:"required" yaml:"required"`
	Content     map[string]MediaType   `json:"content" yaml:"content"`
}

type Response struct {
	Description string                 `json:"description" yaml:"description"`
	Content     map[string]MediaType   `json:"content" yaml:"content"`
}

type MediaType struct {
	Schema   *Schema                   `json:"schema" yaml:"schema"`
	Example  any                       `json:"example" yaml:"example"`
	Examples map[string]Example        `json:"examples" yaml:"examples"`
}

type Example struct {
	Summary string `json:"summary" yaml:"summary"`
	Value   any    `json:"value" yaml:"value"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas" yaml:"schemas"`
}

type Schema struct {
	Type        string                 `json:"type" yaml:"type"`
	Format      string                 `json:"format" yaml:"format"`
	Description string                 `json:"description" yaml:"description"`
	Properties   map[string]Schema     `json:"properties" yaml:"properties"`
	Items        *Schema                `json:"items" yaml:"items"`
	Required     []string               `json:"required" yaml:"required"`
	Enum         []any                  `json:"enum" yaml:"enum"`
	Ref          string                 `json:"$ref" yaml:"$ref"`
}

type ParsedEndpoint struct {
	OperationID   string
	Method        string
	Path          string
	Summary       string
	Description   string
	Parameters    []ParsedParameter
	RequestBody   *ParsedRequestBody
	ResponseSchema *Schema
	Tags          []string
}

type ParsedParameter struct {
	Name        string
	In          string
	Required    bool
	Description string
	Schema      *Schema
}

type ParsedRequestBody struct {
	Required    bool
	Schema      *Schema
	ContentType string
}

func ParseOpenAPI(data []byte) (*OpenAPISpec, error) {
	spec := &OpenAPISpec{}

	if isYAML(data) {
		if err := yaml.Unmarshal(data, spec); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, spec); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	if err := spec.validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	return spec, nil
}

func isYAML(data []byte) bool {
	var js struct{}
	if json.Unmarshal(data, &js) == nil {
		return false
	}
	return true
}

func (s *OpenAPISpec) validate() error {
	if s.Version == "" && s.Version2 == "" {
		return fmt.Errorf("missing OpenAPI version")
	}
	if s.Version != "" && !strings.HasPrefix(s.Version, "3.") {
		return fmt.Errorf("unsupported OpenAPI version: %s (only 3.x supported)", s.Version)
	}
	if s.Version2 != "" && !strings.HasPrefix(s.Version2, "2.") {
		return fmt.Errorf("unsupported Swagger version: %s (only 2.x supported)", s.Version2)
	}
	if s.Paths == nil || len(s.Paths) == 0 {
		return fmt.Errorf("no paths defined")
	}
	return nil
}

func (s *OpenAPISpec) GetBaseURL() string {
	if len(s.Servers) > 0 {
		return s.Servers[0].URL
	}
	return s.BaseURL
}

func (s *OpenAPISpec) GetVersion() string {
	if s.Version != "" {
		return s.Version
	}
	return s.Version2
}

func (s *OpenAPISpec) ParseEndpoints() []ParsedEndpoint {
	var endpoints []ParsedEndpoint

	for path, pathItem := range s.Paths {
		endpoints = append(endpoints, s.parsePathItem(path, pathItem)...)
	}

	return endpoints
}

func (s *OpenAPISpec) parsePathItem(path string, item PathItem) []ParsedEndpoint {
	var endpoints []ParsedEndpoint

	operations := map[string]*Operation{
		"GET":    item.Get,
		"POST":   item.Post,
		"PUT":    item.Put,
		"DELETE": item.Delete,
		"PATCH":  item.Patch,
	}

	for method, op := range operations {
		if op == nil {
			continue
		}

		params := s.mergeParameters(item.Parameters, op.Parameters)
		endpoint := ParsedEndpoint{
			OperationID: op.OperationID,
			Method:      method,
			Path:        path,
			Summary:     op.Summary,
			Description: op.Description,
			Parameters:  s.parseParameters(params),
			Tags:        op.Tags,
		}

		if op.RequestBody != nil {
			endpoint.RequestBody = s.parseRequestBody(op.RequestBody)
		}

		if resp, ok := op.Responses["200"]; ok {
			if mt, ok := resp.Content["application/json"]; ok && mt.Schema != nil {
				endpoint.ResponseSchema = mt.Schema
			}
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func (s *OpenAPISpec) mergeParameters(pathParams, opParams []Parameter) []Parameter {
	paramMap := make(map[string]Parameter)

	for _, p := range pathParams {
		key := p.Name + "_" + p.In
		paramMap[key] = p
	}

	for _, p := range opParams {
		key := p.Name + "_" + p.In
		paramMap[key] = p
	}

	var result []Parameter
	for _, p := range paramMap {
		result = append(result, p)
	}

	return result
}

func (s *OpenAPISpec) parseParameters(params []Parameter) []ParsedParameter {
	var result []ParsedParameter
	for _, p := range params {
		result = append(result, ParsedParameter{
			Name:        p.Name,
			In:          p.In,
			Required:    p.Required,
			Description: p.Description,
			Schema:      p.Schema,
		})
	}
	return result
}

func (s *OpenAPISpec) parseRequestBody(body *RequestBody) *ParsedRequestBody {
	if body == nil {
		return nil
	}

	parsed := &ParsedRequestBody{
		Required: body.Required,
	}

	for contentType, mt := range body.Content {
		if parsed.ContentType == "" {
			parsed.ContentType = contentType
		}
		if mt.Schema != nil {
			parsed.Schema = mt.Schema
			break
		}
	}

	return parsed
}

func (s *OpenAPISpec) ResolveSchema(ref string) *Schema {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil
	}

	name := strings.TrimPrefix(ref, "#/components/schemas/")
	if s.Components.Schemas != nil {
		if schema, ok := s.Components.Schemas[name]; ok {
			return &schema
		}
	}
	return nil
}