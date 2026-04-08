package server

import (
	"net/http"
	"strings"

	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
)

func (srv *Server) handleSkillMarket(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	entries, err := srv.skillEnterpriseSvc.GetSkillMarket(r.Context(), category, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_market", entries)
}

func (srv *Server) handleSkillDependencies(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/skills/"), "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid path")
		return
	}
	id := parts[0]

	deps, err := srv.skillEnterpriseSvc.GetSkillWithDeps(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_deps", deps)
}

func (srv *Server) handleSkillRate(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/skills/"), "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid path")
		return
	}
	id := parts[0]

	var req RateSkillRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}

	user := middleware.UserFromContext(r.Context())
	userID := ""
	if user != nil {
		userID = user.ID
	}

	if err := srv.skillEnterpriseSvc.RateSkill(r.Context(), id, userID, req.Rating, req.Comment); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_rate", map[string]bool{"success": true})
}

func (srv *Server) handleSkillResolveDeps(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SkillID string `json:"skill_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}

	deps, err := srv.skillEnterpriseSvc.ResolveDependencies(r.Context(), req.SkillID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_resolve_deps", deps)
}

func (srv *Server) handleSkillSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")

	entries, err := srv.skillEnterpriseSvc.GetSkillMarket(r.Context(), category, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_search", entries)
}
