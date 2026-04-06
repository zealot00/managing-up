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

type createProviderKeyRequest struct {
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	MonthlyLimit int    `json:"monthly_limit"`
}

type updateProviderKeyRequest struct {
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	MonthlyLimit int    `json:"monthly_limit"`
	IsEnabled    *bool  `json:"is_enabled"`
}

type updateBudgetRequest struct {
	MonthlyLimit int `json:"monthly_limit"`
	DailyLimit   int `json:"daily_limit"`
}

func (s *Server) handleGatewayProviders(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		items := s.repo.ListGatewayProviderKeys(user.ID)
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	case http.MethodPost:
		var req createProviderKeyRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if strings.TrimSpace(req.Provider) == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "provider is required.")
			return
		}
		if strings.TrimSpace(req.APIKey) == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "api_key is required.")
			return
		}

		now := time.Now().UTC()
		key := GatewayProviderKey{
			ID:           fmt.Sprintf("provkey_%d", now.UnixNano()),
			UserID:       user.ID,
			Provider:     strings.TrimSpace(req.Provider),
			Model:        strings.TrimSpace(req.Model),
			KeyHash:      HashGatewayAPIKey(req.APIKey),
			KeyPrefix:    GatewayKeyPrefix(req.APIKey),
			EncryptedKey: req.APIKey,
			IsEnabled:    true,
			MonthlyLimit: req.MonthlyLimit,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.repo.CreateGatewayProviderKey(key); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create provider key.")
			return
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), map[string]any{"item": key})
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleGatewayProviderByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	keyID := strings.TrimPrefix(r.URL.Path, "/api/v1/gateway/providers/")
	if keyID == "" || strings.Contains(keyID, "/") {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid provider key id.")
		return
	}

	key, found := s.repo.GetGatewayProviderKey(keyID)
	if !found {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Provider key not found.")
		return
	}
	if key.UserID != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Access denied.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"item": key})
	case http.MethodPut:
		var req updateProviderKeyRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		if strings.TrimSpace(req.Provider) == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "provider is required.")
			return
		}

		key.Provider = strings.TrimSpace(req.Provider)
		key.Model = strings.TrimSpace(req.Model)
		key.MonthlyLimit = req.MonthlyLimit
		if req.APIKey != "" {
			key.KeyHash = HashGatewayAPIKey(req.APIKey)
			key.KeyPrefix = GatewayKeyPrefix(req.APIKey)
			key.EncryptedKey = req.APIKey
		}
		if req.IsEnabled != nil {
			key.IsEnabled = *req.IsEnabled
		}

		if err := s.repo.UpdateGatewayProviderKey(key); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update provider key.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"item": key})
	case http.MethodDelete:
		if err := s.repo.DeleteGatewayProviderKey(keyID, user.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete provider key.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]string{"status": "deleted"})
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleGatewayProviderToggle(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	keyID := strings.TrimPrefix(r.URL.Path, "/api/v1/gateway/providers/")
	keyID = strings.TrimSuffix(keyID, "/toggle")
	if keyID == "" || strings.Contains(keyID, "/") {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid provider key id.")
		return
	}

	key, found := s.repo.GetGatewayProviderKey(keyID)
	if !found {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Provider key not found.")
		return
	}
	if key.UserID != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Access denied.")
		return
	}

	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if err := s.repo.ToggleGatewayProviderKey(keyID, user.ID, req.Enabled); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to toggle provider key.")
		return
	}

	key.IsEnabled = req.Enabled
	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"item": key})
}

func (s *Server) handleGatewayBudget(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		budget, found := s.repo.GetUserBudget(user.ID)
		if !found {
			now := time.Now().UTC()
			nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			budget = UserBudget{
				ID:            fmt.Sprintf("budget_%s", user.ID),
				UserID:        user.ID,
				MonthlyLimit:  0,
				DailyLimit:    0,
				UsedThisMonth: 0,
				UsedToday:     0,
				ResetAt:       nextMonth,
				UpdatedAt:     now,
			}
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"item": budget})
	case http.MethodPut:
		var req updateBudgetRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		budget, found := s.repo.GetUserBudget(user.ID)
		if !found {
			now := time.Now().UTC()
			nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			budget = UserBudget{
				ID:        fmt.Sprintf("budget_%s", user.ID),
				UserID:    user.ID,
				ResetAt:   nextMonth,
				UpdatedAt: now,
			}
		}

		budget.MonthlyLimit = req.MonthlyLimit
		budget.DailyLimit = req.DailyLimit

		if err := s.repo.CreateOrUpdateUserBudget(budget); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update budget.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"item": budget})
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}
