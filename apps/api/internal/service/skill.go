package service

import (
	"context"
	"strings"
)

type Skill struct {
	ID             string
	Name           string
	OwnerTeam      string
	RiskLevel      string
	Status         string
	CurrentVersion string
	CreatedBy      string
}

type CreateSkillRequest struct {
	Name      string
	OwnerTeam string
	RiskLevel string
}

type SkillRepository interface {
	ListSkills(status string) []Skill
	GetSkill(id string) (Skill, bool)
	CreateSkill(req CreateSkillRequest) Skill
	ListDependencies(ctx context.Context, skillID string) ([]SkillDependency, error)
	UpsertRating(ctx context.Context, skillID, userID string, rating int, comment string) error
	ListSkillsByCategory(ctx context.Context, category, search string) ([]Skill, error)
	GetRatingStats(ctx context.Context, skillID string) (float64, int, error)
	GetInstallCount(ctx context.Context, skillID string) (int, error)
	ResolveDepTree(ctx context.Context, skillID string) ([]DependencyNode, error)
}

var ValidRiskLevels = []string{"low", "medium", "high"}

type SkillService struct {
	repo SkillRepository
}

func NewSkillService(repo SkillRepository) *SkillService {
	return &SkillService{repo: repo}
}

func (s *SkillService) CreateSkill(req CreateSkillRequest) (Skill, error) {
	if req.Name == "" {
		return Skill{}, ErrSkillNameRequired
	}

	if req.OwnerTeam == "" {
		return Skill{}, ErrOwnerTeamRequired
	}

	if !isValidRiskLevel(req.RiskLevel) {
		return Skill{}, ErrInvalidRiskLevel
	}

	if err := s.validateUniqueName(req.Name); err != nil {
		return Skill{}, err
	}

	return s.repo.CreateSkill(req), nil
}

func (s *SkillService) validateUniqueName(name string) error {
	skills := s.repo.ListSkills("")
	for _, skill := range skills {
		if strings.EqualFold(skill.Name, name) {
			return ErrDuplicateSkillName
		}
	}
	return nil
}

func isValidRiskLevel(level string) bool {
	for _, valid := range ValidRiskLevels {
		if level == valid {
			return true
		}
	}
	return false
}
