export type DashboardData = {
  summary: {
    active_skills: number;
    published_versions: number;
    running_executions: number;
    waiting_approvals: number;
    success_rate: number;
    avg_duration_seconds: number;
  };
  recent_executions: Array<{
    id: string;
    skill_name: string;
    status: string;
    started_at: string;
    current_step_id: string;
  }>;
};

export type Skill = {
  id: string;
  name: string;
  owner_team: string;
  risk_level: string;
  status: string;
  current_version: string;
  description?: string;
  draft_source?: string;
  verified?: boolean;
  sop_name?: string;
};

export type SkillVersion = {
  id: string;
  skill_id: string;
  version: string;
  status: string;
  change_summary: string;
  approval_required: boolean;
  created_at: string;
};

export type Execution = {
  id: string;
  skill_id: string;
  skill_name: string;
  status: string;
  triggered_by: string;
  started_at: string;
  current_step_id: string;
  input?: Record<string, unknown>;
};

export type Approval = {
  id: string;
  execution_id: string;
  skill_name: string;
  step_id: string;
  status: string;
  approver_group: string;
  requested_at: string;
};

export type ProcedureDraft = {
  id: string;
  procedure_key: string;
  title: string;
  validation_status: string;
  required_tools: string[];
  source_type: string;
  created_at: string;
};

// P0-P2: Trace Store
export type TraceEvent = {
  id: string;
  execution_id: string;
  step_id: string;
  event_type: string;
  event_data: Record<string, unknown>;
  timestamp: string;
};

// P1: Task Registry v2
export type TestCase = {
  input: Record<string, unknown>;
  expected: unknown;
};

export type Task = {
  id: string;
  name: string;
  description: string;
  skill_id: string;
  tags: string[];
  difficulty: "easy" | "medium" | "hard";
  test_cases: TestCase[];
  created_at: string;
  updated_at: string;
};

// P1: Evaluation Engine
export type Metric = {
  id: string;
  name: string;
  type: string;
  config: Record<string, unknown>;
  created_at: string;
};

export type TaskExecution = {
  id: string;
  task_id: string;
  agent_id: string;
  status: string;
  input: Record<string, unknown>;
  output: Record<string, unknown>;
  duration_ms: number;
  created_at: string;
};

export type Evaluation = {
  id: string;
  task_execution_id: string;
  metric_id: string;
  score: number;
  details: Record<string, unknown>;
  evaluated_at: string;
};

export type Experiment = {
  id: string;
  name: string;
  description: string;
  task_ids: string[];
  agent_ids: string[];
  status: string;
  created_at: string;
  updated_at: string;
};

export type ReplaySnapshot = {
  id: string;
  execution_id: string;
  skill_id: string;
  skill_version: string;
  step_index: number;
  state_snapshot: Record<string, unknown>;
  input_seed: Record<string, unknown>;
  deterministic_seed: number;
  created_at: string;
};

type Envelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
  meta: { request_id: string };
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

async function readEnvelope<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`API request failed for ${path} with status ${response.status}`);
  }

  const body = (await response.json()) as Envelope<T>;
  return body.data;
}

export async function getDashboard(): Promise<DashboardData> {
  return readEnvelope<DashboardData>("/api/v1/dashboard");
}

export async function getSkills(): Promise<{ items: Skill[] }> {
  return readEnvelope<{ items: Skill[] }>("/api/v1/skills");
}

export async function getSkill(id: string): Promise<Skill> {
  return readEnvelope<Skill>(`/api/v1/skills/${id}`);
}

export async function getSkillVersions(): Promise<{ items: SkillVersion[] }> {
  return readEnvelope<{ items: SkillVersion[] }>("/api/v1/skill-versions");
}

export async function getSkillVersionsBySkillId(skillId: string): Promise<{ items: SkillVersion[] }> {
  return readEnvelope<{ items: SkillVersion[] }>(`/api/v1/skills/${skillId}/versions`);
}

export async function getSkillSpec(skillId: string): Promise<{ spec_yaml: string }> {
  return readEnvelope<{ spec_yaml: string }>(`/api/v1/skills/${skillId}/spec`);
}

export async function getExecutions(): Promise<{ items: Execution[] }> {
  return readEnvelope<{ items: Execution[] }>("/api/v1/executions");
}

export async function getExecution(id: string): Promise<Execution> {
  return readEnvelope<Execution>(`/api/v1/executions/${id}`);
}

export async function getApprovals(): Promise<{ items: Approval[] }> {
  return readEnvelope<{ items: Approval[] }>("/api/v1/approvals");
}

export async function getProcedureDrafts(): Promise<{ items: ProcedureDraft[] }> {
  return readEnvelope<{ items: ProcedureDraft[] }>("/api/v1/procedure-drafts");
}

// POST functions

export type CreateSkillRequest = {
  name: string;
  owner_team: string;
  risk_level: string;
};

export type CreateSkillVersionRequest = {
  skill_id: string;
  version: string;
  change_summary: string;
  approval_required: boolean;
  spec_yaml: string;
};

export type CreateExecutionRequest = {
  skill_id: string;
  triggered_by: string;
  input: Record<string, unknown>;
};

export type ApproveExecutionRequest = {
  approver: string;
  decision: "approved" | "rejected";
  note: string;
};

async function postEnvelope<T, R>(path: string, body: T): Promise<R> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`API POST failed for ${path} with status ${response.status}`);
  }

  const bodyEnvelope = (await response.json()) as Envelope<R>;
  if (bodyEnvelope.error) {
    throw new Error(bodyEnvelope.error.message);
  }
  return bodyEnvelope.data;
}

export async function createSkill(req: CreateSkillRequest): Promise<Skill> {
  return postEnvelope<CreateSkillRequest, Skill>("/api/v1/skills", req);
}

export async function createSkillVersion(req: CreateSkillVersionRequest): Promise<SkillVersion> {
  return postEnvelope<CreateSkillVersionRequest, SkillVersion>("/api/v1/skill-versions", req);
}

export async function createExecution(req: CreateExecutionRequest): Promise<Execution> {
  return postEnvelope<CreateExecutionRequest, Execution>("/api/v1/executions", req);
}

export async function approveExecution(
  executionId: string,
  req: ApproveExecutionRequest
): Promise<Approval> {
  return postEnvelope<ApproveExecutionRequest, Approval>(
    `/api/v1/executions/${executionId}/approve`,
    req
  );
}

// Trace Store API
export async function getTraces(executionId: string): Promise<TraceEvent[]> {
  return readEnvelope<TraceEvent[]>(`/api/v1/executions/${executionId}/traces`);
}

// Task Registry API
export async function getTasks(): Promise<{ items: Task[] }> {
  return readEnvelope<{ items: Task[] }>("/api/v1/tasks");
}

export async function getTask(id: string): Promise<Task> {
  return readEnvelope<Task>(`/api/v1/tasks/${id}`);
}

export type CreateTaskRequest = {
  name: string;
  description: string;
  skill_id: string;
  tags: string[];
  difficulty: string;
  test_cases: TestCase[];
};

export async function createTask(req: CreateTaskRequest): Promise<Task> {
  return postEnvelope<CreateTaskRequest, Task>("/api/v1/tasks", req);
}

export async function updateTask(id: string, req: Partial<CreateTaskRequest>): Promise<Task> {
  return postEnvelope<Partial<CreateTaskRequest>, Task>(`/api/v1/tasks/${id}`, req);
}

export async function deleteTask(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/tasks/${id}`, {
    method: "DELETE",
    cache: "no-store",
  });
  if (!response.ok) {
    throw new Error(`API DELETE failed for /api/v1/tasks/${id}`);
  }
}

export type BuildTaskFromTraceRequest = {
  execution_id?: string;
  trace_id?: string;
  name?: string;
  description?: string;
  tags?: string[];
  difficulty?: string;
};

export async function buildTaskFromTrace(req: BuildTaskFromTraceRequest): Promise<Task> {
  return postEnvelope<BuildTaskFromTraceRequest, Task>("/api/v1/tasks/from-trace", req);
}

// Metric API
export async function getMetrics(): Promise<{ items: Metric[] }> {
  return readEnvelope<{ items: Metric[] }>("/api/v1/metrics");
}

export type CreateMetricRequest = {
  name: string;
  type: string;
  config: Record<string, unknown>;
};

export async function createMetric(req: CreateMetricRequest): Promise<Metric> {
  return postEnvelope<CreateMetricRequest, Metric>("/api/v1/metrics", req);
}

// Evaluation API
export async function getTaskExecutions(): Promise<{ items: TaskExecution[] }> {
  return readEnvelope<{ items: TaskExecution[] }>("/api/v1/task-executions");
}

export async function getTaskExecution(id: string): Promise<TaskExecution> {
  return readEnvelope<TaskExecution>(`/api/v1/task-executions/${id}`);
}

export type RunEvaluationRequest = {
  task_id: string;
  agent_id: string;
  input: Record<string, unknown>;
};

export async function runTaskEvaluation(req: RunEvaluationRequest): Promise<TaskExecution> {
  return postEnvelope<RunEvaluationRequest, TaskExecution>("/api/v1/task-executions", req);
}

export async function getEvaluations(taskExecutionId: string): Promise<{ items: Evaluation[] }> {
  return readEnvelope<{ items: Evaluation[] }>(`/api/v1/task-executions/${taskExecutionId}/evaluations`);
}

export type EvaluateRequest = {
  metric_id: string;
};

export async function evaluateTaskExecution(
  taskExecutionId: string,
  req: EvaluateRequest
): Promise<Evaluation> {
  return postEnvelope<EvaluateRequest, Evaluation>(
    `/api/v1/task-executions/${taskExecutionId}/evaluate`,
    req
  );
}

// Experiment API
export async function getExperiments(): Promise<{ items: Experiment[] }> {
  return readEnvelope<{ items: Experiment[] }>("/api/v1/experiments");
}

export async function getExperiment(id: string): Promise<Experiment> {
  return readEnvelope<Experiment>(`/api/v1/experiments/${id}`);
}

export type CreateExperimentRequest = {
  name: string;
  description: string;
  task_ids: string[];
  agent_ids: string[];
};

export async function createExperiment(req: CreateExperimentRequest): Promise<Experiment> {
  return postEnvelope<CreateExperimentRequest, Experiment>("/api/v1/experiments", req);
}

// ReplaySnapshot API
export async function getReplaySnapshots(executionId?: string): Promise<{ items: ReplaySnapshot[] }> {
  const query = executionId ? `?execution_id=${executionId}` : "";
  return readEnvelope<{ items: ReplaySnapshot[] }>(`/api/v1/replay-snapshots${query}`);
}

export async function getReplaySnapshot(id: string): Promise<ReplaySnapshot> {
  return readEnvelope<ReplaySnapshot>(`/api/v1/replay-snapshots/${id}`);
}

// Capabilities API
export type CapabilityScore = {
  experimentId: string;
  experimentName: string;
  score: number;
  timestamp: string;
};

export type CapabilityGraphNode = {
  name: string;
  score: number;
  sampleSize: number;
  scores: CapabilityScore[];
};

export async function getCapabilities(): Promise<{ data: CapabilityGraphNode[] }> {
  const response = await readEnvelope<{ capabilities: CapabilityGraphNode[] }>("/api/v1/capabilities");
  return { data: response.capabilities };
}

export async function getCapability(name: string): Promise<CapabilityGraphNode> {
  return readEnvelope<CapabilityGraphNode>(`/api/v1/capabilities/${name}`);
}

// Experiment API
export async function runExperiment(id: string): Promise<{ status: string; message: string }> {
  return postEnvelope<Record<string, never>, { status: string; message: string }>(`/api/v1/experiments/${id}/run`, {});
}

export async function compareExperiments(id: string, compareWithId: string): Promise<{ experiment: string; compare_with: string; deltas: Array<{ task_id: string; exp_score: number; other_score: number; delta: number }>; regression: boolean }> {
  return readEnvelope<{ experiment: string; compare_with: string; deltas: Array<{ task_id: string; exp_score: number; other_score: number; delta: number }>; regression: boolean }>(`/api/v1/experiments/${id}/compare?compare_with=${compareWithId}`);
}

export type CheckRegressionRequest = {
  current_score: number;
  baseline_score: number;
  threshold?: number;
};

export async function checkRegression(req: CheckRegressionRequest): Promise<{ current_score: number; baseline_score: number; delta: number; regression: boolean }> {
  return postEnvelope<CheckRegressionRequest, { current_score: number; baseline_score: number; delta: number; regression: boolean }>("/api/v1/check-regression", req);
}

// MCP Router API
export type MCPRouterCatalogEntry = {
  id: string;
  server_id: string;
  name: string;
  description?: string;
  transport_type: string;
  task_types: string[];
  tags?: string[];
  trust_score: number;
  use_count: number;
  status: string;
};

export type RouteRequest = {
  task: {
    description?: string;
    structured?: {
      task_type?: string;
      language?: string;
      complexity?: string;
      tags?: string[];
    };
  };
  agent_id?: string;
  correlation_id?: string;
};

export type RouteResponse = {
  matched: boolean;
  target?: {
    server_id: string;
    server_name: string;
    transport: string;
    endpoint?: string;
  };
  match_score?: number;
  routing_time_ms?: number;
};

export async function getMCPRouterCatalog(): Promise<MCPRouterCatalogEntry[]> {
  return readEnvelope<MCPRouterCatalogEntry[]>("/api/v1/router/mcp/catalog");
}

export async function routeMCP(request: RouteRequest): Promise<RouteResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/router/mcp/route`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  const body = await response.json();
  return body.data;
}

export async function matchMCPRouter(taskType: string, tags?: string[]): Promise<RouteResponse | null> {
  const params = new URLSearchParams({ task_type: taskType });
  if (tags?.length) {
    params.set("tags", tags.join(","));
  }
  return readEnvelope<RouteResponse | null>(`/api/v1/router/mcp/match?${params}`);
}

// Skill Market API
export type SkillMarketEntry = {
  id: string;
  name: string;
  version: string;
  description?: string;
  category?: string;
  tags?: string[];
  trust_score: number;
  verified: boolean;
  avg_rating: number;
  rating_count: number;
  install_count: number;
  sop_name?: string;
  sop_version?: string;
};

export async function getSkillMarket(params?: { category?: string; search?: string }): Promise<SkillMarketEntry[]> {
  const url = new URL(`${API_BASE_URL}/api/v1/skills/market`);
  if (params?.category) url.searchParams.set("category", params.category);
  if (params?.search) url.searchParams.set("search", params.search);
  return readEnvelope<SkillMarketEntry[]>(url.pathname + url.search);
}

export async function searchSkills(query: string): Promise<SkillMarketEntry[]> {
  return readEnvelope<SkillMarketEntry[]>(`/api/v1/skills/search?q=${encodeURIComponent(query)}`);
}

export async function rateSkill(skillId: string, rating: number, comment?: string): Promise<void> {
  await postEnvelope<{ rating: number; comment?: string }, void>(`/api/v1/skills/${skillId}/rate`, { rating, comment });
}

export async function getMySkills(): Promise<Skill[]> {
  return readEnvelope<Skill[]>("/api/v1/skills");
}
