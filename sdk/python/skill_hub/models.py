from typing import Any, Optional

from pydantic import BaseModel


class SkillSummary(BaseModel):
    id: str
    name: str
    description: Optional[str] = None
    risk_level: str
    version: Optional[str] = None
    tools: list[str] = []
    created_at: Optional[str] = None


class SkillDetail(BaseModel):
    id: str
    name: str
    owner_team: str
    risk_level: str
    status: str
    current_version: Optional[str] = None
    created_by: Optional[str] = None
    updated_at: Optional[str] = None


class ExecutionRequest(BaseModel):
    skill_id: str
    agent_id: str
    input: dict[str, Any] = {}
    callback_url: Optional[str] = None


class ExecutionResponse(BaseModel):
    execution_id: str
    status: str
    skill_id: str
    current_step: Optional[str] = None
    created_at: Optional[str] = None


class ExecutionDetail(BaseModel):
    id: str
    skill_id: str
    skill_name: str
    status: str
    triggered_by: str
    current_step_id: Optional[str] = None
    started_at: Optional[str] = None
    input: dict[str, Any] = {}


class AgentRegistration(BaseModel):
    agent_id: str
    name: str
    version: Optional[str] = None
    capabilities: list[str] = []
    callback_url: Optional[str] = None


class AgentResponse(BaseModel):
    agent_id: str
    name: str
    version: Optional[str] = None
    capabilities: list[str] = []
    registered_at: Optional[str] = None
