package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

type GatewayAPIKey struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Username   string     `json:"username,omitempty"`
	Role       string     `json:"role,omitempty"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	KeyHash    string     `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type GatewayUsageEvent struct {
	ID               string    `json:"id"`
	APIKeyID         string    `json:"api_key_id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username,omitempty"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	Endpoint         string    `json:"endpoint"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

type GatewayUsageAggregate struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username,omitempty"`
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	RequestCount     int64  `json:"request_count"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

type GatewayUserUsageAggregate struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	RequestCount     int64  `json:"request_count"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

func GenerateGatewayAPIKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random key bytes: %w", err)
	}
	return "skhub_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashGatewayAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func GatewayKeyPrefix(raw string) string {
	if len(raw) <= 10 {
		return raw
	}
	return raw[:10]
}
