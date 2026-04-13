package gateway

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// APIKeyContextKey is the context key for the API key
	APIKeyContextKey contextKey = "api_key"
	// PrincipalContextKey is the context key for the authenticated principal.
	PrincipalContextKey contextKey = "gateway_principal"
)

// publicEndpoints is a map of paths that don't require authentication
var publicEndpoints = map[string]bool{
	"/health":      true,
	"/api/v1/meta": true,
	"/v1/models":   true,
}

// AuthMiddleware returns an HTTP handler that enforces API key authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return AuthMiddlewareWithValidator(nil, next)
}

func AuthMiddlewareWithValidator(validator APIKeyValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("AuthMiddleware: request received", "path", r.URL.Path, "method", r.Method, "content_length", r.ContentLength)
		path := r.URL.Path

		// Check if endpoint is public (no auth required)
		if publicEndpoints[path] {
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key from Authorization header or x-api-key header
		apiKey := extractAPIKey(r)

		if apiKey == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "API key is required")
			return
		}

		// Validate API key when validator is configured.
		ctx := r.Context()
		if validator != nil {
			principal, err := validator.ValidateAPIKey(ctx, apiKey)
			if err != nil || principal == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
				return
			}
			ctx = context.WithValue(ctx, PrincipalContextKey, *principal)
		}

		// Attach API key to request context
		ctx = context.WithValue(ctx, APIKeyContextKey, apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractAPIKey extracts the API key from the request headers
// It checks Authorization: Bearer <key> first, then x-api-key header
func extractAPIKey(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if apiKey := extractBearerToken(authHeader); apiKey != "" {
		return apiKey
	}

	// Fallback to x-api-key header
	apiKeyHeader := r.Header.Get("x-api-key")
	if apiKey := strings.TrimSpace(apiKeyHeader); apiKey != "" {
		return apiKey
	}

	return ""
}

// GetAPIKeyFromContext retrieves the API key from the request context
func GetAPIKeyFromContext(ctx context.Context) string {
	if apiKey, ok := ctx.Value(APIKeyContextKey).(string); ok {
		return apiKey
	}
	return ""
}

func GetPrincipalFromContext(ctx context.Context) *Principal {
	if principal, ok := ctx.Value(PrincipalContextKey).(Principal); ok {
		return &principal
	}
	return nil
}
