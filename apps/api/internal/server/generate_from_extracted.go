package server

import (
	"net/http"

	"github.com/zealot/managing-up/apps/api/internal/generator"
	"github.com/zealot/managing-up/apps/api/internal/llm"
)

func (s *Server) handleGenerateFromExtracted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req struct {
		SkillName     string `json:"skill_name"`
		ExtractedData struct {
			Constraints []map[string]any `json:"constraints"`
			Decisions   []map[string]any `json:"decisions"`
			Roles       []string         `json:"roles"`
		} `json:"extracted_data"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.SkillName == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skill_name is required")
		return
	}

	if len(req.ExtractedData.Constraints) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "extracted_data.constraints is required")
		return
	}

	// Build prompt from extracted data
	prompt := buildPromptFromExtracted(req.SkillName, req.ExtractedData.Constraints, req.ExtractedData.Roles)

	// Call LLM to generate
	cfg := llm.ConfigFromEnv()
	client, err := llm.NewClient(cfg.Provider, cfg.Model, cfg.APIKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create LLM client")
		return
	}

	gen := generator.NewGenerator(client)
	resp, err := gen.GenerateFromSOP(r.Context(), generator.GenerateRequest{
		SOPText:     prompt,
		Description: req.SkillName,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "GENERATION_FAILED", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{
		"generated_yaml": resp.SpecYAML,
		"provider":       resp.Provider,
		"model":          resp.Model,
		"usage":          resp.Usage,
	})
}

func buildPromptFromExtracted(skillName string, constraints []map[string]any, roles []string) string {
	prompt := "请根据以下从SOP文档提取的结构化数据，生成一个标准的Skill YAML规范。\n\n"
	prompt += "Skill名称: " + skillName + "\n\n"
	prompt += "提取的约束规则:\n"

	for i, c := range constraints {
		_ = i // unused
		level := c["level"]
		desc := c["description"]
		prompt += "- " + formatLevel(level) + ": " + formatString(desc) + "\n"
		if cond, ok := c["condition"].(string); ok && cond != "" {
			prompt += "  条件: " + cond + "\n"
		}
		if action, ok := c["action"].(string); ok && action != "" {
			prompt += "  动作: " + action + "\n"
		}
	}

	if len(roles) > 0 {
		prompt += "\n识别的角色:\n"
		for _, role := range roles {
			prompt += "- " + role + "\n"
		}
	}

	prompt += "\n请生成符合规范的SKILL.yaml格式，包含triggers、steps、constraints等完整结构。"
	return prompt
}

func formatLevel(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "MUST"
}

func formatString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
