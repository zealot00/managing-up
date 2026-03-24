interface Envelope<T> {
  data: T;
  error?: { code: string; message: string };
  meta?: { request_id: string };
}

interface SkillListData {
  skills: SkillSummary[];
}

interface SkillSpecEnvelope {
  data: {
    spec_yaml: string;
  };
}

export class SkillHubClient {
  constructor(
    private baseUrl: string,
    private agentId: string
  ) {}

  async listSkills(opts?: { riskLevel?: string; toolRef?: string }): Promise<SkillSummary[]> {
    const params = new URLSearchParams();
    if (opts?.riskLevel) params.set("risk_level", opts.riskLevel);
    if (opts?.toolRef) params.set("tool_ref", opts.toolRef);

    const resp = await fetch(`${this.baseUrl}/api/v1/skills?${params}`);
    if (!resp.ok) throw new Error(`Failed to list skills: ${resp.status}`);
    const json = await resp.json() as Envelope<SkillListData>;
    return json.data.skills;
  }

  async getSkill(skillId: string): Promise<SkillDetail> {
    const resp = await fetch(`${this.baseUrl}/api/v1/skills/${skillId}`);
    if (!resp.ok) throw new Error(`Failed to get skill: ${resp.status}`);
    const json = await resp.json() as Envelope<SkillDetail>;
    return json.data;
  }

  async getSkillSpec(skillId: string): Promise<string> {
    const resp = await fetch(`${this.baseUrl}/api/v1/skills/${skillId}/spec`);
    if (!resp.ok) throw new Error(`Failed to get skill spec: ${resp.status}`);
    const contentType = resp.headers.get("content-type") || "";
    if (contentType.includes("yaml")) {
      return resp.text();
    }
    const json = await resp.json() as SkillSpecEnvelope;
    return json.data.spec_yaml;
  }

  async execute(
    skillId: string,
    input: Record<string, unknown>,
    callbackUrl?: string
  ): Promise<ExecutionResponse> {
    const body: Record<string, unknown> = {
      skill_id: skillId,
      agent_id: this.agentId,
      input,
    };
    if (callbackUrl) body.callback_url = callbackUrl;

    const resp = await fetch(`${this.baseUrl}/api/v1/executions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!resp.ok) throw new Error(`Failed to execute: ${resp.status}`);
    const json = await resp.json() as Envelope<ExecutionResponse>;
    return json.data;
  }

  async getExecution(executionId: string): Promise<ExecutionDetail> {
    const resp = await fetch(
      `${this.baseUrl}/api/v1/executions/${executionId}`
    );
    if (!resp.ok)
      throw new Error(`Failed to get execution: ${resp.status}`);
    const json = await resp.json() as Envelope<ExecutionDetail>;
    return json.data;
  }

  async register(
    name: string,
    version: string,
    capabilities: string[]
  ): Promise<AgentResponse> {
    const body = {
      agent_id: this.agentId,
      name,
      version,
      capabilities,
    };

    const resp = await fetch(`${this.baseUrl}/api/v1/agents`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!resp.ok) throw new Error(`Failed to register agent: ${resp.status}`);
    const json = await resp.json() as Envelope<AgentResponse>;
    return json.data;
  }
}

export interface SkillSummary {
  id: string;
  name: string;
  description?: string;
  risk_level: string;
  version?: string;
  tools: string[];
  created_at?: string;
}

export interface SkillDetail {
  id: string;
  name: string;
  owner_team: string;
  risk_level: string;
  status: string;
  current_version?: string;
  created_by?: string;
  updated_at?: string;
}

export interface ExecutionResponse {
  execution_id: string;
  status: string;
  skill_id: string;
  current_step?: string;
  created_at?: string;
}

export interface ExecutionDetail {
  id: string;
  skill_id: string;
  skill_name: string;
  status: string;
  triggered_by: string;
  current_step_id?: string;
  started_at?: string;
  input: Record<string, unknown>;
}

export interface AgentResponse {
  agent_id: string;
  name: string;
  version?: string;
  capabilities: string[];
  registered_at?: string;
}
