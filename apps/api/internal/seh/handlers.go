package seh

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	repo       *Repo
	store      *Store
	authConfig AuthConfig
}

func NewServer(authConfig AuthConfig) *Server {
	dsn := os.Getenv("DATABASE_URL")
	var repo *Repo
	var store *Store

	if dsn != "" {
		var err error
		repo, err = NewRepo(dsn)
		if err != nil {
			log.Printf("warning: failed to connect to SEH database: %v, using file storage", err)
			store = NewStore("")
			if err := store.InitMockData(); err != nil {
				log.Printf("warning: failed to init mock data: %v", err)
			}
			repo = nil
		}
	} else {
		store = NewStore("")
		if err := store.InitMockData(); err != nil {
			log.Printf("warning: failed to init mock data: %v", err)
		}
	}
	return &Server{repo: repo, store: store, authConfig: authConfig}
}

func NewServerWithRepo(repo *Repo, authConfig AuthConfig) *Server {
	return &Server{repo: repo, authConfig: authConfig}
}

func NewServerWithStore(store *Store, authConfig AuthConfig) *Server {
	return &Server{store: store, authConfig: authConfig}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/auth/token") {
		s.HandleAuthToken(w, r)
		return
	}

	if r.URL.Path == "/datasets/synthesize" || strings.HasPrefix(r.URL.Path, "/datasets/synthesize/") {
		s.HandleDatasetSynthesize(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/datasets/") && strings.HasSuffix(r.URL.Path, "/lineage") {
		datasetID := strings.TrimPrefix(r.URL.Path, "/datasets/")
		datasetID = strings.TrimSuffix(datasetID, "/lineage")
		if datasetID != "" && !strings.Contains(datasetID, "/") {
			s.HandleDatasetLineage(w, r, datasetID)
			return
		}
	}

	if r.URL.Path == "/datasets" || strings.HasPrefix(r.URL.Path, "/datasets") {
		s.HandleDatasets(w, r)
		return
	}

	if r.URL.Path == "/runs" || strings.HasPrefix(r.URL.Path, "/runs") {
		s.HandleRuns(w, r)
		return
	}

	if r.URL.Path == "/policies" || strings.HasPrefix(r.URL.Path, "/policies") {
		s.HandlePolicies(w, r)
		return
	}

	if r.URL.Path == "/cases" || strings.HasPrefix(r.URL.Path, "/cases") {
		s.HandleCases(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/skills/") && strings.Contains(r.URL.Path, "/releases") {
		s.HandleReleases(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/releases/") {
		s.HandleReleaseByID(w, r)
		return
	}

	if r.URL.Path == "/routing/recommend" || r.URL.Path == "/routing/feedback" {
		s.HandleRouting(w, r)
		return
	}

	if r.URL.Path == "/insights/failure-patterns" {
		s.HandleInsights(w, r)
		return
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) HandleAuthToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
		return
	}

	var req AuthTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON body", http.StatusBadRequest, "BAD_REQUEST")
		return
	}

	if req.APIKey == "" {
		writeError(w, "api_key is required", http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	role, valid := ValidateStaticToken(req.APIKey)
	if !valid {
		writeError(w, "Invalid api_key", http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	token, err := GenerateToken(s.authConfig, req.APIKey, role, 24*time.Hour)
	if err != nil {
		writeError(w, "Failed to generate token", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	issuedAt := time.Now().UTC().Format(time.RFC3339)
	if err := s.repo.StoreAuthToken(token, role, HashAPIKey(req.APIKey)); err != nil {
	}

	resp := AuthTokenResponse{
		Token:    token,
		Role:     role,
		IssuedAt: issuedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleDatasets(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/datasets" && r.Method == http.MethodGet {
		s.handleListDatasets(w, r)
		return
	}

	if path == "/datasets" && r.Method == http.MethodPost {
		s.handleCreateDataset(w, r)
		return
	}

	if r.Method == http.MethodDelete && strings.HasPrefix(path, "/datasets/") {
		datasetID := strings.TrimPrefix(path, "/datasets/")
		if datasetID == "" || strings.Contains(datasetID, "/") {
			writeError(w, "Invalid dataset ID", http.StatusBadRequest, "BAD_REQUEST")
			return
		}

		if s.store != nil {
			if err := s.store.DeleteDataset(datasetID); err != nil {
				writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
		} else if s.repo != nil {
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if strings.HasPrefix(path, "/datasets/") {
		parts := strings.Split(strings.TrimPrefix(path, "/datasets/"), "/")
		datasetID := parts[0]

		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				s.handleGetDataset(w, r, datasetID)
				return
			}
		}

		if len(parts) == 2 {
			switch parts[1] {
			case "cases":
				if r.Method == http.MethodGet {
					s.handleGetDatasetCases(w, r, datasetID)
					return
				}
				if r.Method == http.MethodPost {
					s.handleCreateDatasetCase(w, r, datasetID)
					return
				}
			case "verify":
				if r.Method == http.MethodGet {
					s.handleVerifyDataset(w, r, datasetID)
					return
				}
			}
		}
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	var datasets []DatasetSummaryDTO
	var err error

	if s.store != nil {
		datasets, err = s.store.ListDatasets()
	} else if s.repo != nil {
		datasets, err = s.repo.ListDatasets()
	} else {
		writeError(w, "No storage configured", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	if err != nil {
		writeError(w, "Failed to list datasets", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	total := len(datasets)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	if end < offset {
		offset = 0
		end = 0
	}

	paginated := datasets[offset:end]
	if paginated == nil {
		paginated = []DatasetSummaryDTO{}
	}

	pagination := Pagination{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: end < total,
	}

	writeJSON(w, http.StatusOK, DatasetListResponse{
		Datasets:   paginated,
		Pagination: pagination,
	})
}

func (s *Server) handleGetDataset(w http.ResponseWriter, r *http.Request, datasetID string) {
	var dataset *DatasetDetailDTO
	var err error

	if s.store != nil {
		dataset, err = s.store.GetDataset(datasetID)
	} else if s.repo != nil {
		dataset, err = s.repo.GetDataset(datasetID)
	} else {
		writeError(w, "No storage configured", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	if err != nil {
		writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
		return
	}

	writeJSON(w, http.StatusOK, dataset)
}

func (s *Server) handleCreateDataset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name, Version, Owner, Description string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request", http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if req.Name == "" {
		writeError(w, "name is required", http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		return
	}

	dataset := DatasetDetailDTO{
		Name:        req.Name,
		Version:     req.Version,
		Owner:       req.Owner,
		Description: req.Description,
		CaseCount:   0,
		Checksum:    "",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if s.store != nil {
		created, err := s.store.CreateDataset(dataset)
		if err != nil {
			writeError(w, "Failed to create dataset", http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		dataset = created
	} else if s.repo != nil {
		id := fmt.Sprintf("ds_%d", time.Now().UnixNano())
		dataset.DatasetID = id
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"dataset_id": dataset.DatasetID,
		"created_at": dataset.CreatedAt,
	})
}

func (s *Server) handleCreateDatasetCase(w http.ResponseWriter, r *http.Request, datasetID string) {
	var req EvaluationCaseDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
		return
	}

	if s.store != nil {
		if _, err := s.store.GetDataset(datasetID); err != nil {
			writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
			return
		}
	} else if s.repo != nil {
		if _, err := s.repo.GetDataset(datasetID); err != nil {
			writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
			return
		}
	}

	req.CaseID = "case_" + randomID()
	req.Status = "pending_review"

	if s.store != nil {
		if err := s.store.CreateCase(req); err != nil {
			writeError(w, "Failed to create case", http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
	} else if s.repo != nil {
		if err := s.repo.CreateCase(req); err != nil {
			writeError(w, "Failed to create case", http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"case_id": req.CaseID, "status": req.Status, "created_at": time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) handleUpdateCase(w http.ResponseWriter, r *http.Request, caseID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
		return
	}

	if err := s.repo.UpdateCase(caseID, updates); err != nil {
		writeError(w, "Failed to update case", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	c, err := s.repo.GetCase(caseID)
	if err != nil {
		writeError(w, "Case not found", http.StatusNotFound, "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleUpdateRelease(w http.ResponseWriter, r *http.Request, releaseID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
		return
	}

	var release map[string]interface{}
	var err error

	if s.store != nil {
		if err = s.store.UpdateRelease(releaseID, updates); err != nil {
			writeError(w, "Failed to update release", http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		release, err = s.store.GetRelease(releaseID)
	} else if s.repo != nil {
		if err = s.repo.UpdateRelease(releaseID, updates); err != nil {
			writeError(w, "Failed to update release", http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		release, err = s.repo.GetRelease(releaseID)
	} else {
		writeError(w, "No storage configured", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	if err != nil {
		writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, release)
}

func (s *Server) handleGetDatasetCases(w http.ResponseWriter, r *http.Request, datasetID string) {
	limit, offset := parsePagination(r)

	cases, err := s.repo.GetDatasetCases(datasetID)
	if err != nil {
		writeError(w, "Failed to get cases", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	dataset, err := s.repo.GetDataset(datasetID)
	if err != nil {
		writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
		return
	}

	total := len(cases)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	if end < offset {
		offset = 0
		end = 0
	}

	paginated := cases[offset:end]
	if paginated == nil {
		paginated = []EvaluationCaseDTO{}
	}

	pagination := Pagination{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: end < total,
	}

	writeJSON(w, http.StatusOK, DatasetCasesResponse{
		Manifest:   dataset.Manifest,
		Cases:      paginated,
		Pagination: pagination,
	})
}

func (s *Server) handleVerifyDataset(w http.ResponseWriter, r *http.Request, datasetID string) {
	result, err := s.repo.VerifyDataset(datasetID)
	if err != nil {
		writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) HandleRuns(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/runs" {
		switch r.Method {
		case http.MethodGet:
			s.handleListRuns(w, r)
			return
		case http.MethodPost:
			s.handleCreateRun(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/runs/") {
		parts := strings.Split(strings.TrimPrefix(path, "/runs/"), "/")
		runID := parts[0]

		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				s.handleGetRun(w, r, runID)
				return
			}
		}

		if len(parts) == 2 && parts[1] == "gate" {
			if r.Method == http.MethodPost {
				s.handleGateRun(w, r, runID)
				return
			}
		}
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	runs, err := s.repo.ListRuns()
	if err != nil {
		writeError(w, "Failed to list runs", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	total := len(runs)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	if end < offset {
		offset = 0
		end = 0
	}

	paginated := runs[offset:end]
	if paginated == nil {
		paginated = []RunResultDTO{}
	}

	pagination := Pagination{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: end < total,
	}

	writeJSON(w, http.StatusOK, RunListResponse{
		Runs:       paginated,
		Pagination: pagination,
	})
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := s.repo.GetRun(runID)
	if err != nil {
		writeError(w, "Run not found", http.StatusNotFound, "NOT_FOUND")
		return
	}

	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req RunCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON body", http.StatusBadRequest, "BAD_REQUEST")
		return
	}

	if req.DatasetID == "" {
		writeError(w, "dataset_id is required", http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		return
	}

	if _, err := s.repo.GetDataset(req.DatasetID); err != nil {
		writeError(w, "Dataset not found", http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		return
	}

	// For async runs, results can be empty initially
	// Client can poll GET /runs/{run_id} to get results later
	isAsync := len(req.Results) == 0

	dataset, _ := s.repo.GetDataset(req.DatasetID)
	createdAt := time.Now().UTC().Format(time.RFC3339)

	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	run := RunResultDTO{
		RunID:     runID,
		DatasetID: req.DatasetID,
		Skill:     req.Skill,
		Runtime:   createdAt,
		Metrics:   req.Metrics,
		Results:   req.Results,
		CreatedAt: createdAt,
	}

	if dataset != nil {
	}

	if err := s.repo.CreateRun(run); err != nil {
		writeError(w, "Failed to create run", http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	resp := RunCreateResponse{
		RunID:       runID,
		Score:       req.Metrics.Score,
		SuccessRate: req.Metrics.SuccessRate,
		CreatedAt:   createdAt,
	}

	if isAsync {
		writeJSON(w, http.StatusAccepted, resp)
	} else {
		writeJSON(w, http.StatusCreated, resp)
	}
}

func (s *Server) handleGateRun(w http.ResponseWriter, r *http.Request, runID string) {
	var req GateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON body", http.StatusBadRequest, "BAD_REQUEST")
		return
	}

	if req.PolicyID == "" {
		writeError(w, "policy_id is required", http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	run, err := s.repo.GetRun(runID)
	if err != nil {
		writeError(w, "Run not found", http.StatusNotFound, "NOT_FOUND")
		return
	}

	policy, err := s.repo.GetPolicy(req.PolicyID)
	if err != nil {
		writeError(w, "Policy not found", http.StatusNotFound, "NOT_FOUND")
		return
	}

	details := evaluateGate(run, policy)
	passed := true
	for _, d := range details {
		if !d.Passed {
			passed = false
			break
		}
	}

	evaluatedAt := time.Now().UTC().Format(time.RFC3339)

	resp := GateResponse{
		Passed:      passed,
		RunID:       runID,
		PolicyID:    req.PolicyID,
		Details:     details,
		EvaluatedAt: evaluatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func evaluateGate(run *RunResultDTO, policy *GovernancePolicyDTO) []GateDetail {
	var details []GateDetail

	if policy.MinGoldenWeight > 0 {
		goldenWeight := calculateGoldenWeight(run)
		details = append(details, GateDetail{
			Rule:     "min_golden_weight",
			Passed:   goldenWeight >= policy.MinGoldenWeight,
			Required: policy.MinGoldenWeight,
			Actual:   goldenWeight,
		})
	}

	if policy.MinSourceDiversity > 0 {
		diversity := calculateSourceDiversity(run)
		details = append(details, GateDetail{
			Rule:     "min_source_diversity",
			Passed:   diversity >= float64(policy.MinSourceDiversity),
			Required: float64(policy.MinSourceDiversity),
			Actual:   diversity,
		})
	}

	if len(policy.SourcePolicies) > 0 {
		for _, sp := range policy.SourcePolicies {
			actualRate := calculateSourceSuccessRate(run, sp.Source)
			details = append(details, GateDetail{
				Rule:     fmt.Sprintf("source_%s_min_success_rate", sp.Source),
				Passed:   actualRate >= sp.MinSuccessRate,
				Required: sp.MinSuccessRate,
				Actual:   actualRate,
			})
		}
	}

	if policy.RequireProvenance {
		hasProvenance := checkProvenance(run)
		details = append(details, GateDetail{
			Rule:     "require_provenance",
			Passed:   hasProvenance,
			Required: 1,
			Actual:   boolToFloat(hasProvenance),
		})
	}

	successRatePassed := run.Metrics.SuccessRate >= 0.8
	details = append(details, GateDetail{
		Rule:     "min_success_rate",
		Passed:   successRatePassed,
		Required: 0.8,
		Actual:   run.Metrics.SuccessRate,
	})

	scorePassed := run.Metrics.Score >= 0.75
	details = append(details, GateDetail{
		Rule:     "min_score",
		Passed:   scorePassed,
		Required: 0.75,
		Actual:   run.Metrics.Score,
	})

	return details
}

func calculateGoldenWeight(run *RunResultDTO) float64 {
	if len(run.Results) == 0 {
		return 0
	}
	goldenCount := 0
	for _, r := range run.Results {
		if r.Provenance.Source == "golden" {
			goldenCount++
		}
	}
	return float64(goldenCount) / float64(len(run.Results))
}

func calculateSourceDiversity(run *RunResultDTO) float64 {
	sources := make(map[string]bool)
	for _, r := range run.Results {
		if r.Provenance.Source != "" {
			sources[r.Provenance.Source] = true
		}
	}
	return float64(len(sources))
}

func calculateSourceSuccessRate(run *RunResultDTO, source string) float64 {
	var total, success int
	for _, r := range run.Results {
		if r.Provenance.Source == source {
			total++
			if r.Success {
				success++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total)
}

func checkProvenance(run *RunResultDTO) bool {
	for _, r := range run.Results {
		if r.Provenance.Source == "" {
			return false
		}
	}
	return len(run.Results) > 0
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":{"code":500,"message":"%s"}}`, err.Error()), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, message string, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    status,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func readBody(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(data)
}

func (s *Server) HandlePolicies(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method == http.MethodPost && path == "/policies" {
		var policy GovernancePolicyDTO
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		if policy.Name == "" {
			writeError(w, "name is required", http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			return
		}

		if s.store != nil {
			created, err := s.store.CreatePolicy(policy)
			if err != nil {
				writeError(w, "Failed to create policy", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"policy_id":  created.PolicyID,
				"created_at": created.CreatedAt,
			})
		} else if s.repo != nil {
			policy.PolicyID = "pol_" + randomID()
			policy.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"policy_id":  policy.PolicyID,
				"created_at": policy.CreatedAt,
			})
		} else {
			writeError(w, "No storage configured", http.StatusInternalServerError, "INTERNAL_ERROR")
		}
		return
	}

	if r.Method == http.MethodGet && path == "/policies" {
		policies, _ := s.repo.ListPolicies()
		writeJSON(w, http.StatusOK, policies)
		return
	}

	if strings.HasPrefix(path, "/policies/") && r.Method == http.MethodGet {
		policyID := strings.TrimPrefix(path, "/policies/")
		policy, err := s.repo.GetPolicy(policyID)
		if err != nil || policy == nil {
			writeError(w, "Policy not found", http.StatusNotFound, "NOT_FOUND")
			return
		}
		writeJSON(w, http.StatusOK, policy)
		return
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) HandleCases(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method == http.MethodPost && path == "/cases" {
		var req EvaluationCaseDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		req.CaseID = "case_" + randomID()
		req.Status = "pending_review"
		if err := s.repo.CreateCase(req); err != nil {
			writeError(w, "Failed to create case", http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"case_id": req.CaseID, "status": req.Status, "created_at": time.Now().UTC().Format(time.RFC3339)})
		return
	}

	if r.Method == http.MethodGet && path == "/cases" {
		cases, err := s.repo.ListCases()
		if err != nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, cases)
		return
	}

	if strings.HasPrefix(path, "/cases/") {
		caseID := strings.TrimPrefix(path, "/cases/")
		if r.Method == http.MethodPatch && !strings.Contains(caseID, "/") {
			s.handleUpdateCase(w, r, caseID)
			return
		}
		if strings.HasSuffix(caseID, "/review") {
			caseID = strings.TrimSuffix(caseID, "/review")
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			approved, _ := req["approved"].(bool)
			status := "rejected"
			if approved {
				status = "approved"
			}
			updates := map[string]interface{}{"status": status}
			if err := s.repo.UpdateCase(caseID, updates); err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"case_id": caseID, "status": status})
			return
		}
		if strings.HasSuffix(caseID, "/deprecate") {
			caseID = strings.TrimSuffix(caseID, "/deprecate")
			updates := map[string]interface{}{"status": "deprecated"}
			if err := s.repo.UpdateCase(caseID, updates); err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"case_id": caseID, "status": "deprecated"})
			return
		}
		if strings.HasSuffix(caseID, "/lineage") {
			caseID = strings.TrimSuffix(caseID, "/lineage")

			var lineage map[string]interface{}
			if s.store != nil {
				lineage, _ = s.store.GetCaseLineage(caseID)
			} else {
				lineage = map[string]interface{}{
					"ancestors":   []interface{}{},
					"descendants": []interface{}{},
				}
			}
			writeJSON(w, http.StatusOK, lineage)
			return
		}
		// GET single case
		c, err := s.repo.GetCase(caseID)
		if err != nil || c == nil {
			writeError(w, "Case not found", http.StatusNotFound, "NOT_FOUND")
			return
		}
		writeJSON(w, http.StatusOK, c)
		return
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) HandleReleases(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		releaseID := "rel_" + randomID()
		req["release_id"] = releaseID

		if s.store != nil {
			if err := s.store.CreateRelease(req); err != nil {
				writeError(w, "Failed to create release", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
		} else if s.repo != nil {
			if err := s.repo.CreateRelease(req); err != nil {
				writeError(w, "Failed to create release", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"release_id": releaseID, "status": "pending_approval", "created_at": time.Now().UTC().Format(time.RFC3339)})
		return
	}
	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) HandleReleaseByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	releaseID := strings.TrimPrefix(path, "/releases/")

	if r.Method == http.MethodPatch && !strings.Contains(releaseID, "/") {
		s.handleUpdateRelease(w, r, releaseID)
		return
	}

	if strings.HasSuffix(releaseID, "/approve") {
		releaseID = strings.TrimSuffix(releaseID, "/approve")

		var req struct {
			ApprovedBy string `json:"approved_by"`
		}
		if r.ContentLength > 0 {
			json.NewDecoder(r.Body).Decode(&req)
		}
		if req.ApprovedBy == "" {
			claims, _ := GetClaimsFromContext(r.Context())
			if claims != nil {
				req.ApprovedBy = claims.Subject
			} else {
				req.ApprovedBy = "system"
			}
		}

		if s.store != nil {
			if err := s.store.ApproveRelease(releaseID, req.ApprovedBy); err != nil {
				writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
		} else if s.repo != nil {
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release_id":  releaseID,
			"status":      "approved",
			"approved_by": req.ApprovedBy,
			"approved_at": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if strings.HasSuffix(releaseID, "/reject") {
		releaseID = strings.TrimSuffix(releaseID, "/reject")

		var req struct {
			Reason string `json:"reason"`
		}
		if r.ContentLength > 0 {
			json.NewDecoder(r.Body).Decode(&req)
		}

		if s.store != nil {
			if err := s.store.RejectRelease(releaseID, req.Reason); err != nil {
				writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release_id":      releaseID,
			"status":          "rejected",
			"rejected_reason": req.Reason,
		})
		return
	}

	if strings.HasSuffix(releaseID, "/rollback") {
		releaseID = strings.TrimSuffix(releaseID, "/rollback")

		if s.store != nil {
			release, err := s.store.GetReleaseForUpdate(releaseID)
			if err != nil || release == nil {
				writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
			if release["status"] != "approved" {
				writeError(w, "Can only rollback approved releases", http.StatusUnprocessableEntity, "INVALID_STATE")
				return
			}
			if err := s.store.RollbackRelease(releaseID); err != nil {
				writeError(w, "Failed to rollback", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release_id": releaseID,
			"status":     "rolled_back",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"release_id": releaseID, "status": "pending_approval"})
}

func (s *Server) HandleRouting(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/routing/recommend" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"skill": "default",
			"recommendations": []map[string]interface{}{
				{"model_id": "gpt-4o", "score": 0.95, "confidence": 0.9, "avg_latency_ms": 1500, "avg_cost": 0.01},
			},
			"generated_at": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if path == "/routing/feedback" && r.Method == http.MethodPost {
		writeJSON(w, http.StatusOK, map[string]interface{}{"acknowledged": true})
		return
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) HandleInsights(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"clusters":    []map[string]interface{}{},
		"suggestions": []string{},
	})
}

func (s *Server) HandleDatasetSynthesize(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/datasets/synthesize" && r.Method == http.MethodPost {
		jobID := "job_" + randomID()
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"job_id":     jobID,
			"status":     "queued",
			"created_at": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if strings.HasPrefix(path, "/datasets/synthesize/") && r.Method == http.MethodGet {
		jobID := strings.TrimPrefix(path, "/datasets/synthesize/")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"job_id":       jobID,
			"status":       "completed",
			"progress":     map[string]interface{}{"current": 10, "total": 10},
			"created_at":   time.Now().UTC().Format(time.RFC3339),
			"completed_at": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	writeError(w, "Not found", http.StatusNotFound, "NOT_FOUND")
}

func (s *Server) HandleDatasetLineage(w http.ResponseWriter, r *http.Request, datasetID string) {
	var lineage map[string]interface{}
	if s.store != nil {
		lineage, _ = s.store.GetDatasetLineage(datasetID)
	} else {
		lineage = map[string]interface{}{
			"versions": []interface{}{},
		}
	}
	writeJSON(w, http.StatusOK, lineage)
}

func randomID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())[:12]
}

var _ http.Handler = (*Server)(nil)
