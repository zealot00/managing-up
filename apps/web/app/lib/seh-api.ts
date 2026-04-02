const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const SEH_BASE_URL = `${API_BASE_URL}/v1/seh`;

type Envelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
  meta: { request_id: string };
};

async function readEnvelope<T>(path: string): Promise<T> {
  const response = await fetch(`${path}`, {
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`API request failed for ${path} with status ${response.status}`);
  }

  const body = (await response.json()) as Envelope<T>;
  return body.data;
}

// Dataset Types
export type DatasetSummary = {
  dataset_id: string;
  name: string;
  version: string;
  owner: string;
  case_count: number;
  created_at: string;
};

export type DatasetDetail = {
  dataset_id: string;
  name: string;
  version: string;
  owner: string;
  description: string;
  manifest: Record<string, string>;
  case_count: number;
  checksum: string;
  created_at: string;
};

export type DatasetListResponse = {
  datasets: DatasetSummary[];
  pagination: {
    limit: number;
    offset: number;
    total: number;
    has_more: boolean;
  };
};

// Run Types
export type RunMetrics = {
  success_rate: number;
  avg_tokens: number;
  p95_latency: number;
  cost_factor: number;
  cost_usd: number;
  stability_variance: number;
  score: number;
  classification_factor: number;
};

export type CaseRunResult = {
  case_id: string;
  success: boolean;
  latency_ms: number;
  token_usage: number;
  output: Record<string, unknown>;
  classification: string;
  error?: string;
};

export type RunResult = {
  run_id: string;
  dataset_id: string;
  skill: string;
  runtime: string;
  metrics: RunMetrics;
  results: CaseRunResult[];
  created_at: string;
};

export type RunListResponse = {
  runs: RunResult[];
  pagination: {
    limit: number;
    offset: number;
    total: number;
    has_more: boolean;
  };
};

export type RunCreateResponse = {
  run_id: string;
  score: number;
  success_rate: number;
  created_at: string;
};

// Policy Types
export type SourcePolicy = {
  source: string;
  weight: number;
  count_in_score: boolean;
  min_success_rate: number;
};

export type GovernancePolicy = {
  policy_id: string;
  name: string;
  require_provenance: boolean;
  require_approved_for_score: boolean;
  min_source_diversity: number;
  min_golden_weight: number;
  source_policies: SourcePolicy[] | null;
  created_at: string;
};

// Release Types
export type Release = {
  release_id: string;
  status: string;
  created_at: string;
  approved_by?: string;
  approved_at?: string;
  rejected_reason?: string;
};

// Lineage Types
export type CaseLineage = {
  ancestors: unknown[];
  descendants: unknown[];
};

export type DatasetLineage = {
  versions: unknown[];
};

// Case Types
export type EvaluationCase = {
  case_id: string;
  skill: string;
  source: string;
  status: string;
  provenance: {
    approved_by?: string;
    contributor_id?: string;
    method: string;
    attack_category?: string;
    generator_id?: string;
  };
  input: Record<string, unknown>;
  expected: Record<string, unknown>;
  tags: string[];
};

// Dashboard Summary
export type SEHDashboardSummary = {
  total_datasets: number;
  total_runs: number;
  total_policies: number;
  total_cases: number;
  recent_runs: RunResult[];
  avg_score: number;
  avg_success_rate: number;
};

// SEH API Functions

export async function getSEHDatasets(limit = 20, offset = 0): Promise<DatasetListResponse> {
  return readEnvelope<DatasetListResponse>(`${SEH_BASE_URL}/datasets?limit=${limit}&offset=${offset}`);
}

export async function getSEHDataset(datasetId: string): Promise<DatasetDetail> {
  return readEnvelope<DatasetDetail>(`${SEH_BASE_URL}/datasets/${datasetId}`);
}

export async function createSEHDataset(data: {
  name: string;
  version: string;
  owner: string;
  description?: string;
}): Promise<{ dataset_id: string; created_at: string }> {
  const response = await fetch(`${SEH_BASE_URL}/datasets`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error(`Failed to create dataset: ${response.status}`);
  return response.json();
}

export async function deleteSEHDataset(datasetId: string): Promise<void> {
  const response = await fetch(`${SEH_BASE_URL}/datasets/${datasetId}`, {
    method: "DELETE",
  });
  if (!response.ok && response.status !== 204) {
    throw new Error(`Failed to delete dataset: ${response.status}`);
  }
}

export async function getSEHRuns(limit = 20, offset = 0): Promise<RunListResponse> {
  return readEnvelope<RunListResponse>(`${SEH_BASE_URL}/runs?limit=${limit}&offset=${offset}`);
}

export async function getSEHRun(runId: string): Promise<RunResult> {
  return readEnvelope<RunResult>(`${SEH_BASE_URL}/runs/${runId}`);
}

export async function getSEHPolicies(): Promise<GovernancePolicy[]> {
  return readEnvelope<GovernancePolicy[]>(`${SEH_BASE_URL}/policies`);
}

export async function createSEHPolicy(data: {
  name: string;
  require_provenance?: boolean;
  require_approved_for_score?: boolean;
  min_source_diversity?: number;
  min_golden_weight?: number;
  source_policies?: SourcePolicy[];
}): Promise<{ policy_id: string; created_at: string }> {
  const response = await fetch(`${SEH_BASE_URL}/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error(`Failed to create policy: ${response.status}`);
  return response.json();
}

export async function createSEHRelease(skillId: string): Promise<Release> {
  const response = await fetch(`${SEH_BASE_URL}/skills/${skillId}/releases`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  if (!response.ok) throw new Error(`Failed to create release: ${response.status}`);
  return response.json();
}

export async function approveSEHRelease(
  releaseId: string,
  approvedBy: string
): Promise<{ release_id: string; status: string; approved_by: string; approved_at: string }> {
  const response = await fetch(`${SEH_BASE_URL}/releases/${releaseId}/approve`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ approved_by: approvedBy }),
  });
  if (!response.ok) throw new Error(`Failed to approve release: ${response.status}`);
  return response.json();
}

export async function rejectSEHRelease(
  releaseId: string,
  reason: string
): Promise<{ release_id: string; status: string; rejected_reason: string }> {
  const response = await fetch(`${SEH_BASE_URL}/releases/${releaseId}/reject`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ reason }),
  });
  if (!response.ok) throw new Error(`Failed to reject release: ${response.status}`);
  return response.json();
}

export async function rollbackSEHRelease(
  releaseId: string
): Promise<{ release_id: string; status: string }> {
  const response = await fetch(`${SEH_BASE_URL}/releases/${releaseId}/rollback`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  if (!response.ok) throw new Error(`Failed to rollback release: ${response.status}`);
  return response.json();
}

export async function getSEHDatasetLineage(datasetId: string): Promise<DatasetLineage> {
  return readEnvelope<DatasetLineage>(`${SEH_BASE_URL}/datasets/${datasetId}/lineage`);
}

export async function getSEHCaseLineage(caseId: string): Promise<CaseLineage> {
  return readEnvelope<CaseLineage>(`${SEH_BASE_URL}/cases/${caseId}/lineage`);
}

export async function getSEHCases(): Promise<EvaluationCase[]> {
  return readEnvelope<EvaluationCase[]>(`${SEH_BASE_URL}/cases`);
}

export async function getSEHDashboardSummary(): Promise<SEHDashboardSummary> {
  try {
    const [datasets, runs, policies, cases] = await Promise.all([
      getSEHDatasets(1, 0),
      getSEHRuns(5, 0),
      getSEHPolicies(),
      getSEHCases().catch(() => []),
    ]);

    const runsWithScores = runs.runs.filter((r) => r.metrics?.score > 0);
    const avgScore = runsWithScores.length > 0
      ? runsWithScores.reduce((sum, r) => sum + r.metrics.score, 0) / runsWithScores.length
      : 0;
    const avgSuccessRate = runsWithScores.length > 0
      ? runsWithScores.reduce((sum, r) => sum + r.metrics.success_rate, 0) / runsWithScores.length
      : 0;

    return {
      total_datasets: datasets.pagination.total,
      total_runs: runs.pagination.total,
      total_policies: policies.length,
      total_cases: cases.length,
      recent_runs: runs.runs.slice(0, 5),
      avg_score: avgScore,
      avg_success_rate: avgSuccessRate,
    };
  } catch (error) {
    console.error("Failed to fetch SEH dashboard summary:", error);
    return {
      total_datasets: 0,
      total_runs: 0,
      total_policies: 0,
      total_cases: 0,
      recent_runs: [],
      avg_score: 0,
      avg_success_rate: 0,
    };
  }
}
