package service

import "errors"

var (
	ErrSkillNotFound          = errors.New("skill not found")
	ErrExecutionNotFound      = errors.New("execution not found")
	ErrInvalidRiskLevel       = errors.New("invalid risk level: must be one of low, medium, high")
	ErrInvalidDecision        = errors.New("invalid decision: must be approved or rejected")
	ErrSkillNameRequired      = errors.New("skill name is required")
	ErrOwnerTeamRequired      = errors.New("owner team is required")
	ErrSkillIDRequired        = errors.New("skill_id is required")
	ErrTriggeredByRequired    = errors.New("triggered_by is required")
	ErrApproverRequired       = errors.New("approver is required")
	ErrDuplicateSkillName     = errors.New("skill with this name already exists")
	ErrTaskNotFound           = errors.New("task not found")
	ErrTaskNameRequired       = errors.New("task name is required")
	ErrInvalidDifficulty      = errors.New("invalid difficulty: must be one of easy, medium, hard")
	ErrMetricNotFound         = errors.New("metric not found")
	ErrMetricNameRequired     = errors.New("metric name is required")
	ErrInvalidMetricType      = errors.New("invalid metric type")
	ErrEvaluationNotFound     = errors.New("evaluation not found")
	ErrExperimentNameRequired = errors.New("experiment name is required")
	ErrInvalidRating          = errors.New("invalid rating: must be between 1 and 5")
)
