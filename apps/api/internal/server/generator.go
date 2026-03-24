package server

import (
	"errors"
	"net/http"

	"github.com/zealot/managing-up/apps/api/internal/generator"
	"github.com/zealot/managing-up/apps/api/internal/llm"
)

func (s *Server) handleGenerateSkill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req struct {
		SOPText     string `json:"sop_text"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	cfg := llm.ConfigFromEnv()
	client, err := llm.NewClient(cfg.Provider, cfg.Model, cfg.APIKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create LLM client")
		return
	}

	gen := generator.NewGenerator(client)
	resp, err := gen.GenerateFromSOP(r.Context(), generator.GenerateRequest{
		SOPText:     req.SOPText,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, generator.ErrInvalidSpec):
			writeError(w, http.StatusBadRequest, "INVALID_SPEC", err.Error())
		case errors.Is(err, generator.ErrValidationFailed):
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		case errors.Is(err, generator.ErrLLMCallFailed):
			writeError(w, http.StatusBadRequest, "LLM_CALL_FAILED", err.Error())
		default:
			writeError(w, http.StatusBadRequest, "GENERATION_FAILED", err.Error())
		}
		return
	}

	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{
		"spec":     resp.Spec,
		"provider": resp.Provider,
		"model":    resp.Model,
		"usage":    resp.Usage,
	})
}
