package seh

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	MockDataDir = ".mock-seh"

	DatasetsFile   = "datasets.json"
	RunsFile       = "runs.json"
	CasesFile      = "cases.json"
	PoliciesFile   = "policies.json"
	AuthTokensFile = "auth_tokens.json"
)

type Store struct {
	basePath string
	mu       sync.RWMutex
}

func NewStore(basePath string) *Store {
	if basePath == "" {
		basePath = MockDataDir
	}
	return &Store{basePath: basePath}
}

func (s *Store) ensureDir() error {
	return os.MkdirAll(s.basePath, 0755)
}

func (s *Store) readJSON(filename string, v any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fpath := filepath.Join(s.basePath, filename)
	data, err := os.ReadFile(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", fpath)
		}
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *Store) writeJSON(filename string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureDir(); err != nil {
		return err
	}

	fpath := filepath.Join(s.basePath, filename)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	tmp := fpath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, fpath)
}

func (s *Store) fileExists(filename string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fpath := filepath.Join(s.basePath, filename)
	_, err := os.Stat(fpath)
	return err == nil
}

func (s *Store) InitMockData() error {
	if s.fileExists(DatasetsFile) && s.fileExists(RunsFile) {
		return nil
	}

	datasets := []DatasetDetailDTO{
		{
			DatasetID:   "ds_abc12345",
			Name:        "Code Generation Benchmark",
			Version:     "v1.0",
			Owner:       "engineering",
			Description: "High-quality code generation test cases",
			Manifest: DatasetManifest{
				DatasetName: "Code Generation Benchmark",
				Version:     "v1.0",
				Owner:       "engineering",
				Description: "High-quality code generation test cases",
			},
			CaseCount: 5,
			Checksum:  "sha256:a1b2c3d4e5f6",
			CreatedAt: "2026-03-15T10:00:00Z",
		},
		{
			DatasetID:   "ds_def67890",
			Name:        "Text Summarization Test",
			Version:     "v1.0",
			Owner:       "ml-team",
			Description: "Text summarization evaluation cases",
			Manifest: DatasetManifest{
				DatasetName: "Text Summarization Test",
				Version:     "v1.0",
				Owner:       "ml-team",
				Description: "Text summarization evaluation cases",
			},
			CaseCount: 5,
			Checksum:  "sha256:f6e5d4c3b2a1",
			CreatedAt: "2026-03-20T14:30:00Z",
		},
	}

	cases := []EvaluationCaseDTO{
		{CaseID: "case_001", Skill: "code_generation", Source: "golden", Status: "approved", Provenance: CaseProvenance{ApprovedBy: "admin", ContributorID: "user1", Method: "manual"}, Input: map[string]any{"prompt": "Write a hello world"}, Expected: map[string]any{"output": "hello world"}, Tags: []string{"easy"}},
		{CaseID: "case_002", Skill: "code_generation", Source: "golden", Status: "approved", Provenance: CaseProvenance{ApprovedBy: "admin", ContributorID: "user1", Method: "manual"}, Input: map[string]any{"prompt": "Fibonacci"}, Expected: map[string]any{"output": "fibonacci"}, Tags: []string{"medium"}},
		{CaseID: "case_003", Skill: "code_generation", Source: "adversarial", Status: "approved", Provenance: CaseProvenance{ApprovedBy: "admin", AttackCategory: "injection", Method: "synthetic"}, Input: map[string]any{"prompt": "Injection test"}, Expected: map[string]any{"output": "blocked"}, Tags: []string{"security"}},
		{CaseID: "case_004", Skill: "code_generation", Source: "synthetic", Status: "approved", Provenance: CaseProvenance{GeneratorID: "gen1", Method: "llm"}, Input: map[string]any{"prompt": "Sort array"}, Expected: map[string]any{"output": "sorted"}, Tags: []string{"algorithm"}},
		{CaseID: "case_005", Skill: "code_generation", Source: "golden", Status: "approved", Provenance: CaseProvenance{ApprovedBy: "admin", Method: "manual"}, Input: map[string]any{"prompt": "Binary search"}, Expected: map[string]any{"output": "found"}, Tags: []string{"algorithm"}},
		{CaseID: "case_006", Skill: "text_summarization", Source: "golden", Status: "approved", Provenance: CaseProvenance{ApprovedBy: "admin", Method: "manual"}, Input: map[string]any{"text": "Long article"}, Expected: map[string]any{"summary": "brief"}, Tags: []string{"nlp"}},
		{CaseID: "case_007", Skill: "text_summarization", Source: "production", Status: "approved", Provenance: CaseProvenance{ContributorID: "user2", Method: "curated"}, Input: map[string]any{"text": "News article"}, Expected: map[string]any{"summary": "headline"}, Tags: []string{"news"}},
		{CaseID: "case_008", Skill: "text_summarization", Source: "synthetic", Status: "approved", Provenance: CaseProvenance{GeneratorID: "gen2", Method: "llm"}, Input: map[string]any{"text": "Scientific paper"}, Expected: map[string]any{"summary": "abstract"}, Tags: []string{"academic"}},
		{CaseID: "case_009", Skill: "text_summarization", Source: "adversarial", Status: "approved", Provenance: CaseProvenance{AttackCategory: "bias", Method: "synthetic"}, Input: map[string]any{"text": "Biased content"}, Expected: map[string]any{"summary": "neutral"}, Tags: []string{"fairness"}},
		{CaseID: "case_010", Skill: "text_summarization", Source: "golden", Status: "approved", Provenance: CaseProvenance{ApprovedBy: "admin", Method: "manual"}, Input: map[string]any{"text": "Product review"}, Expected: map[string]any{"summary": "sentiment"}, Tags: []string{"commerce"}},
	}

	runs := []RunResultDTO{
		{
			RunID:     "run_highscore",
			DatasetID: "ds_abc12345",
			Skill:     "code_generation",
			Runtime:   "2026-03-25T10:00:00Z",
			Metrics: RunMetricsDTO{
				SuccessRate:          0.95,
				AvgTokens:            1250.5,
				P95Latency:           2500,
				CostFactor:           0.001,
				ClassificationFactor: 0.92,
				CostUSD:              1.25,
				StabilityVariance:    0.02,
				Score:                0.94,
			},
			Results: []CaseRunResultDTO{
				{CaseID: "case_001", Success: true, LatencyMs: 1500, TokenUsage: 1000, Output: map[string]any{"result": "ok"}, Classification: "correct"},
				{CaseID: "case_002", Success: true, LatencyMs: 2000, TokenUsage: 1500, Output: map[string]any{"result": "ok"}, Classification: "correct"},
				{CaseID: "case_003", Success: true, LatencyMs: 1800, TokenUsage: 1200, Output: map[string]any{"result": "ok"}, Classification: "correct"},
			},
			CreatedAt: "2026-03-25T10:00:00Z",
		},
		{
			RunID:     "run_medscore",
			DatasetID: "ds_abc12345",
			Skill:     "code_generation",
			Runtime:   "2026-03-24T10:00:00Z",
			Metrics: RunMetricsDTO{
				SuccessRate:          0.70,
				AvgTokens:            1100.0,
				P95Latency:           3500,
				CostFactor:           0.001,
				ClassificationFactor: 0.68,
				CostUSD:              1.10,
				StabilityVariance:    0.08,
				Score:                0.69,
			},
			Results: []CaseRunResultDTO{
				{CaseID: "case_001", Success: true, LatencyMs: 1500, TokenUsage: 1000, Output: map[string]any{"result": "ok"}, Classification: "correct"},
				{CaseID: "case_002", Success: false, LatencyMs: 3000, TokenUsage: 1200, Output: map[string]any{"result": "fail"}, Error: "timeout", Classification: "correct"},
			},
			CreatedAt: "2026-03-24T10:00:00Z",
		},
		{
			RunID:     "run_lowscore",
			DatasetID: "ds_def67890",
			Skill:     "text_summarization",
			Runtime:   "2026-03-23T10:00:00Z",
			Metrics: RunMetricsDTO{
				SuccessRate:          0.40,
				AvgTokens:            800.0,
				P95Latency:           4500,
				CostFactor:           0.001,
				ClassificationFactor: 0.35,
				CostUSD:              0.80,
				StabilityVariance:    0.15,
				Score:                0.38,
			},
			Results: []CaseRunResultDTO{
				{CaseID: "case_006", Success: false, LatencyMs: 4000, TokenUsage: 900, Output: map[string]any{"result": "fail"}, Error: "quality_low", Classification: "incorrect"},
				{CaseID: "case_007", Success: false, LatencyMs: 4500, TokenUsage: 700, Output: map[string]any{"result": "fail"}, Error: "quality_low", Classification: "incorrect"},
			},
			CreatedAt: "2026-03-23T10:00:00Z",
		},
	}

	policies := []GovernancePolicyDTO{
		{
			PolicyID:                "pol_strict01",
			Name:                    "Strict Policy",
			RequireProvenance:       true,
			RequireApprovedForScore: true,
			MinSourceDiversity:      3,
			MinGoldenWeight:         0.5,
			SourcePolicies: []SourcePolicy{
				{Source: "golden", Weight: 0.4, CountInScore: true, MinSuccessRate: 0.9},
				{Source: "adversarial", Weight: 0.3, CountInScore: true, MinSuccessRate: 0.7},
				{Source: "synthetic", Weight: 0.3, CountInScore: true, MinSuccessRate: 0.8},
			},
			CreatedAt: "2026-03-01T00:00:00Z",
		},
		{
			PolicyID:                "pol_relaxed02",
			Name:                    "Relaxed Policy",
			RequireProvenance:       false,
			RequireApprovedForScore: false,
			MinSourceDiversity:      2,
			MinGoldenWeight:         0.3,
			SourcePolicies: []SourcePolicy{
				{Source: "golden", Weight: 0.3, CountInScore: true, MinSuccessRate: 0.7},
				{Source: "production", Weight: 0.4, CountInScore: true, MinSuccessRate: 0.6},
				{Source: "synthetic", Weight: 0.3, CountInScore: true, MinSuccessRate: 0.5},
			},
			CreatedAt: "2026-03-10T00:00:00Z",
		},
		{
			PolicyID:                "pol_default",
			Name:                    "Default Policy",
			RequireProvenance:       false,
			RequireApprovedForScore: false,
			MinSourceDiversity:      1,
			MinGoldenWeight:         0.0,
			SourcePolicies: []SourcePolicy{
				{Source: "golden", Weight: 1.0, CountInScore: true, MinSuccessRate: 0.5},
			},
			CreatedAt: "2026-03-15T00:00:00Z",
		},
	}

	if err := s.writeJSON(DatasetsFile, datasets); err != nil {
		return fmt.Errorf("failed to write datasets: %w", err)
	}
	if err := s.writeJSON(CasesFile, cases); err != nil {
		return fmt.Errorf("failed to write cases: %w", err)
	}
	if err := s.writeJSON(RunsFile, runs); err != nil {
		return fmt.Errorf("failed to write runs: %w", err)
	}
	if err := s.writeJSON(PoliciesFile, policies); err != nil {
		return fmt.Errorf("failed to write policies: %w", err)
	}

	return nil
}

func (s *Store) ListDatasets() ([]DatasetSummaryDTO, error) {
	var datasets []DatasetDetailDTO
	if err := s.readJSON(DatasetsFile, &datasets); err != nil {
		return nil, err
	}

	summaries := make([]DatasetSummaryDTO, len(datasets))
	for i, d := range datasets {
		summaries[i] = DatasetSummaryDTO{
			DatasetID: d.DatasetID,
			Name:      d.Name,
			Version:   d.Version,
			Owner:     d.Owner,
			CaseCount: d.CaseCount,
			CreatedAt: d.CreatedAt,
		}
	}
	return summaries, nil
}

func (s *Store) GetDataset(datasetID string) (*DatasetDetailDTO, error) {
	var datasets []DatasetDetailDTO
	if err := s.readJSON(DatasetsFile, &datasets); err != nil {
		return nil, err
	}

	for i := range datasets {
		if datasets[i].DatasetID == datasetID {
			return &datasets[i], nil
		}
	}
	return nil, fmt.Errorf("dataset not found: %s", datasetID)
}

func (s *Store) GetDatasetCases(datasetID string) ([]EvaluationCaseDTO, error) {
	var allCases []EvaluationCaseDTO
	if err := s.readJSON(CasesFile, &allCases); err != nil {
		return nil, err
	}

	var cases []EvaluationCaseDTO
	for _, c := range allCases {
		if s.hasTag(c.Tags, datasetID) || s.caseBelongsToDataset(c.CaseID, datasetID) {
			cases = append(cases, c)
		}
	}

	if len(cases) == 0 {
		cases = allCases[:min(5, len(allCases))]
	}

	return cases, nil
}

func (s *Store) CreateDataset(dataset DatasetDetailDTO) (DatasetDetailDTO, error) {
	datasets, err := s.ListDatasetsDetail()
	if err != nil {
		datasets = []DatasetDetailDTO{}
	}

	dataset.DatasetID = "ds_" + randomID()
	dataset.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	datasets = append(datasets, dataset)

	if err := s.writeJSON(DatasetsFile, datasets); err != nil {
		return DatasetDetailDTO{}, err
	}
	return dataset, nil
}

func (s *Store) ListDatasetsDetail() ([]DatasetDetailDTO, error) {
	var datasets []DatasetDetailDTO
	if err := s.readJSON(DatasetsFile, &datasets); err != nil {
		return nil, err
	}
	return datasets, nil
}

func (s *Store) DeleteDataset(datasetID string) error {
	datasets, err := s.ListDatasetsDetail()
	if err != nil {
		return err
	}

	for i, d := range datasets {
		if d.DatasetID == datasetID {
			datasets = append(datasets[:i], datasets[i+1:]...)
			return s.writeJSON(DatasetsFile, datasets)
		}
	}
	return fmt.Errorf("dataset not found: %s", datasetID)
}

func (s *Store) caseBelongsToDataset(caseID, datasetID string) bool {
	switch datasetID {
	case "ds_abc12345":
		return caseID >= "case_001" && caseID <= "case_005"
	case "ds_def67890":
		return caseID >= "case_006" && caseID <= "case_010"
	}
	return false
}

func (s *Store) VerifyDataset(datasetID string) (*DatasetVerifyDTO, error) {
	dataset, err := s.GetDataset(datasetID)
	if err != nil {
		return nil, err
	}

	cases, err := s.GetDatasetCases(datasetID)
	if err != nil {
		return nil, err
	}

	checksum := s.computeDatasetChecksum(dataset, cases)

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

func (s *Store) computeDatasetChecksum(dataset *DatasetDetailDTO, cases []EvaluationCaseDTO) string {
	h := sha256.New()
	h.Write([]byte(dataset.DatasetID + dataset.Name + dataset.Version))
	h.Write([]byte(fmt.Sprintf("%d", len(cases))))
	for _, c := range cases {
		h.Write([]byte(c.CaseID))
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

func (s *Store) ListRuns() ([]RunResultDTO, error) {
	var runs []RunResultDTO
	if err := s.readJSON(RunsFile, &runs); err != nil {
		return nil, err
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt > runs[j].CreatedAt
	})

	return runs, nil
}

func (s *Store) GetRun(runID string) (*RunResultDTO, error) {
	runs, err := s.ListRuns()
	if err != nil {
		return nil, err
	}

	for i := range runs {
		if runs[i].RunID == runID {
			return &runs[i], nil
		}
	}
	return nil, fmt.Errorf("run not found: %s", runID)
}

func (s *Store) CreateRun(run RunResultDTO) error {
	runs, err := s.ListRuns()
	if err != nil {
		runs = []RunResultDTO{}
	}

	run.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	runs = append(runs, run)

	return s.writeJSON(RunsFile, runs)
}

func (s *Store) GetPolicy(policyID string) (*GovernancePolicyDTO, error) {
	var policies []GovernancePolicyDTO
	if err := s.readJSON(PoliciesFile, &policies); err != nil {
		return nil, err
	}

	for i := range policies {
		if policies[i].PolicyID == policyID {
			return &policies[i], nil
		}
	}
	return nil, fmt.Errorf("policy not found: %s", policyID)
}

func (s *Store) ListPolicies() ([]GovernancePolicyDTO, error) {
	var policies []GovernancePolicyDTO
	if err := s.readJSON(PoliciesFile, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

func (s *Store) CreatePolicy(policy GovernancePolicyDTO) (GovernancePolicyDTO, error) {
	policies, err := s.ListPolicies()
	if err != nil {
		policies = []GovernancePolicyDTO{}
	}

	policy.PolicyID = "pol_" + randomID()
	policy.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	policies = append(policies, policy)

	if err := s.writeJSON(PoliciesFile, policies); err != nil {
		return GovernancePolicyDTO{}, err
	}
	return policy, nil
}

func (s *Store) StoreAuthToken(token, role, apiKeyHash string) error {
	records := []AuthTokenRecord{}
	if s.fileExists(AuthTokensFile) {
		if err := s.readJSON(AuthTokensFile, &records); err != nil {
			records = []AuthTokenRecord{}
		}
	}

	records = append(records, AuthTokenRecord{
		Token:      token,
		Role:       role,
		APIKeyHash: apiKeyHash,
		IssuedAt:   time.Now().UTC(),
	})

	return s.writeJSON(AuthTokensFile, records)
}

func (s *Store) hasTag(tags []string, datasetID string) bool {
	for _, t := range tags {
		if t == datasetID {
			return true
		}
	}
	return false
}

func (s *Store) ListCases() ([]EvaluationCaseDTO, error) {
	var cases []EvaluationCaseDTO
	err := s.readJSON("cases.json", &cases)
	if err != nil {
		return []EvaluationCaseDTO{}, nil
	}
	return cases, nil
}

func (s *Store) CreateCase(caseDTO EvaluationCaseDTO) error {
	cases, _ := s.ListCases()
	cases = append(cases, caseDTO)
	return s.writeJSON("cases.json", cases)
}

func (s *Store) GetCase(caseID string) (*EvaluationCaseDTO, error) {
	cases, err := s.ListCases()
	if err != nil {
		return nil, err
	}
	for i := range cases {
		if cases[i].CaseID == caseID {
			return &cases[i], nil
		}
	}
	return nil, nil
}

func (s *Store) UpdateCase(caseID string, updates map[string]interface{}) error {
	cases, err := s.ListCases()
	if err != nil {
		return err
	}
	for i := range cases {
		if cases[i].CaseID == caseID {
			if status, ok := updates["status"].(string); ok {
				cases[i].Status = status
			}
			return s.writeJSON("cases.json", cases)
		}
	}
	return fmt.Errorf("case not found")
}

func (s *Store) ListReleases() ([]map[string]interface{}, error) {
	var releases []map[string]interface{}
	err := s.readJSON("releases.json", &releases)
	if err != nil {
		return []map[string]interface{}{}, nil
	}
	return releases, nil
}

func (s *Store) CreateRelease(release map[string]interface{}) error {
	releases, _ := s.ListReleases()
	releases = append(releases, release)
	return s.writeJSON("releases.json", releases)
}

func (s *Store) GetRelease(releaseID string) (map[string]interface{}, error) {
	releases, err := s.ListReleases()
	if err != nil {
		return nil, err
	}
	for _, r := range releases {
		if r["release_id"] == releaseID {
			return r, nil
		}
	}
	return nil, nil
}

func (s *Store) UpdateRelease(releaseID string, release map[string]interface{}) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			releases[i] = release
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found")
}

func (s *Store) GetReleaseForUpdate(releaseID string) (map[string]interface{}, error) {
	releases, err := s.ListReleases()
	if err != nil {
		return nil, err
	}
	for _, r := range releases {
		if r["release_id"] == releaseID {
			return r, nil
		}
	}
	return nil, fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) ApproveRelease(releaseID, approvedBy string) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			r["status"] = "approved"
			r["approved_by"] = approvedBy
			r["approved_at"] = time.Now().UTC().Format(time.RFC3339)
			releases[i] = r
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) RejectRelease(releaseID, rejectedReason string) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			r["status"] = "rejected"
			r["rejected_reason"] = rejectedReason
			releases[i] = r
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) RollbackRelease(releaseID string) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			r["status"] = "rolled_back"
			releases[i] = r
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) GetCaseLineage(caseID string) (map[string]interface{}, error) {
	var lineage map[string]interface{}
	if err := s.readJSON("lineage_cases.json", &lineage); err != nil {
		return map[string]interface{}{
			"ancestors":   []interface{}{},
			"descendants": []interface{}{},
		}, nil
	}
	return lineage, nil
}

func (s *Store) GetDatasetLineage(datasetID string) (map[string]interface{}, error) {
	var lineage map[string]interface{}
	if err := s.readJSON("lineage_datasets.json", &lineage); err != nil {
		return map[string]interface{}{
			"versions": []interface{}{},
		}, nil
	}
	return lineage, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
