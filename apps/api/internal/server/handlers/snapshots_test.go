package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSnapshotsHandler_GetLatestSnapshot(t *testing.T) {
	t.Parallel()

	now := time.Now()
	handler := NewSnapshotsHandler(&mockSnapshotRepo{
		snapshot: &SnapshotDTO{
			ID:           "snap_001",
			SkillID:      "skill_001",
			Version:      "v1.0.0",
			SnapshotType: "evaluation",
			Metrics:      map[string]float64{"accuracy": 0.95},
			OverallScore: 0.95,
			Passed:       true,
			EvaluatedAt:  now.Format(time.RFC3339),
			CreatedAt:    now.Format(time.RFC3339),
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots?skill_id=skill_001&version=v1.0.0", nil)
	rec := httptest.NewRecorder()

	handler.GetLatestSnapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data object in response")
	}

	if data["found"] != true {
		t.Fatalf("expected found to be true")
	}
}

func TestSnapshotsHandler_GetLatestSnapshot_NotFound(t *testing.T) {
	t.Parallel()

	handler := NewSnapshotsHandler(&mockSnapshotRepo{
		snapshot: nil,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots?skill_id=skill_001&version=v1.0.0", nil)
	rec := httptest.NewRecorder()

	handler.GetLatestSnapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data object in response")
	}

	if data["found"] != false {
		t.Fatalf("expected found to be false")
	}
}

func TestSnapshotsHandler_GetLatestSnapshot_MissingParams(t *testing.T) {
	t.Parallel()

	handler := NewSnapshotsHandler(&mockSnapshotRepo{})

	tests := []struct {
		name string
		path string
	}{
		{"missing skill_id", "/api/v1/snapshots?version=v1.0.0"},
		{"missing version", "/api/v1/snapshots?skill_id=skill_001"},
		{"missing both", "/api/v1/snapshots"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.GetLatestSnapshot(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}
}

func TestSnapshotsHandler_ListSnapshots(t *testing.T) {
	t.Parallel()

	now := time.Now()
	handler := NewSnapshotsHandler(&mockSnapshotRepo{
		snapshots: []SnapshotDTO{
			{
				ID:           "snap_001",
				SkillID:      "skill_001",
				Version:      "v1.0.0",
				SnapshotType: "evaluation",
				Metrics:      map[string]float64{"accuracy": 0.95},
				OverallScore: 0.95,
				Passed:       true,
				EvaluatedAt:  now.Format(time.RFC3339),
				CreatedAt:    now.Format(time.RFC3339),
			},
			{
				ID:           "snap_002",
				SkillID:      "skill_001",
				Version:      "v1.0.1",
				SnapshotType: "evaluation",
				Metrics:      map[string]float64{"accuracy": 0.97},
				OverallScore: 0.97,
				Passed:       true,
				EvaluatedAt:  now.Format(time.RFC3339),
				CreatedAt:    now.Format(time.RFC3339),
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/list?skill_id=skill_001", nil)
	rec := httptest.NewRecorder()

	handler.ListSnapshots(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestSnapshotsHandler_ListSnapshots_WithLimit(t *testing.T) {
	t.Parallel()

	handler := NewSnapshotsHandler(&mockSnapshotRepo{
		snapshots: []SnapshotDTO{
			{ID: "snap_001", SkillID: "skill_001"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/list?skill_id=skill_001&limit=5", nil)
	rec := httptest.NewRecorder()

	handler.ListSnapshots(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !handler.repo.(*mockSnapshotRepo).listCalledWithLimit(5) {
		t.Fatalf("expected limit to be passed as 5")
	}
}

func TestSnapshotsHandler_ListSnapshots_MissingSkillID(t *testing.T) {
	t.Parallel()

	handler := NewSnapshotsHandler(&mockSnapshotRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/list", nil)
	rec := httptest.NewRecorder()

	handler.ListSnapshots(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

type mockSnapshotRepo struct {
	snapshot  *SnapshotDTO
	snapshots []SnapshotDTO
	limitUsed int
}

func (m *mockSnapshotRepo) GetLatestSnapshot(ctx context.Context, skillID, version string) (*SnapshotDTO, error) {
	return m.snapshot, nil
}

func (m *mockSnapshotRepo) ListSnapshots(ctx context.Context, skillID string, limit int) ([]SnapshotDTO, error) {
	m.limitUsed = limit
	return m.snapshots, nil
}

func (m *mockSnapshotRepo) listCalledWithLimit(expected int) bool {
	return m.limitUsed == expected
}