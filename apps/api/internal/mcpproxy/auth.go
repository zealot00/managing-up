package mcpproxy

import (
	"context"
	"strings"
)

type contextKey string

const principalKey contextKey = "principal"

// Principal represents the authenticated identity extracted from an MCP request.
type Principal struct {
	APIKeyID string
	UserID   string
}

// PrincipalFromContext retrieves the Principal from context, or nil if unauthenticated.
func PrincipalFromContext(ctx context.Context) *Principal {
	if v := ctx.Value(principalKey); v != nil {
		if p, ok := v.(*Principal); ok {
			return p
		}
	}
	return nil
}

// APIKeyResolver resolves raw API key strings to database IDs.
type APIKeyResolver interface {
	ResolveAPIKey(rawKey string) (dbKeyID, userID string, ok bool)
}

// extractBearerToken extracts the raw API key from an Authorization header.
func extractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
}
