# Skill Hub TypeScript SDK

Install:

```bash
npm install @skill-hub/typescript
```

Usage:

```typescript
import { SkillHubClient } from "@skill-hub/typescript";

const client = new SkillHubClient(
  "http://localhost:8080",
  "my-agent-v1"
);

// Register agent
await client.register("My Agent", "1.0.0", ["code_execution", "file_read"]);

// List skills
const skills = await client.listSkills({ riskLevel: "low" });

// Get skill spec
const specYaml = await client.getSkillSpec("skill_001");

// Execute
const result = await client.execute("skill_001", { server_id: "srv-001" });
console.log(result.execution_id);
```
