package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zealot/managing-up/apps/api/internal/models"
	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo defines the repository interface needed by the UserHandler.
type UserRepo interface {
	GetUserByID(id string) (models.User, bool)
	UpdateUserPassword(userID string, passwordHash string) error
	GetUserPreferences(userID string) (models.UserPreferences, bool)
	UpdateUserPreferences(userID string, req models.UpdatePreferencesRequest) (models.UserPreferences, error)
}

// UserHandler handles user profile and preferences operations.
type UserHandler struct {
	repo UserRepo
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(repo UserRepo) *UserHandler {
	return &UserHandler{repo: repo}
}

// GetProfile handles GET /api/v1/user/profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	fullUser, ok := h.repo.GetUserByID(user.ID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found.")
		return
	}

	writeEnvelope(w, http.StatusOK, "req_default", map[string]interface{}{
		"id":         fullUser.ID,
		"username":   fullUser.Username,
		"role":       fullUser.Role,
		"created_at": fullUser.CreatedAt,
	})
}

// ChangePassword handles PUT /api/v1/user/password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	var req models.ChangePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.CurrentPassword == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "current_password is required.")
		return
	}
	if req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "new_password is required.")
		return
	}
	if len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "new_password must be at least 6 characters.")
		return
	}

	fullUser, ok := h.repo.GetUserByID(user.ID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(fullUser.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", "Current password is incorrect.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to hash password.")
		return
	}

	if err := h.repo.UpdateUserPassword(user.ID, string(hash)); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update password.")
		return
	}

	writeEnvelope(w, http.StatusOK, "req_default", map[string]string{
		"status": "password_updated",
	})
}

// GetPreferences handles GET /api/v1/user/preferences
func (h *UserHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	prefs, _ := h.repo.GetUserPreferences(user.ID)

	writeEnvelope(w, http.StatusOK, "req_default", prefs)
}

// UpdatePreferences handles PUT /api/v1/user/preferences
func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	var req models.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body.")
		return
	}

	if req.Language != nil {
		validLangs := map[string]bool{"en": true, "zh": true}
		if !validLangs[*req.Language] {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "language must be 'en' or 'zh'.")
			return
		}
	}

	prefs, err := h.repo.UpdateUserPreferences(user.ID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update preferences.")
		return
	}

	writeEnvelope(w, http.StatusOK, "req_default", prefs)
}

// HandlePreferences dispatches GET/PUT for /api/v1/user/preferences
func (h *UserHandler) HandlePreferences(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetPreferences(w, r)
	case http.MethodPut:
		h.UpdatePreferences(w, r)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

// --- Adapter for server.Repository ---

// UserHandlerRepoAdapter adapts the server.Repository to the UserRepo interface.
type UserHandlerRepoAdapter struct {
	GetUserByIDFn           func(id string) (models.User, bool)
	UpdateUserPasswordFn    func(userID string, passwordHash string) error
	GetUserPreferencesFn    func(userID string) (models.UserPreferences, bool)
	UpdateUserPreferencesFn func(userID string, req models.UpdatePreferencesRequest) (models.UserPreferences, error)
}

func (a *UserHandlerRepoAdapter) GetUserByID(id string) (models.User, bool) {
	return a.GetUserByIDFn(id)
}

func (a *UserHandlerRepoAdapter) UpdateUserPassword(userID string, passwordHash string) error {
	return a.UpdateUserPasswordFn(userID, passwordHash)
}

func (a *UserHandlerRepoAdapter) GetUserPreferences(userID string) (models.UserPreferences, bool) {
	return a.GetUserPreferencesFn(userID)
}

func (a *UserHandlerRepoAdapter) UpdateUserPreferences(userID string, req models.UpdatePreferencesRequest) (models.UserPreferences, error) {
	return a.UpdateUserPreferencesFn(userID, req)
}
