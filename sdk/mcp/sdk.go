package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	name        string
	version     string
	description string
	platformURL string
	token       string
	serverID    string
	mcpServer   *server.MCPServer
	httpServer  *server.StreamableHTTPServer

	requestsTotal   int64
	requestsSuccess int64
	requestsError   int64
	latencySum      float64
	latencyMu       sync.Mutex
	toolCalls       map[string]int64
	toolMu          sync.RWMutex
}

type Config struct {
	Name        string
	Version     string
	Description string
}

func NewServer(config Config) *Server {
	platformURL := os.Getenv("MANAGING_UP_PLATFORM_URL")
	if platformURL == "" {
		platformURL = "http://localhost:8080"
	}

	name := config.Name
	if name == "" {
		name = os.Getenv("MANAGING_UP_NAME")
	}
	version := config.Version
	if version == "" {
		version = os.Getenv("MANAGING_UP_VERSION")
		if version == "" {
			version = "1.0.0"
		}
	}

	return &Server{
		name:        name,
		version:     version,
		description: config.Description,
		platformURL: platformURL,
		token:       os.Getenv("MANAGING_UP_TOKEN"),
		mcpServer: server.NewMCPServer(name, version,
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
		toolCalls: make(map[string]int64),
	}
}

func (s *Server) AddTool(tool mcp.Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	s.toolMu.Lock()
	defer s.toolMu.Unlock()
	s.toolCalls[tool.Name] = 0
	s.mcpServer.AddTool(tool, handler)
}

type Metrics struct {
	RequestsTotal   int64            `json:"requests_total"`
	RequestsSuccess int64            `json:"requests_success"`
	RequestsError   int64            `json:"requests_error"`
	LatencySeconds  float64          `json:"latency_seconds"`
	ToolCalls       map[string]int64 `json:"tool_calls"`
}

func (s *Server) Metrics() Metrics {
	s.toolMu.RLock()
	defer s.toolMu.RUnlock()
	s.latencyMu.Lock()
	latency := s.latencySum
	s.latencyMu.Unlock()
	return Metrics{
		RequestsTotal:   atomic.LoadInt64(&s.requestsTotal),
		RequestsSuccess: atomic.LoadInt64(&s.requestsSuccess),
		RequestsError:   atomic.LoadInt64(&s.requestsError),
		LatencySeconds:  latency,
		ToolCalls:       s.toolCalls,
	}
}

type RegistrationRequest struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Description   string `json:"description,omitempty"`
	URL           string `json:"url,omitempty"`
	TransportType string `json:"transport_type"`
	Status        string `json:"status"`
}

type RegistrationResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *Server) Register(ctx context.Context) (*RegistrationResponse, error) {
	req := RegistrationRequest{
		Name:          s.name,
		Version:       s.version,
		Description:   s.description,
		TransportType: "http",
		Status:        "pending",
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := s.platformURL + "/api/v1/mcp-servers"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data RegistrationResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	s.serverID = result.Data.ID
	return &result.Data, nil
}

func (s *Server) StartStdio(ctx context.Context) error {
	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("failed to serve stdio: %w", err)
	}
	return nil
}

func (s *Server) StartHTTP(ctx context.Context, addr string) error {
	s.httpServer = server.NewStreamableHTTPServer(s.mcpServer)
	log.Printf("Starting MCP server %s@%s on %s", s.name, s.version, addr)
	log.Printf("MCP endpoint: http://localhost%s/mcp", addr)
	return s.httpServer.Start(addr)
}

func (s *Server) recordRequest(success bool, latency time.Duration) {
	atomic.AddInt64(&s.requestsTotal, 1)
	if success {
		atomic.AddInt64(&s.requestsSuccess, 1)
	} else {
		atomic.AddInt64(&s.requestsError, 1)
	}
	s.latencyMu.Lock()
	s.latencySum += latency.Seconds()
	s.latencyMu.Unlock()
}

func (s *Server) recordToolCall(toolName string) {
	s.toolMu.Lock()
	defer s.toolMu.Unlock()
	s.toolCalls[toolName]++
}
