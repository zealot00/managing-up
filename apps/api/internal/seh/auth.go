package seh

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	Subject  string   `json:"sub,omitempty"`
	Audience []string `json:"aud,omitempty"`
	Issuer   string   `json:"iss,omitempty"`
	Role     string   `json:"role,omitempty"`
}

type AuthConfig struct {
	Secret         string
	Issuer         string
	Audience       string
	SkipValidation bool
}

type contextKey string

const ContextKeyClaims contextKey = "seh_claims"

func GetClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*JWTClaims)
	return claims, ok
}

func validateToken(tokenString string, cfg AuthConfig) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func AuthMiddleware(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/seh/healthz" || r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.SkipValidation {
				ctx := context.WithValue(r.Context(), ContextKeyClaims, &JWTClaims{
					Subject: "test-user",
					Role:    "admin",
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, "Authorization header required", http.StatusUnauthorized, "UNAUTHORIZED")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeError(w, "Invalid authorization format. Use: Bearer <token>", http.StatusUnauthorized, "UNAUTHORIZED")
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				writeError(w, "Token is empty", http.StatusUnauthorized, "UNAUTHORIZED")
				return
			}

			claims, err := validateToken(tokenString, cfg)
			if err != nil {
				writeError(w, "Invalid token", http.StatusUnauthorized, "UNAUTHORIZED")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GenerateToken(cfg AuthConfig, subject string, role string, expiresIn time.Duration) (string, error) {
	now := time.Now()

	claims := &JWTClaims{
		Subject:  subject,
		Role:     role,
		Issuer:   cfg.Issuer,
		Audience: []string{cfg.Audience},
	}
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(expiresIn))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

var staticTokens = map[string]string{
	"mock_token_admin":    "admin",
	"mock_token_reviewer": "reviewer",
	"mock_token_approver": "approver",
}

func ValidateStaticToken(token string) (string, bool) {
	role, exists := staticTokens[token]
	return role, exists
}

func HashAPIKey(apiKey string) string {
	h := sha256.New()
	h.Write([]byte(apiKey))
	return hex.EncodeToString(h.Sum(nil))
}
