# Skill Hub Python SDK

Install:

```bash
pip install skill-hub
```

Usage:

```python
from skill_hub import SkillHubClient

client = SkillHubClient(
    base_url="http://localhost:8080",
    agent_id="my-agent-v1"
)

# Register agent
client.register("My Agent", "1.0.0", ["code_execution", "file_read"])

# List skills
skills = client.list_skills(risk_level="low")

# Get skill spec
spec_yaml = client.get_skill_spec("skill_001")

# Execute
result = client.execute("skill_001", {"server_id": "srv-001"})
print(result["execution_id"])
```
