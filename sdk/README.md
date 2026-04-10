# Managing-Up SDK

SDK for building MCP servers that integrate with the Managing-Up platform.

## Languages

- [Go](./mcp) - Full-featured Go SDK with HTTP/Stdio support
- [Python](./python) - Python SDK with async support
- [TypeScript](./typescript) - TypeScript/Node.js SDK
- [Rust](./rust) - Rust SDK

## Features

- **Auto-Registration**: Register MCP server with platform on startup
- **Metrics Collection**: Built-in request/response metrics
- **Tool Tracking**: Track tool usage counts
- **Prometheus Compatible**: Metrics format compatible with Prometheus

## Environment Variables

All SDKs support these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `MANAGING_UP_PLATFORM_URL` | Platform API URL | `http://localhost:8080` |
| `MANAGING_UP_NAME` | MCP server name | (required) |
| `MANAGING_UP_VERSION` | MCP server version | `1.0.0` |
| `MANAGING_UP_DESCRIPTION` | Server description | - |
| `MANAGING_UP_TOKEN` | Auth token for registration | - |
| `MANAGING_UP_TRANSPORT_TYPE` | Transport type (`stdio`, `http`, `sse`) | `http` |

## Quick Start

### Go

```go
import (
    "context"
    sdk "github.com/zealot/managing-up/sdk/mcp"
    "github.com/mark3labs/mcp-go/mcp"
)

server := sdk.NewServer(sdk.Config{
    Name:        "my-mcp",
    Version:     "1.0.0",
    Description: "My MCP server",
})

server.AddTool(mcp.NewTool("hello", mcp.WithDescription("Say hello")), helloHandler)

if _, err := server.Register(ctx); err != nil {
    log.Printf("Warning: registration failed: %v", err)
}

server.StartHTTP(ctx, ":8080")
```

### Python

```python
from managing_up_sdk import MCPServer, MCPServerConfig

config = MCPServerConfig(
    name="my-mcp",
    version="1.0.0",
)
server = MCPServer(config)

server.add_tool("hello", "Say hello", {"type": "object"}, hello_handler)

asyncio.run(server.register())
asyncio.run(server.serve_stdio())
```

### TypeScript

```typescript
import { MCPServer } from '@managing-up/mcp-sdk';

const server = new MCPServer({
    name: 'my-mcp',
    version: '1.0.0',
    transportType: TransportType.HTTP,
});

server.addTool({
    name: 'hello',
    description: 'Say hello',
    inputSchema: { type: 'object' },
    handler: async (args) => `Hello, ${args.name}!`,
});

await server.register();
```

### Rust

```rust
use managing_up_mcp_sdk::{MCPServer, MCPServerConfig, TransportType};

let mut server = MCPServer::new(MCPServerConfig {
    name: "my-mcp".to_string(),
    version: "1.0.0".to_string(),
    transport_type: TransportType::Stdio,
    ..Default::default()
});

server.add_tool("hello", "Say hello", input_schema, |args| async move {
    format!("Hello, {}", args["name"])
});

server.register().await?;
server.serve_stdio().await?;
```
