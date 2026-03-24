from typing import Optional

import requests


class SkillHubClient:
    def __init__(self, base_url: str, agent_id: str):
        self.base_url = base_url.rstrip("/")
        self.agent_id = agent_id
        self.session = requests.Session()
        self.session.headers["User-Agent"] = f"agent/{agent_id}"

    def list_skills(
        self, risk_level: Optional[str] = None, tool_ref: Optional[str] = None
    ):
        params = {}
        if risk_level:
            params["risk_level"] = risk_level
        if tool_ref:
            params["tool_ref"] = tool_ref

        resp = self.session.get(f"{self.base_url}/api/v1/skills", params=params)
        resp.raise_for_status()
        return resp.json()["data"]["skills"]

    def get_skill(self, skill_id: str):
        resp = self.session.get(f"{self.base_url}/api/v1/skills/{skill_id}")
        resp.raise_for_status()
        return resp.json()["data"]

    def get_skill_spec(self, skill_id: str) -> str:
        resp = self.session.get(f"{self.base_url}/api/v1/skills/{skill_id}/spec")
        resp.raise_for_status()
        return resp.text

    def execute(self, skill_id: str, input: dict, callback_url: Optional[str] = None):
        payload = {
            "skill_id": skill_id,
            "agent_id": self.agent_id,
            "input": input,
        }
        if callback_url:
            payload["callback_url"] = callback_url

        resp = self.session.post(f"{self.base_url}/api/v1/executions", json=payload)
        resp.raise_for_status()
        return resp.json()["data"]

    def get_execution(self, execution_id: str):
        resp = self.session.get(f"{self.base_url}/api/v1/executions/{execution_id}")
        resp.raise_for_status()
        return resp.json()["data"]

    def register(self, name: str, version: str, capabilities: list[str]):
        payload = {
            "agent_id": self.agent_id,
            "name": name,
            "version": version,
            "capabilities": capabilities,
        }
        resp = self.session.post(f"{self.base_url}/api/v1/agents", json=payload)
        resp.raise_for_status()
        return resp.json()["data"]
