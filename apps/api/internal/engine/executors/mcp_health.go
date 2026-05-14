package executors

import (
	"context"
	"log/slog"
	"time"
)

// StartHealthChecker starts a background goroutine that periodically checks
// the health of all registered MCP servers. It returns a stop function.
func StartHealthChecker(ctx context.Context, registry *MCPRegistry, interval time.Duration) func() {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				results := registry.HealthCheck(ctx)
				for name, err := range results {
					if err != nil {
						slog.Warn("MCP server health check failed", "server", name, "error", err)
					}
				}
			}
		}
	}()

	return cancel
}
