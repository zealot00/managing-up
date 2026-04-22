package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	svc *OrchestrationService
}

func NewServer(svc *OrchestrationService) *Server {
	return &Server{svc: svc}
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	resp := s.svc.Health()
	writeJSON(w, http.StatusOK, resp)
}

func getIdempotencyKey(r *http.Request) string {
	return r.Header.Get("Idempotency-Key")
}

func (s *Server) HandleOrchestratorRuns(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		idempKey := getIdempotencyKey(r)
		if idempKey != "" && s.svc.idempStore != nil {
			if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
				writeJSON(w, statusCode, cached)
				return
			}
		}
		var req CreateRunRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		resp := s.svc.CreateRun(req)
		w.Header().Set("Location", resp.Links.Self)
		if idempKey != "" && s.svc.idempStore != nil {
			s.svc.idempStore.Set(idempKey, resp, http.StatusAccepted)
		}
		writeJSON(w, http.StatusAccepted, resp)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) HandleOrchestratorRunByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/artifacts") {
		s.HandleRunArtifacts(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		runID := extractRunID(path)
		if runID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "runId is required")
			return
		}
		resp := s.svc.GetRun(runID)
		writeJSON(w, http.StatusOK, resp)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) HandleRunArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	runID := extractRunIDFromArtifacts(r.URL.Path)
	if runID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "runId is required")
		return
	}
	resp := s.svc.ListRunArtifacts(runID)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleExtraction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/enhance"):
		s.HandleEnhanceExtraction(w, r)
	case strings.HasSuffix(path, "/compare"):
		s.HandleCompareExtraction(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint not found")
	}
}

func (s *Server) HandleEnhanceExtraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	idempKey := getIdempotencyKey(r)
	if idempKey != "" && s.svc.idempStore != nil {
		if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
			writeJSON(w, statusCode, cached)
			return
		}
	}
	var req EnhanceExtractionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp := s.svc.EnhanceExtraction(req)
	if idempKey != "" && s.svc.idempStore != nil {
		s.svc.idempStore.Set(idempKey, resp, http.StatusOK)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleCompareExtraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	idempKey := getIdempotencyKey(r)
	if idempKey != "" && s.svc.idempStore != nil {
		if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
			writeJSON(w, statusCode, cached)
			return
		}
	}
	var req CompareExtractionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp := s.svc.CompareExtraction(req)
	if idempKey != "" && s.svc.idempStore != nil {
		s.svc.idempStore.Set(idempKey, resp, http.StatusOK)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleOrchestratorSkills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		skills, err := s.svc.repo.ListSkills()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to list skills: %v", err))
			return
		}
		if skills == nil {
			skills = []Skill{}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items": skills,
			"total": len(skills),
		})
	case http.MethodPost:
		idempKey := getIdempotencyKey(r)
		if idempKey != "" && s.svc.idempStore != nil {
			if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
				writeJSON(w, statusCode, cached)
				return
			}
		}
		var req CreateSkillRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		resp := s.svc.CreateSkill(req)
		if idempKey != "" && s.svc.idempStore != nil {
			s.svc.idempStore.Set(idempKey, resp, http.StatusCreated)
		}
		writeJSON(w, http.StatusCreated, resp)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) HandleOrchestratorSkillByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/versions"):
		s.HandleSkillVersions(w, r)
	case strings.Contains(path, "/versions/"):
		s.HandleSkillVersionByID(w, r)
	case strings.HasSuffix(path, "/diff"):
		s.HandleDiffSkillVersions(w, r)
	case strings.HasSuffix(path, "/rollback"):
		s.HandleRollbackSkill(w, r)
	case strings.HasSuffix(path, "/promote"):
		s.HandlePromoteSkill(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint not found")
	}
}

func (s *Server) HandleSkillVersions(w http.ResponseWriter, r *http.Request) {
	skillID := extractSkillIDFromVersions(r.URL.Path)
	if skillID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skillId is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		resp := s.svc.ListSkillVersions(skillID)
		writeJSON(w, http.StatusOK, resp)
	case http.MethodPost:
		idempKey := getIdempotencyKey(r)
		if idempKey != "" && s.svc.idempStore != nil {
			if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
				writeJSON(w, statusCode, cached)
				return
			}
		}
		var req CreateSkillVersionRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		resp := s.svc.CreateSkillVersion(skillID, req)
		if idempKey != "" && s.svc.idempStore != nil {
			s.svc.idempStore.Set(idempKey, resp, http.StatusCreated)
		}
		writeJSON(w, http.StatusCreated, resp)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) HandleSkillVersionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	path := r.URL.Path
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skillId and version are required")
		return
	}
	skillID := parts[len(parts)-3]
	version := parts[len(parts)-1]
	resp := s.svc.GetSkillVersion(skillID, version)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleDiffSkillVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	skillID := extractSkillIDFromPath(r.URL.Path)
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "from and to query params are required")
		return
	}
	resp := s.svc.DiffSkillVersions(skillID, from, to)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleRollbackSkill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	skillID := extractSkillIDFromPath(r.URL.Path)
	var req RollbackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp := s.svc.Rollback(skillID, req)
	writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) HandlePromoteSkill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	idempKey := getIdempotencyKey(r)
	if idempKey != "" && s.svc.idempStore != nil {
		if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
			writeJSON(w, statusCode, cached)
			return
		}
	}
	skillID := extractSkillIDFromPath(r.URL.Path)
	var req PromoteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp, err := s.svc.Promote(skillID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "PROMOTION_FAILED", err.Error())
		return
	}
	if idempKey != "" && s.svc.idempStore != nil {
		s.svc.idempStore.Set(idempKey, resp, http.StatusAccepted)
	}
	writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) HandleTests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/runs"):
		if r.Method == http.MethodPost {
			s.HandleCreateTestRun(w, r)
		} else {
			writeMethodNotAllowed(w, r.Method)
		}
	case strings.Contains(path, "/runs/") && strings.HasSuffix(path, "/report"):
		s.HandleTestReport(w, r)
	case strings.Contains(path, "/runs/"):
		s.HandleTestRunByID(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint not found")
	}
}

func (s *Server) HandleCreateTestRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	idempKey := getIdempotencyKey(r)
	if idempKey != "" && s.svc.idempStore != nil {
		if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
			writeJSON(w, statusCode, cached)
			return
		}
	}
	var req CreateTestRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp := s.svc.CreateTestRun(req)
	if idempKey != "" && s.svc.idempStore != nil {
		s.svc.idempStore.Set(idempKey, resp, http.StatusAccepted)
	}
	writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) HandleTestRunByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	testRunID := extractTestRunID(r.URL.Path)
	if testRunID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "testRunId is required")
		return
	}
	resp := s.svc.GetTestRun(testRunID)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleTestReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	testRunID := extractTestRunIDFromReport(r.URL.Path)
	if testRunID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "testRunId is required")
		return
	}
	resp := s.svc.GetTestReport(testRunID)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleGates(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/evaluate") {
		s.HandleEvaluateGate(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint not found")
}

func (s *Server) HandleEvaluateGate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	idempKey := getIdempotencyKey(r)
	if idempKey != "" && s.svc.idempStore != nil {
		if cached, statusCode, found := s.svc.idempStore.Get(idempKey); found {
			writeJSON(w, statusCode, cached)
			return
		}
	}
	var req GateEvaluateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	resp := s.svc.EvaluateGate(req)
	if idempKey != "" && s.svc.idempStore != nil {
		s.svc.idempStore.Set(idempKey, resp, http.StatusOK)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandlePolicies(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if !strings.HasPrefix(path, "/v1/policies/") {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint not found")
		return
	}

	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}
	policyID := strings.TrimPrefix(path, "/v1/policies/")
	if policyID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "policyId is required")
		return
	}
	resp := s.svc.GetPolicy(policyID)
	writeJSON(w, http.StatusOK, resp)
}

func extractRunID(path string) string {
	return strings.TrimPrefix(path, "/v1/runs/")
}

func extractRunIDFromArtifacts(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, "/v1/runs/"), "/artifacts")
}

func extractSkillIDFromVersions(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "versions" && i > 0 {
			return parts[i-1]
		}
	}
	return ""
}

func extractSkillIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "skills" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func extractTestRunID(path string) string {
	prefix := "/v1/tests/runs/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	if strings.Contains(id, "/") {
		return strings.Split(id, "/")[0]
	}
	return id
}

func extractTestRunIDFromReport(path string) string {
	prefix := "/v1/tests/runs/"
	suffix := "/report"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	id := strings.TrimSuffix(path, suffix)
	return strings.TrimPrefix(id, prefix)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, method string) {
	writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", fmt.Sprintf("Method %s is not allowed.", method))
}
