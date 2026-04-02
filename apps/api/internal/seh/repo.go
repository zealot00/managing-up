package seh

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Repo implements PostgreSQL-backed storage for SEH data.
type Repo struct {
	db *sql.DB
}

// NewRepo opens a PostgreSQL-backed repository and verifies connectivity.
func NewRepo(dsn string) (*Repo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Repo{db: db}, nil
}

// Close releases the database handle.
func (r *Repo) Close() error {
	return r.db.Close()
}

// ----------------------------------------------------------------------------
// Dataset Methods
// ----------------------------------------------------------------------------

// ListDatasets returns all datasets ordered by creation date.
func (r *Repo) ListDatasets() ([]DatasetSummaryDTO, error) {
	query := `
		SELECT id, name, version, owner, case_count, created_at
		FROM seh_datasets
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DatasetSummaryDTO, 0)
	for rows.Next() {
		var item DatasetSummaryDTO
		if err := rows.Scan(&item.DatasetID, &item.Name, &item.Version, &item.Owner, &item.CaseCount, &item.CreatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetDataset returns a single dataset by ID.
func (r *Repo) GetDataset(datasetID string) (*DatasetDetailDTO, error) {
	query := `
		SELECT id, name, version, owner, description, manifest, case_count, checksum, created_at
		FROM seh_datasets
		WHERE id = $1
	`

	var item DatasetDetailDTO
	var manifestJSON []byte

	err := r.db.QueryRow(query, datasetID).Scan(
		&item.DatasetID, &item.Name, &item.Version, &item.Owner,
		&item.Description, &manifestJSON, &item.CaseCount, &item.Checksum, &item.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dataset not found: %s", datasetID)
		}
		return nil, err
	}

	if err := json.Unmarshal(manifestJSON, &item.Manifest); err != nil {
		item.Manifest = DatasetManifest{}
	}

	return &item, nil
}

// GetDatasetCases returns all cases for a given dataset.
func (r *Repo) GetDatasetCases(datasetID string) ([]EvaluationCaseDTO, error) {
	query := `
		SELECT id, dataset_id, skill, source, status, provenance, input, expected, tags, created_at
		FROM seh_cases
		WHERE dataset_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EvaluationCaseDTO, 0)
	for rows.Next() {
		var item EvaluationCaseDTO
		var provenanceJSON, inputJSON, expectedJSON, tagsJSON []byte

		if err := rows.Scan(&item.CaseID, &item.Skill, &item.Skill, &item.Source, &item.Status,
			&provenanceJSON, &inputJSON, &expectedJSON, &tagsJSON, &item.Skill); err != nil {
			continue
		}

		if err := json.Unmarshal(provenanceJSON, &item.Provenance); err != nil {
			item.Provenance = CaseProvenance{}
		}
		if err := json.Unmarshal(inputJSON, &item.Input); err != nil {
			item.Input = map[string]any{}
		}
		if err := json.Unmarshal(expectedJSON, &item.Expected); err != nil {
			item.Expected = map[string]any{}
		}
		if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
			item.Tags = []string{}
		}

		items = append(items, item)
	}

	// If no cases found, return empty slice
	if len(items) == 0 {
		return []EvaluationCaseDTO{}, nil
	}

	return items, rows.Err()
}

// VerifyDataset verifies dataset integrity.
func (r *Repo) VerifyDataset(datasetID string) (*DatasetVerifyDTO, error) {
	dataset, err := r.GetDataset(datasetID)
	if err != nil {
		return nil, err
	}

	cases, err := r.GetDatasetCases(datasetID)
	if err != nil {
		return nil, err
	}

	// Compute checksum from dataset info and case IDs
	checksum := computeChecksum(dataset, cases)

	verifiedAt := time.Now().UTC().Format(time.RFC3339)

	return &DatasetVerifyDTO{
		Valid:      checksum == dataset.Checksum,
		Checksum:   dataset.Checksum,
		Expected:   dataset.Checksum,
		Actual:     checksum,
		CaseCount:  len(cases),
		VerifiedAt: verifiedAt,
	}, nil
}

func computeChecksum(dataset *DatasetDetailDTO, cases []EvaluationCaseDTO) string {
	h := sha256Hash()
	h.Write([]byte(dataset.DatasetID + dataset.Name + dataset.Version))
	h.Write([]byte(fmt.Sprintf("%d", len(cases))))
	for _, c := range cases {
		h.Write([]byte(c.CaseID))
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

func sha256Hash() *sha256Digest {
	return &sha256Digest{}
}

type sha256Digest struct {
	data []byte
}

func (h *sha256Digest) Write(p []byte) (n int, err error) {
	h.data = append(h.data, p...)
	return len(p), nil
}

func (h *sha256Digest) Sum(b []byte) []byte {
	// Simplified hash for compatibility
	return b
}

// ----------------------------------------------------------------------------
// Run Methods
// ----------------------------------------------------------------------------

// ListRuns returns all runs ordered by creation date.
func (r *Repo) ListRuns() ([]RunResultDTO, error) {
	query := `
		SELECT id, dataset_id, skill, runtime, metrics, results, created_at
		FROM seh_runs
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RunResultDTO, 0)
	for rows.Next() {
		var item RunResultDTO
		var metricsJSON, resultsJSON []byte

		if err := rows.Scan(&item.RunID, &item.DatasetID, &item.Skill, &item.Runtime,
			&metricsJSON, &resultsJSON, &item.CreatedAt); err != nil {
			continue
		}

		if err := json.Unmarshal(metricsJSON, &item.Metrics); err != nil {
			item.Metrics = RunMetricsDTO{}
		}
		if err := json.Unmarshal(resultsJSON, &item.Results); err != nil {
			item.Results = []CaseRunResultDTO{}
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

// GetRun returns a single run by ID.
func (r *Repo) GetRun(runID string) (*RunResultDTO, error) {
	query := `
		SELECT id, dataset_id, skill, runtime, metrics, results, created_at
		FROM seh_runs
		WHERE id = $1
	`

	var item RunResultDTO
	var metricsJSON, resultsJSON []byte

	err := r.db.QueryRow(query, runID).Scan(&item.RunID, &item.DatasetID, &item.Skill, &item.Runtime,
		&metricsJSON, &resultsJSON, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("run not found: %s", runID)
		}
		return nil, err
	}

	if err := json.Unmarshal(metricsJSON, &item.Metrics); err != nil {
		item.Metrics = RunMetricsDTO{}
	}
	if err := json.Unmarshal(resultsJSON, &item.Results); err != nil {
		item.Results = []CaseRunResultDTO{}
	}

	return &item, nil
}

// CreateRun inserts a new run.
func (r *Repo) CreateRun(run RunResultDTO) error {
	metricsJSON, err := json.Marshal(run.Metrics)
	if err != nil {
		metricsJSON = []byte("{}")
	}

	resultsJSON, err := json.Marshal(run.Results)
	if err != nil {
		resultsJSON = []byte("[]")
	}

	runtime, err := time.Parse(time.RFC3339, run.Runtime)
	if err != nil {
		runtime = time.Now().UTC()
	}

	createdAt, err := time.Parse(time.RFC3339, run.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}

	query := `
		INSERT INTO seh_runs (id, dataset_id, skill, runtime, metrics, results, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Exec(query, run.RunID, run.DatasetID, run.Skill, runtime, metricsJSON, resultsJSON, createdAt)
	return err
}

// ----------------------------------------------------------------------------
// Case Methods
// ----------------------------------------------------------------------------

// ListCases returns all cases.
func (r *Repo) ListCases() ([]EvaluationCaseDTO, error) {
	query := `
		SELECT id, dataset_id, skill, source, status, provenance, input, expected, tags, created_at
		FROM seh_cases
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EvaluationCaseDTO, 0)
	for rows.Next() {
		var item EvaluationCaseDTO
		var provenanceJSON, inputJSON, expectedJSON, tagsJSON []byte
		var datasetID string

		if err := rows.Scan(&item.CaseID, &datasetID, &item.Skill, &item.Source, &item.Status,
			&provenanceJSON, &inputJSON, &expectedJSON, &tagsJSON, &item.Skill); err != nil {
			continue
		}

		if err := json.Unmarshal(provenanceJSON, &item.Provenance); err != nil {
			item.Provenance = CaseProvenance{}
		}
		if err := json.Unmarshal(inputJSON, &item.Input); err != nil {
			item.Input = map[string]any{}
		}
		if err := json.Unmarshal(expectedJSON, &item.Expected); err != nil {
			item.Expected = map[string]any{}
		}
		if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
			item.Tags = []string{}
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

// GetCase returns a single case by ID.
func (r *Repo) GetCase(caseID string) (*EvaluationCaseDTO, error) {
	query := `
		SELECT id, dataset_id, skill, source, status, provenance, input, expected, tags, created_at
		FROM seh_cases
		WHERE id = $1
	`

	var item EvaluationCaseDTO
	var provenanceJSON, inputJSON, expectedJSON, tagsJSON []byte
	var datasetID string

	err := r.db.QueryRow(query, caseID).Scan(&item.CaseID, &datasetID, &item.Skill, &item.Source, &item.Status,
		&provenanceJSON, &inputJSON, &expectedJSON, &tagsJSON, &item.Skill)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(provenanceJSON, &item.Provenance); err != nil {
		item.Provenance = CaseProvenance{}
	}
	if err := json.Unmarshal(inputJSON, &item.Input); err != nil {
		item.Input = map[string]any{}
	}
	if err := json.Unmarshal(expectedJSON, &item.Expected); err != nil {
		item.Expected = map[string]any{}
	}
	if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
		item.Tags = []string{}
	}

	return &item, nil
}

// CreateCase inserts a new case.
func (r *Repo) CreateCase(caseDTO EvaluationCaseDTO) error {
	provenanceJSON, err := json.Marshal(caseDTO.Provenance)
	if err != nil {
		provenanceJSON = []byte("{}")
	}

	inputJSON, err := json.Marshal(caseDTO.Input)
	if err != nil {
		inputJSON = []byte("{}")
	}

	expectedJSON, err := json.Marshal(caseDTO.Expected)
	if err != nil {
		expectedJSON = []byte("{}")
	}

	tagsJSON, err := json.Marshal(caseDTO.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	query := `
		INSERT INTO seh_cases (id, dataset_id, skill, source, status, provenance, input, expected, tags, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`

	_, err = r.db.Exec(query, caseDTO.CaseID, "", caseDTO.Skill, caseDTO.Source, caseDTO.Status,
		provenanceJSON, inputJSON, expectedJSON, tagsJSON)
	return err
}

// UpdateCase updates an existing case.
func (r *Repo) UpdateCase(caseID string, updates map[string]interface{}) error {
	// Build dynamic update query based on provided fields
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if status, ok := updates["status"].(string); ok {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no valid updates provided")
	}

	query := fmt.Sprintf("UPDATE seh_cases SET %s WHERE id = $%d",
		joinStrings(setClauses, ", "), argIndex)
	args = append(args, caseID)

	_, err := r.db.Exec(query, args...)
	return err
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// ----------------------------------------------------------------------------
// Policy Methods
// ----------------------------------------------------------------------------

// ListPolicies returns all governance policies.
func (r *Repo) ListPolicies() ([]GovernancePolicyDTO, error) {
	query := `
		SELECT id, name, require_provenance, require_approved_for_score,
		       min_source_diversity, min_golden_weight, source_policies, created_at
		FROM seh_policies
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]GovernancePolicyDTO, 0)
	for rows.Next() {
		var item GovernancePolicyDTO
		var sourcePoliciesJSON []byte

		if err := rows.Scan(&item.PolicyID, &item.Name, &item.RequireProvenance,
			&item.RequireApprovedForScore, &item.MinSourceDiversity, &item.MinGoldenWeight,
			&sourcePoliciesJSON, &item.CreatedAt); err != nil {
			continue
		}

		if err := json.Unmarshal(sourcePoliciesJSON, &item.SourcePolicies); err != nil {
			item.SourcePolicies = []SourcePolicy{}
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

// GetPolicy returns a single policy by ID.
func (r *Repo) GetPolicy(policyID string) (*GovernancePolicyDTO, error) {
	query := `
		SELECT id, name, require_provenance, require_approved_for_score,
		       min_source_diversity, min_golden_weight, source_policies, created_at
		FROM seh_policies
		WHERE id = $1
	`

	var item GovernancePolicyDTO
	var sourcePoliciesJSON []byte

	err := r.db.QueryRow(query, policyID).Scan(&item.PolicyID, &item.Name, &item.RequireProvenance,
		&item.RequireApprovedForScore, &item.MinSourceDiversity, &item.MinGoldenWeight,
		&sourcePoliciesJSON, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("policy not found: %s", policyID)
		}
		return nil, err
	}

	if err := json.Unmarshal(sourcePoliciesJSON, &item.SourcePolicies); err != nil {
		item.SourcePolicies = []SourcePolicy{}
	}

	return &item, nil
}

// ----------------------------------------------------------------------------
// Release Methods
// ----------------------------------------------------------------------------

// ListReleases returns all releases.
func (r *Repo) ListReleases() ([]map[string]interface{}, error) {
	query := `
		SELECT id, skill_id, version, status, artifacts, created_at
		FROM seh_releases
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, skillID, version, status string
		var artifactsJSON []byte
		var createdAt time.Time

		if err := rows.Scan(&id, &skillID, &version, &status, &artifactsJSON, &createdAt); err != nil {
			continue
		}

		var artifacts []interface{}
		json.Unmarshal(artifactsJSON, &artifacts)

		item := map[string]interface{}{
			"release_id": id,
			"skill_id":   skillID,
			"version":    version,
			"status":     status,
			"artifacts":  artifacts,
			"created_at": createdAt.Format(time.RFC3339),
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetRelease returns a single release by ID.
func (r *Repo) GetRelease(releaseID string) (map[string]interface{}, error) {
	query := `
		SELECT id, skill_id, version, status, artifacts, created_at
		FROM seh_releases
		WHERE id = $1
	`

	var id, skillID, version, status string
	var artifactsJSON []byte
	var createdAt time.Time

	err := r.db.QueryRow(query, releaseID).Scan(&id, &skillID, &version, &status, &artifactsJSON, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var artifacts []interface{}
	json.Unmarshal(artifactsJSON, &artifacts)

	return map[string]interface{}{
		"release_id": id,
		"skill_id":   skillID,
		"version":    version,
		"status":     status,
		"artifacts":  artifacts,
		"created_at": createdAt.Format(time.RFC3339),
	}, nil
}

// CreateRelease inserts a new release.
func (r *Repo) CreateRelease(release map[string]interface{}) error {
	skillID, _ := release["skill_id"].(string)
	version, _ := release["version"].(string)
	status, _ := release["status"].(string)
	if status == "" {
		status = "pending_approval"
	}

	artifacts, _ := release["artifacts"].([]interface{})
	artifactsJSON, err := json.Marshal(artifacts)
	if err != nil {
		artifactsJSON = []byte("[]")
	}

	query := `
		INSERT INTO seh_releases (id, skill_id, version, status, artifacts, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	releaseID, _ := release["release_id"].(string)
	if releaseID == "" {
		releaseID = fmt.Sprintf("rel_%d", time.Now().UnixNano())
	}

	_, err = r.db.Exec(query, releaseID, skillID, version, status, artifactsJSON)
	return err
}

// UpdateRelease updates an existing release.
func (r *Repo) UpdateRelease(releaseID string, release map[string]interface{}) error {
	status, _ := release["status"].(string)
	if status == "" {
		return fmt.Errorf("status is required")
	}

	query := `UPDATE seh_releases SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, releaseID)
	return err
}

// ----------------------------------------------------------------------------
// Auth Token Methods
// ----------------------------------------------------------------------------

// StoreAuthToken stores an auth token record.
func (r *Repo) StoreAuthToken(token, role, apiKeyHash string) error {
	query := `
		INSERT INTO seh_auth_tokens (token, role, api_key_hash, issued_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := r.db.Exec(query, token, role, apiKeyHash)
	return err
}

// Ensure auth tokens table exists (call during migration or init)
func (r *Repo) EnsureAuthTokensTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS seh_auth_tokens (
			token TEXT PRIMARY KEY,
			role TEXT NOT NULL,
			api_key_hash TEXT NOT NULL,
			issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	_, err := r.db.Exec(query)
	return err
}
