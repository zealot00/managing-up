package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	httpMode = flag.Bool("http", false, "Serve over HTTP instead of stdio")
	httpPort = flag.String("port", "8080", "HTTP server port")
)

func main() {
	flag.Parse()

	s := server.NewMCPServer(
		"Hello MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool(
		"hello",
		mcp.WithDescription("Returns a greeting message"),
		mcp.WithString("name",
			mcp.Description("Name of the person to greet"),
			mcp.Required(),
		),
	), helloHandler)

	s.AddTool(mcp.NewTool(
		"echo",
		mcp.WithDescription("Echoes back the input text"),
		mcp.WithString("text",
			mcp.Description("Text to echo back"),
			mcp.Required(),
		),
	), echoHandler)

	s.AddTool(mcp.NewTool(
		"get_time",
		mcp.WithDescription("Returns the current time"),
		mcp.WithString("format",
			mcp.Description("Time format (e.g., '2006-01-02 15:04:05')"),
			mcp.DefaultString("2006-01-02 15:04:05"),
		),
	), getTimeHandler)

	s.AddTool(mcp.NewTool(
		"calculator",
		mcp.WithDescription("Performs basic arithmetic operations"),
		mcp.WithNumber("a",
			mcp.Description("First number"),
			mcp.Required(),
		),
		mcp.WithNumber("b",
			mcp.Description("Second number"),
			mcp.Required(),
		),
		mcp.WithString("operation",
			mcp.Description("Operation: add, subtract, multiply, divide"),
			mcp.Enum("add", "subtract", "multiply", "divide"),
			mcp.DefaultString("add"),
		),
	), calculatorHandler)

	if *httpMode {
		serveHTTP(s)
	} else {
		serveStdio(s)
	}
}

func serveStdio(s *server.MCPServer) {
	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}

func serveHTTP(s *server.MCPServer) {
	httpServer := server.NewStreamableHTTPServer(s)
	addr := ":" + *httpPort

	log.Printf("Starting HTTP MCP server on %s", addr)
	log.Printf("MCP endpoint: http://localhost%s/mcp", addr)

	if err := httpServer.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := mcp.ParseString(request, "name", "World")
	message := fmt.Sprintf("Hello, %s! Welcome to the MCP demo server.", name)
	return mcp.NewToolResultText(message), nil
}

func echoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := mcp.ParseString(request, "text", "")
	return mcp.NewToolResultText(text), nil
}

func getTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := mcp.ParseString(request, "format", "2006-01-02 15:04:05")

	now := time.Now()
	formatted := now.Format(format)

	output := map[string]any{
		"current_time":   formatted,
		"unix_timestamp": now.Unix(),
		"iso8601":        now.Format(time.RFC3339),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func calculatorHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a := mcp.ParseFloat64(request, "a", 0)
	b := mcp.ParseFloat64(request, "b", 0)
	operation := mcp.ParseString(request, "operation", "add")

	var result float64
	var opSymbol string

	switch operation {
	case "add", "addition", "+":
		result = a + b
		opSymbol = "+"
	case "subtract", "subtraction", "-":
		result = a - b
		opSymbol = "-"
	case "multiply", "multiplication", "*", "x":
		result = a * b
		opSymbol = "*"
	case "divide", "division", "/":
		if b == 0 {
			return mcp.NewToolResultError("Cannot divide by zero"), nil
		}
		result = a / b
		opSymbol = "/"
	default:
		return mcp.NewToolResultErrorf("Unknown operation: %s", operation), nil
	}

	output := map[string]any{
		"a":          a,
		"b":          b,
		"operation":  operation,
		"result":     result,
		"expression": fmt.Sprintf("%.2f %s %.2f = %.2f", a, opSymbol, b, result),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
