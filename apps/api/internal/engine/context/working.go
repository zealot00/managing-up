package context

import "time"

type WorkingContext struct {
	*BaseContext
	CurrentTask    *TaskState
	CompletedSteps []CompletedStep
	PendingActions []PendingAction
}

type TaskState struct {
	TaskID      string `json:"task_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
}

type CompletedStep struct {
	StepID      string    `json:"step_id"`
	Description string    `json:"description"`
	Result      string    `json:"result"`
	Timestamp   time.Time `json:"timestamp"`
}

type PendingAction struct {
	ActionID    string   `json:"action_id"`
	Description string   `json:"description"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

func NewWorkingContext(id ContextID) *WorkingContext {
	now := time.Now()
	return &WorkingContext{
		BaseContext: &BaseContext{
			id:        id,
			ctype:     TypeWorking,
			createdAt: now,
			updatedAt: now,
		},
		CurrentTask:    nil,
		CompletedSteps: []CompletedStep{},
		PendingActions: []PendingAction{},
	}
}

func (w *WorkingContext) SetTask(taskID, description string) {
	w.CurrentTask = &TaskState{
		TaskID:      taskID,
		Description: description,
		Status:      "in_progress",
		Progress:    0,
	}
	w.touch()
}

func (w *WorkingContext) CompleteStep(stepID, description, result string) {
	w.CompletedSteps = append(w.CompletedSteps, CompletedStep{
		StepID:      stepID,
		Description: description,
		Result:      result,
		Timestamp:   time.Now(),
	})
	if w.CurrentTask != nil {
		w.CurrentTask.Progress = len(w.CompletedSteps)
	}
	w.touch()
}

func (w *WorkingContext) AddPendingAction(actionID, description string, dependsOn []string) {
	w.PendingActions = append(w.PendingActions, PendingAction{
		ActionID:    actionID,
		Description: description,
		DependsOn:   dependsOn,
	})
	w.touch()
}

func (w *WorkingContext) RemovePendingAction(actionID string) {
	for i, a := range w.PendingActions {
		if a.ActionID == actionID {
			w.PendingActions = append(w.PendingActions[:i], w.PendingActions[i+1:]...)
			w.touch()
			return
		}
	}
}

func (w *WorkingContext) IsActionReady(actionID string) bool {
	for _, a := range w.PendingActions {
		if a.ActionID == actionID {
			for _, dep := range a.DependsOn {
				if !w.isStepCompleted(dep) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (w *WorkingContext) isStepCompleted(stepID string) bool {
	for _, s := range w.CompletedSteps {
		if s.StepID == stepID {
			return true
		}
	}
	return false
}

func (w *WorkingContext) CompleteTask() {
	if w.CurrentTask != nil {
		w.CurrentTask.Status = "completed"
		w.CurrentTask.Progress = 100
	}
	w.touch()
}
