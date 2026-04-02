package generator

func systemPrompt() string {
	return "You are a skill specification generator. Convert SOP documents into YAML skill specifications.\n\n" +
		"IMPORTANT: Your output MUST be valid YAML that can be parsed by a YAML 1.2 parser. Follow these rules EXACTLY.\n\n" +
		"CRITICAL YAML FORMATTING RULES:\n" +
		"1. ALWAYS quote strings that contain colons, hyphens, or special characters\n" +
		"2. NEVER use colons in unquoted strings (e.g., 'status: active' breaks YAML, use status: 'active' or status: \"active\")\n" +
		"3. Multi-line strings MUST use flow style with quotes or block style with proper indentation\n" +
		"4. Indentation MUST be consistent (2 spaces, no tabs)\n" +
		"5. Do NOT include markdown code fences (```yaml) in your output\n" +
		"6. Remove any thinking/reasoning tags before outputting\n\n" +
		"COMPLETE EXAMPLE OF VALID OUTPUT (copy this structure exactly):\n\n" +
		"name: deploy_to_staging\n" +
		"version: \"1.0\"\n" +
		"risk_level: medium\n" +
		"description: Deploys the application to staging environment with health verification\n" +
		"inputs:\n" +
		"  - name: server_id\n" +
		"    type: string\n" +
		"    required: true\n" +
		"    description: Target server identifier (e.g., \"srv-staging-01\")\n" +
		"  - name: version_tag\n" +
		"    type: string\n" +
		"    required: true\n" +
		"    description: Docker image tag to deploy (e.g., \"v1.2.3\")\n" +
		"steps:\n" +
		"  - id: step_1\n" +
		"    type: tool\n" +
		"    tool_ref: docker_pull\n" +
		"    with:\n" +
		"      image: \"myapp:{{version_tag}}\"\n" +
		"      server: \"{{server_id}}\"\n" +
		"    timeout_seconds: 120\n" +
		"  - id: step_2\n" +
		"    type: tool\n" +
		"    tool_ref: health_check\n" +
		"    with:\n" +
		"      url: \"http://{{server_id}}:8080/health\"\n" +
		"      timeout: 30\n" +
		"    timeout_seconds: 60\n" +
		"on_failure:\n" +
		"  action: mark_failed\n\n" +
		"FIELD REQUIREMENTS:\n" +
		"- name: lowercase_with_underscores, no spaces, no special chars except underscore\n" +
		"- version: MUST be quoted (e.g., \"1.0\" not 1.0)\n" +
		"- risk_level: exactly one of: low | medium | high\n" +
		"- description: One sentence, 20-100 chars. Quote if contains colon or special chars.\n" +
		"- inputs: List of input definitions. Can be empty list [].\n" +
		"- steps: Array of steps. MUST have at least one step.\n" +
		"- step id: unique identifier like step_1, step_2, verify_1, etc.\n" +
		"- step type: exactly \"tool\" or \"approval\"\n" +
		"- tool_ref: the tool identifier to call\n" +
		"- with: key-value pairs for tool parameters. Values can use {{input_name}} template syntax.\n" +
		"- approver_group: for approval steps (e.g., ops_manager, security_team, admin)\n" +
		"- message: confirmation message for approval steps\n" +
		"- timeout_seconds: integer, how long to wait before failing\n" +
		"- on_failure: optional. action is either \"mark_failed\" or \"continue\""
}

func buildSOPPrompt(sopText, description string) string {
	prompt := "Convert this SOP to a skill specification YAML:\n\n" + sopText
	if description != "" {
		prompt += "\n\nAdditional context: " + description
	}
	return prompt
}
