# Managing-Up MCP SDK

SDK for building MCP servers that integrate with the Managing-Up platform.

## Features

- **Auto-Registration**: Register MCP server with platform on startup
- **Metrics Collection**: Built-in request/response metrics
- **Tool Tracking**: Track tool usage counts
- **Prometheus Compatible**: Metrics format compatible with Prometheus

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	sdk "github.com/zealot/managing-up/sdk/mcp"
)

func main() {
	// Create server - reads config from env vars
	server := sdk.NewServer(sdk.Config{
		Name:        "my-mcp-server",
		Version:     "1.0.0",
		Description: "My MCP server with auto-registration",
	})

	// Add tools
	server.AddTool(mcp.NewTool(
		"hello",
		mcp.WithDescription("Returns a greeting"),
		mcp.WithString("name",
			mcp.Description("Name to greet"),
			mcp.Required(),
		),
	), helloHandler)

	// Register with platform
	ctx := context.Background()
	if err := server.Register(ctx); err != nil {
		log.Printf("Warning: failed to register: %v", err)
	}

	// Start server
	if err := server.StartHTTP(ctx, ":8080"); err != nil {
		log.Fatal(err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := mcp.ParseString(request, "name", "World")
	return mcp.NewToolResultText("Hello, " + name + "!"), nil
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MANAGING_UP_PLATFORM_URL` | Platform API URL | `http://localhost:8080` |
| `MANAGING_UP_NAME` | MCP server name | (required or via config) |
| `MANAGING_UP_VERSION` | MCP server version | `1.0.0` |
| `MANAGING_UP_TOKEN` | Auth token for registration | - |

## Metrics

The SDK tracks:

- `requests_total` - Total request count
- `requests_success` - Successful requests
- `requests_error` - Failed requests  
- `latency_seconds` - Request latency sum
- `tool_calls` - Per-tool call counts

Get metrics via `server.Metrics()`.

## Example with Prometheus

```go
import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

http.Handle("/metrics", promhttp.Handler())
```
