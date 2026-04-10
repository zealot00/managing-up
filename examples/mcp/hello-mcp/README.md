# Hello MCP Server

A simple Model Context Protocol (MCP) server example for testing and demonstration.

## Features

- `hello` - Returns a greeting message
- `echo` - Echoes back the input text
- `get_time` - Returns the current time
- `calculator` - Performs basic arithmetic operations

## Running

### Stdio Mode (Local)

```bash
go build -o hello-mcp .
./hello-mcp
```

### HTTP Mode (Remote)

```bash
go build -o hello-mcp .
./hello-mcp --http --port 8080
```

The MCP endpoint will be available at `http://localhost:8080/mcp`

## Testing with MCP Inspector

```bash
npx @modelcontextprotocol/inspector http://localhost:8080/mcp
```

## Testing via Curl (Stdio Mode)

```bash
# List tools
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./hello-mcp

# Call hello tool
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-11-25","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"hello","arguments":{"name":"World"}},"id":2}' | ./hello-mcp

# Call calculator tool
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"calculator","arguments":{"a":10,"b":5,"operation":"add"}},"id":3}' | ./hello-mcp
```

## Using with Gateway

### Stdio Mode (Local)

Register with transport_type `stdio`:

```json
{
  "name": "hello-mcp",
  "description": "Demo MCP server with basic tools",
  "transport_type": "stdio",
  "command": "/path/to/hello-mcp"
}
```

### HTTP Mode (Remote)

Register with transport_type `http`:

```json
{
  "name": "hello-mcp",
  "description": "Demo MCP server via HTTP",
  "transport_type": "http",
  "url": "http://mcp-server.example.com:8080/mcp"
}
```

## Development

```bash
go build -o hello-mcp .
go run main.go --http --port 9090
```
