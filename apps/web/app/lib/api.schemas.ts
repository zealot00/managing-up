import { z } from "zod";

// Envelope wrapper for API responses
export const EnvelopeSchema = <T extends z.ZodTypeAny>(dataSchema: T) => z.object({
  data: dataSchema,
  error: z.object({
    code: z.string(),
    message: z.string(),
  }).nullable(),
  meta: z.object({
    request_id: z.string(),
  }),
});

// Core types
export const SkillSchema = z.object({
  id: z.string(),
  name: z.string(),
  owner_team: z.string(),
  risk_level: z.string(),
  status: z.string(),
  current_version: z.string().nullable(),
});

export const SkillVersionSchema = z.object({
  id: z.string(),
  skill_id: z.string(),
  version: z.string(),
  status: z.string(),
  change_summary: z.string(),
  approval_required: z.boolean(),
  created_at: z.string(),
});

export const ExecutionSchema = z.object({
  id: z.string(),
  skill_id: z.string(),
  skill_name: z.string(),
  status: z.string(),
  triggered_by: z.string(),
  started_at: z.string(),
  current_step_id: z.string(),
  input: z.record(z.string(), z.unknown()).optional(),
});

export const ApprovalSchema = z.object({
  id: z.string(),
  execution_id: z.string(),
  skill_name: z.string(),
  step_id: z.string(),
  status: z.string(),
  approver_group: z.string(),
  requested_at: z.string(),
});

export const ProcedureDraftSchema = z.object({
  id: z.string(),
  procedure_key: z.string(),
  title: z.string(),
  validation_status: z.string(),
  required_tools: z.array(z.string()),
  source_type: z.string(),
  created_at: z.string(),
});

// P0-P2: Trace Store
export const TraceEventSchema = z.object({
  id: z.string(),
  execution_id: z.string(),
  step_id: z.string(),
  event_type: z.string(),
  event_data: z.record(z.string(), z.unknown()),
  timestamp: z.string(),
});

// P1: Task Registry v2
export const TestCaseSchema = z.object({
  input: z.record(z.string(), z.unknown()),
  expected: z.unknown(),
});

export const TaskSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  skill_id: z.string(),
  tags: z.array(z.string()),
  difficulty: z.enum(["easy", "medium", "hard"]),
  test_cases: z.array(TestCaseSchema),
  created_at: z.string(),
  updated_at: z.string(),
});

// P1: Evaluation Engine
export const MetricSchema = z.object({
  id: z.string(),
  name: z.string(),
  type: z.string(),
  config: z.record(z.string(), z.unknown()),
  created_at: z.string(),
});

export const TaskExecutionSchema = z.object({
  id: z.string(),
  task_id: z.string(),
  agent_id: z.string(),
  status: z.string(),
  input: z.record(z.string(), z.unknown()),
  output: z.record(z.string(), z.unknown()),
  duration_ms: z.number(),
  created_at: z.string(),
});

export const EvaluationSchema = z.object({
  id: z.string(),
  task_execution_id: z.string(),
  metric_id: z.string(),
  score: z.number(),
  details: z.record(z.string(), z.unknown()),
  evaluated_at: z.string(),
});

export const ExperimentSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  task_ids: z.array(z.string()),
  agent_ids: z.array(z.string()),
  status: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const ReplaySnapshotSchema = z.object({
  id: z.string(),
  execution_id: z.string(),
  skill_id: z.string(),
  skill_version: z.string(),
  step_index: z.number(),
  state_snapshot: z.record(z.string(), z.unknown()),
  input_seed: z.record(z.string(), z.unknown()),
  deterministic_seed: z.number(),
  created_at: z.string(),
});

// Dashboard
export const DashboardDataSchema = z.object({
  summary: z.object({
    active_skills: z.number(),
    published_versions: z.number(),
    running_executions: z.number(),
    waiting_approvals: z.number(),
    success_rate: z.number(),
    avg_duration_seconds: z.number(),
  }),
  recent_executions: z.array(z.object({
    id: z.string(),
    skill_name: z.string(),
    status: z.string(),
    started_at: z.string(),
    current_step_id: z.string(),
  })),
});

// Capabilities
export const CapabilityScoreSchema = z.object({
  experimentId: z.string(),
  experimentName: z.string(),
  score: z.number(),
  timestamp: z.string(),
});

export const CapabilityGraphNodeSchema = z.object({
  name: z.string(),
  score: z.number(),
  sampleSize: z.number(),
  scores: z.array(CapabilityScoreSchema),
});

// Experiment comparison
export const ExperimentDeltaSchema = z.object({
  task_id: z.string(),
  exp_score: z.number(),
  other_score: z.number(),
  delta: z.number(),
});

export const ExperimentCompareSchema = z.object({
  experiment: z.string(),
  compare_with: z.string(),
  deltas: z.array(ExperimentDeltaSchema),
  regression: z.boolean(),
});

// Export inferred types
export type Skill = z.infer<typeof SkillSchema>;
export type SkillVersion = z.infer<typeof SkillVersionSchema>;
export type Execution = z.infer<typeof ExecutionSchema>;
export type Approval = z.infer<typeof ApprovalSchema>;
export type ProcedureDraft = z.infer<typeof ProcedureDraftSchema>;
export type TraceEvent = z.infer<typeof TraceEventSchema>;
export type TestCase = z.infer<typeof TestCaseSchema>;
export type Task = z.infer<typeof TaskSchema>;
export type Metric = z.infer<typeof MetricSchema>;
export type TaskExecution = z.infer<typeof TaskExecutionSchema>;
export type Evaluation = z.infer<typeof EvaluationSchema>;
export type Experiment = z.infer<typeof ExperimentSchema>;
export type ReplaySnapshot = z.infer<typeof ReplaySnapshotSchema>;
export type DashboardData = z.infer<typeof DashboardDataSchema>;
export type CapabilityScore = z.infer<typeof CapabilityScoreSchema>;
export type CapabilityGraphNode = z.infer<typeof CapabilityGraphNodeSchema>;
export type ExperimentDelta = z.infer<typeof ExperimentDeltaSchema>;
export type ExperimentCompare = z.infer<typeof ExperimentCompareSchema>;