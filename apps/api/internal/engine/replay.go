package engine

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type Snapshot struct {
	ID                string          `json:"id"`
	ExecutionID       string          `json:"execution_id"`
	SkillID           string          `json:"skill_id"`
	SkillVersion      string          `json:"skill_version"`
	StepIndex         int             `json:"step_index"`
	StateSnapshot     json.RawMessage `json:"state_snapshot"`
	InputSeed         json.RawMessage `json:"input_seed"`
	DeterministicSeed int64           `json:"deterministic_seed"`
}

type ReplayState struct {
	ExecutionID      string         `json:"execution_id"`
	CurrentStepIndex int            `json:"current_step_index"`
	Variables        map[string]any `json:"variables"`
	StepResults      []StepResult   `json:"step_results"`
}

type StepResult struct {
	StepID string         `json:"step_id"`
	Output map[string]any `json:"output,omitempty"`
	Error  string         `json:"error,omitempty"`
}

type DeterministicRNG struct {
	seed int64
}

func NewDeterministicRNG(seed int64) *DeterministicRNG {
	return &DeterministicRNG{seed: seed}
}

func (r *DeterministicRNG) NextInt63() int64 {
	r.seed = (r.seed*6364136223846793005 + 1442695040888963407) % (1 << 62)
	return r.seed
}

func (r *DeterministicRNG) Int63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return r.NextInt63() % n
}

type SnapshotStore interface {
	Save(snapshot Snapshot) error
	Get(id string) (Snapshot, bool)
	ListByExecution(execID string) []Snapshot
}

type ReplayEngine struct {
	specParser    *SkillSpecParser
	snapshotStore SnapshotStore
	rng           *DeterministicRNG
}

func NewReplayEngine(store SnapshotStore) *ReplayEngine {
	return &ReplayEngine{
		specParser:    NewSkillSpecParser(),
		snapshotStore: store,
	}
}

func (e *ReplayEngine) CreateSnapshot(ctx context.Context, executionID, skillID, skillVersion string, stepIndex int, state ReplayState) (Snapshot, error) {
	seed, err := generateSeed()
	if err != nil {
		return Snapshot{}, fmt.Errorf("failed to generate seed: %w", err)
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return Snapshot{}, fmt.Errorf("failed to marshal state: %w", err)
	}

	inputSeedJSON, err := json.Marshal(state.Variables)
	if err != nil {
		return Snapshot{}, fmt.Errorf("failed to marshal input seed: %w", err)
	}

	snapshot := Snapshot{
		ID:                fmt.Sprintf("snap_%d", time.Now().UnixNano()),
		ExecutionID:       executionID,
		SkillID:           skillID,
		SkillVersion:      skillVersion,
		StepIndex:         stepIndex,
		StateSnapshot:     stateJSON,
		InputSeed:         inputSeedJSON,
		DeterministicSeed: seed,
	}

	if err := e.snapshotStore.Save(snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("failed to save snapshot: %w", err)
	}

	return snapshot, nil
}

func (e *ReplayEngine) Replay(ctx context.Context, snapshotID string) (ReplayState, error) {
	snapshot, ok := e.snapshotStore.Get(snapshotID)
	if !ok {
		return ReplayState{}, fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	var state ReplayState
	if err := json.Unmarshal(snapshot.StateSnapshot, &state); err != nil {
		return ReplayState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	e.rng = NewDeterministicRNG(snapshot.DeterministicSeed)

	return state, nil
}

func (e *ReplayEngine) GetDeterministicRNG() *DeterministicRNG {
	if e.rng == nil {
		e.rng = NewDeterministicRNG(0)
	}
	return e.rng
}

func (e *ReplayEngine) GenerateDeterministicChoice(options []string) string {
	if len(options) == 0 {
		return ""
	}
	rng := e.GetDeterministicRNG()
	idx := rng.Int63n(int64(len(options)))
	return options[idx]
}

func generateSeed() (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

type DeterministicExecutor struct {
	snapshotStore SnapshotStore
	specParser    *SkillSpecParser
	toolGateway   *ToolGateway
}

func NewDeterministicExecutor(store SnapshotStore, toolGateway *ToolGateway) *DeterministicExecutor {
	return &DeterministicExecutor{
		snapshotStore: store,
		specParser:    NewSkillSpecParser(),
		toolGateway:   toolGateway,
	}
}

func (e *DeterministicExecutor) ExecuteFromSnapshot(ctx context.Context, snapshotID string) error {
	snapshot, ok := e.snapshotStore.Get(snapshotID)
	if !ok {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	var state ReplayState
	if err := json.Unmarshal(snapshot.StateSnapshot, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	spec, err := e.specParser.Parse("")
	if err != nil {
		return fmt.Errorf("failed to parse skill spec: %w", err)
	}

	for i := snapshot.StepIndex; i < len(spec.Steps); i++ {
		step := spec.Steps[i]

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if step.Type == "tool" {
			toolInput := make(map[string]any)
			for k, v := range step.With {
				toolInput[k] = v
			}

			inv := ToolInvocation{
				ExecutionID:    state.ExecutionID,
				StepID:         step.ID,
				ToolRef:        step.ToolRef,
				Input:          toolInput,
				TimeoutSeconds: step.TimeoutSeconds,
			}

			result, err := e.toolGateway.Invoke(ctx, inv)
			if err != nil {
				state.StepResults = append(state.StepResults, StepResult{
					StepID: step.ID,
					Error:  err.Error(),
				})
				continue
			}

			state.StepResults = append(state.StepResults, StepResult{
				StepID: step.ID,
				Output: result.Output,
			})
		}
	}

	return nil
}
