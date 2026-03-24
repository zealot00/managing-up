package service

type Execution struct {
	ID            string
	SkillID       string
	SkillName     string
	Status        string
	TriggeredBy   string
	CurrentStepID string
	Input         map[string]any
}

type Approval struct {
	ID             string
	ExecutionID    string
	SkillName      string
	StepID         string
	Status         string
	ApproverGroup  string
	ApprovedBy     string
	ResolutionNote string
}

type CreateExecutionRequest struct {
	SkillID     string
	TriggeredBy string
	Input       map[string]any
}

type ApproveExecutionRequest struct {
	Approver string
	Decision string
	Note     string
}

var validDecisions = []string{"approved", "rejected"}

type ExecutionRepository interface {
	GetSkill(id string) (Skill, bool)
	CreateExecution(req CreateExecutionRequest) (Execution, bool)
	ApproveExecution(executionID string, req ApproveExecutionRequest) (Approval, bool)
}

type ExecutionService struct {
	repo ExecutionRepository
}

func NewExecutionService(repo ExecutionRepository) *ExecutionService {
	return &ExecutionService{repo: repo}
}

func (s *ExecutionService) CreateExecution(req CreateExecutionRequest) (Execution, error) {
	if req.SkillID == "" {
		return Execution{}, ErrSkillIDRequired
	}

	if req.TriggeredBy == "" {
		return Execution{}, ErrTriggeredByRequired
	}

	if _, ok := s.repo.GetSkill(req.SkillID); !ok {
		return Execution{}, ErrSkillNotFound
	}

	execution, ok := s.repo.CreateExecution(req)
	if !ok {
		return Execution{}, ErrSkillNotFound
	}

	return execution, nil
}

func (s *ExecutionService) ApproveExecution(executionID string, req ApproveExecutionRequest) (Approval, error) {
	if req.Approver == "" {
		return Approval{}, ErrApproverRequired
	}

	if !isValidDecision(req.Decision) {
		return Approval{}, ErrInvalidDecision
	}

	approval, ok := s.repo.ApproveExecution(executionID, req)
	if !ok {
		return Approval{}, ErrExecutionNotFound
	}

	return approval, nil
}

func isValidDecision(decision string) bool {
	for _, valid := range validDecisions {
		if decision == valid {
			return true
		}
	}
	return false
}
