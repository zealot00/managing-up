package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
	authMW  *middleware.AuthMiddleware
}

func NewAuthHandler(authSvc *service.AuthService, authMW *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		authMW:  authMW,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", fmt.Sprintf("Content-Type must be application/json, got: [%s]", contentType))
		return
	}

	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "username is required.")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "password is required.")
		return
	}

	result, err := h.authSvc.Login(service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password.")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Login failed.")
		return
	}

	token, err := h.authMW.GenerateToken(toMiddlewareUser(result.User))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token.")
		return
	}

	h.authMW.SetAuthCookie(w, token)

	writeEnvelope(w, http.StatusOK, "req_default", map[string]interface{}{
		"user": toServerUser(result.User),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	h.authMW.ClearAuthCookie(w)

	writeEnvelope(w, http.StatusOK, "req_default", map[string]string{
		"status": "logged_out",
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	writeEnvelope(w, http.StatusOK, "req_default", map[string]interface{}{
		"user": user,
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Envelope struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
	Meta  Meta      `json:"meta"`
}

type Meta struct {
	RequestID  string      `json:"request_id"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func toServerUser(user models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func toMiddlewareUser(user models.User) middleware.User {
	return middleware.User{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func isJSONRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return len(contentType) > 0 && (contentType == "application/json" || strings.HasPrefix(contentType, "application/json;"))
}

func writeEnvelope(w http.ResponseWriter, status int, requestID string, payload any) {
	writeJSON(w, status, Envelope{
		Data: payload,
		Meta: Meta{
			RequestID: requestID,
		},
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Envelope{
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Meta: Meta{
			RequestID: "req_default",
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, method string) {
	writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method "+method+" is not allowed.")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
