package generator

func systemPrompt() string {
	return "You are a skill specification generator. Convert SOP documents into YAML skill specifications.\n\n" +
		"Output MUST be valid YAML with this schema:\n" +
		"```yaml\n" +
		"name: skill_name          # lowercase_with_underscores\n" +
		"version: \"1.0\"\n" +
		"risk_level: low|medium|high\n" +
		"description: One sentence description\n" +
		"inputs:\n" +
		"  - name: input_name\n" +
		"    type: string|number|boolean|object\n" +
		"    required: true|false\n" +
		"    description: input description\n" +
		"steps:\n" +
		"  - id: step_1\n" +
		"    type: tool|approval\n" +
		"    tool_ref: tool_name      # if type=tool\n" +
		"    with:\n" +
		"      key: value             # templated values like {{input_name}}\n" +
		"    approver_group: group    # if type=approval\n" +
		"    message: \"Confirm?\"      # if type=approval\n" +
		"    timeout_seconds: 30\n" +
		"on_failure:\n" +
		"  action: mark_failed|continue\n" +
		"```\n\n" +
		"Rules:\n" +
		"- Skill name must be lowercase_with_underscores\n" +
		"- Steps must have unique IDs\n" +
		"- At least one step is required\n" +
		"- risk_level: low (read-only), medium (write operations), high (system changes, approvals)\n" +
		"- approval steps: set approver_group to appropriate team (ops_manager, security_team, etc.)\n"
}

func buildSOPPrompt(sopText, description string) string {
	prompt := "Convert this SOP to a skill specification YAML:\n\n" + sopText
	if description != "" {
		prompt += "\n\nAdditional context: " + description
	}
	return prompt
}
