package tool

import "context"

// Tool is a callable function available to the Agent.
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]any) (any, error)
}
