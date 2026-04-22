package handlers

import (
	"context"
	"net/http"
	"strconv"
)

type SnapshotRepo interface {
	GetLatestSnapshot(ctx context.Context, skillID, version string) (*SnapshotDTO, error)
	ListSnapshots(ctx context.Context, skillID string, limit int) ([]SnapshotDTO, error)
}

type SnapshotDTO struct {
	ID            string             `json:"id"`
	SkillID       string             `json:"skill_id"`
	Version       string             `json:"version"`
	SnapshotType  string             `json:"snapshot_type"`
	DatasetID     string             `json:"dataset_id,omitempty"`
	RunID         string             `json:"run_id,omitempty"`
	Metrics       map[string]float64 `json:"metrics"`
	OverallScore  float64            `json:"overall_score"`
	Passed        bool               `json:"passed"`
	GatePolicyID  string             `json:"gate_policy_id,omitempty"`
	EvaluatedAt   string             `json:"evaluated_at"`
	CreatedAt     string             `json:"created_at"`
}

type SnapshotsHandler struct {
	repo SnapshotRepo
}

func NewSnapshotsHandler(repo SnapshotRepo) *SnapshotsHandler {
	return &SnapshotsHandler{repo: repo}
}

func (h *SnapshotsHandler) GetLatestSnapshot(w http.ResponseWriter, r *http.Request) {
	skillID := r.URL.Query().Get("skill_id")
	version := r.URL.Query().Get("version")

	if skillID == "" || version == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skill_id and version are required")
		return
	}

	ctx := r.Context()
	snapshot, err := h.repo.GetLatestSnapshot(ctx, skillID, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if snapshot == nil {
		writeEnvelope(w, http.StatusOK, "snapshot", map[string]interface{}{
			"found":    false,
			"skill_id": skillID,
			"version":  version,
		})
		return
	}

	writeEnvelope(w, http.StatusOK, "snapshot", map[string]interface{}{
		"found":    true,
		"snapshot": snapshot,
	})
}

func (h *SnapshotsHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	skillID := r.URL.Query().Get("skill_id")
	if skillID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skill_id is required")
		return
	}

	limit := 10
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := r.Context()
	snapshots, err := h.repo.ListSnapshots(ctx, skillID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "snapshots", snapshots)
}