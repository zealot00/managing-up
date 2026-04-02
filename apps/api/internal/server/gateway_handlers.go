package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
)

type createGatewayKeyRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleGatewayKeys(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		items := s.repo.ListGatewayAPIKeys(user.ID)
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	case http.MethodPost:
		var req createGatewayKeyRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required.")
			return
		}

		rawKey, err := GenerateGatewayAPIKey()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate gateway API key.")
			return
		}

		now := time.Now().UTC()
		key := GatewayAPIKey{
			ID:        fmt.Sprintf("gk_%d", now.UnixNano()),
			UserID:    user.ID,
			Name:      strings.TrimSpace(req.Name),
			KeyPrefix: GatewayKeyPrefix(rawKey),
			KeyHash:   HashGatewayAPIKey(rawKey),
			CreatedAt: now,
		}
		if err := s.repo.CreateGatewayAPIKey(key); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create gateway API key.")
			return
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), map[string]any{
			"key":      rawKey,
			"key_meta": key,
		})
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleGatewayKeyByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}
	if r.Method != http.MethodDelete {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	keyID := strings.TrimPrefix(r.URL.Path, "/api/v1/gateway/keys/")
	if keyID == "" || strings.Contains(keyID, "/") {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid key id.")
		return
	}

	if err := s.repo.RevokeGatewayAPIKey(keyID, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke gateway API key.")
		return
	}
	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]string{"status": "revoked"})
}

func (s *Server) handleGatewayUsage(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	from, to, err := parseGatewayUsageTimeRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	userID := user.ID
	if requested := strings.TrimSpace(r.URL.Query().Get("user_id")); requested != "" {
		if user.Role != "admin" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Only admin can query other users' usage.")
			return
		}
		userID = requested
	}

	items := s.repo.ListGatewayUsageByUser(userID, from, to)
	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
}

func (s *Server) handleGatewayUsageByUsers(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}
	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role is required.")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	from, to, err := parseGatewayUsageTimeRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	items := s.repo.ListGatewayUsageByUsers(from, to)
	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
}

func parseGatewayUsageTimeRange(r *http.Request) (*time.Time, *time.Time, error) {
	var from *time.Time
	var to *time.Time

	if rawFrom := strings.TrimSpace(r.URL.Query().Get("from")); rawFrom != "" {
		t, err := parseGatewayTime(rawFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid from query param, expected RFC3339 or YYYY-MM-DD")
		}
		from = &t
	}
	if rawTo := strings.TrimSpace(r.URL.Query().Get("to")); rawTo != "" {
		t, err := parseGatewayTime(rawTo)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid to query param, expected RFC3339 or YYYY-MM-DD")
		}
		to = &t
	}
	return from, to, nil
}

func parseGatewayTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("invalid time value")
}
