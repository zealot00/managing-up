package orchestrator

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in a JWT token.
type JWTClaims struct {
	jwt.RegisteredClaims
	Subject   string   `json:"sub,omitempty"`
	Audience  []string `json:"aud,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	ExpiresAt *jwt.NumericDate
	IssuedAt  *jwt.NumericDate
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Secret         string
	Issuer         string
	Audience       string
	SkipValidation bool // For testing/development
}

// AuthMiddleware creates JWT authentication middleware.
func AuthMiddleware(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required")
				return
			}

			// Parse Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization format. Use: Bearer <token>")
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token is empty")
				return
			}

			// For development/testing, skip validation if configured
			if cfg.SkipValidation {
				// Parse without validation and set a minimal context
				ctx := context.WithValue(r.Context(), ContextKeyClaims, &JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "test-user",
					},
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Parse and validate the token
			claims, err := validateToken(tokenString, cfg)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", fmt.Sprintf("Invalid token: %v", err))
				return
			}

			// Set claims in context
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ContextKeyClaims is the context key for JWT claims.
const ContextKeyClaims contextKey = "jwt_claims"

type contextKey string

// GetClaimsFromContext retrieves JWT claims from context.
func GetClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*JWTClaims)
	return claims, ok
}

// validateToken parses and validates a JWT token.
func validateToken(tokenString string, cfg AuthConfig) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer if configured
	if cfg.Issuer != "" {
		if subtle.ConstantTimeCompare([]byte(claims.Issuer), []byte(cfg.Issuer)) != 1 {
			return nil, fmt.Errorf("invalid issuer")
		}
	}

	// Validate audience if configured
	if cfg.Audience != "" {
		found := false
		for _, aud := range claims.Audience {
			if subtle.ConstantTimeCompare([]byte(aud), []byte(cfg.Audience)) == 1 {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	return claims, nil
}

// isPublicEndpoint checks if the path is a public endpoint that doesn't require auth.
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/v1/healthz",
		"/v1/models",
	}
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	return false
}

// GenerateToken generates a JWT token for testing purposes.
func GenerateToken(cfg AuthConfig, subject string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    cfg.Issuer,
			Audience:  []string{cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
	}
	json.NewEncoder(w).Encode(resp)
}
