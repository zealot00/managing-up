import { z } from "zod";

export const createTaskSchema = z.object({
  name: z.string().min(1, "Task name is required"),
  description: z.string().optional(),
  skill_id: z.string().optional(),
  difficulty: z.enum(["easy", "medium", "hard"]),
  tags: z.string().optional(),
  test_cases: z.string().optional(),
});

export const updateTaskSchema = createTaskSchema;

export const triggerExecutionSchema = z.object({
  skill_id: z.string().min(1, "Please select a skill"),
  triggered_by: z.string().min(1, "Triggered by is required"),
  input: z.string().optional(),
});

export const createSkillSchema = z.object({
  name: z.string().min(1, "Skill name is required"),
  owner_team: z.string().min(1, "Owner team is required"),
  risk_level: z.enum(["low", "medium", "high"]),
});

export const createExperimentSchema = z.object({
  name: z.string().min(1, "Experiment name is required"),
  description: z.string().optional(),
  task_ids: z.string().optional(),
  agent_ids: z.string().optional(),
});

export const createMetricSchema = z.object({
  name: z.string().min(1, "Metric name is required"),
  type: z.enum(["exact_match", "llm_judge", "custom"]),
  config: z.string().optional(),
});

export const createDatasetSchema = z.object({
  name: z.string().min(1, "Dataset name is required"),
  version: z.string().min(1, "Version is required"),
  owner: z.string().min(1, "Owner is required"),
  description: z.string().optional(),
});

export const createSkillVersionSchema = z.object({
  skill_id: z.string().min(1, "Please select a skill"),
  version: z.string().min(1, "Version is required"),
  change_summary: z.string().min(1, "Change summary is required"),
  approval_required: z.boolean(),
  spec_yaml: z.string().min(1, "YAML spec is required"),
});

export const approvalSchema = z.object({
  approver: z.string().min(1, "Approver name is required"),
  note: z.string().optional(),
});

export const runEvaluationSchema = z.object({
  task_id: z.string().min(1, "Please select a task"),
  agent_id: z.string().min(1, "Agent ID is required"),
  input: z.string().optional(),
});

export const jsonStringSchema = z.string().refine(
  (val) => {
    if (!val.trim()) return true;
    try {
      JSON.parse(val);
      return true;
    } catch {
      return false;
    }
  },
  { message: "Invalid JSON format" }
);