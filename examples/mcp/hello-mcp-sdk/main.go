package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	sdk "github.com/zealot/managing-up/sdk/mcp"
)

func main() {
	server := sdk.NewServer(sdk.Config{
		Name:        "hello-mcp-sdk",
		Version:     "1.0.0",
		Description: "Demo MCP server using Managing-Up SDK",
	})

	server.AddTool(mcp.NewTool(
		"hello",
		mcp.WithDescription("Returns a greeting message"),
		mcp.WithString("name",
			mcp.Description("Name of the person to greet"),
			mcp.Required(),
		),
	), helloHandler)

	server.AddTool(mcp.NewTool(
		"echo",
		mcp.WithDescription("Echoes back the input text"),
		mcp.WithString("text",
			mcp.Description("Text to echo back"),
			mcp.Required(),
		),
	), echoHandler)

	server.AddTool(mcp.NewTool(
		"get_time",
		mcp.WithDescription("Returns the current time"),
		mcp.WithString("format",
			mcp.Description("Time format (e.g., '2006-01-02 15:04:05')"),
			mcp.DefaultString("2006-01-02 15:04:05"),
		),
	), getTimeHandler)

	server.AddTool(mcp.NewTool(
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

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			m := server.Metrics()
			json.NewEncoder(w).Encode(m)
		})
		log.Println("Health server on :8081")
		http.ListenAndServe(":8081", mux)
	}()

	ctx := context.Background()

	if _, err := server.Register(ctx); err != nil {
		log.Printf("Warning: registration failed: %v (will continue without registration)", err)
	} else {
		log.Println("Registered with Managing-Up platform")
	}

	log.Println("Starting MCP server on :8080")
	if err := server.StartHTTP(ctx, ":8080"); err != nil {
		log.Fatal(err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := mcp.ParseString(request, "name", "World")
	message := fmt.Sprintf("Hello, %s! Welcome to the MCP SDK demo server.", name)
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
