package service

import (
	"context"
	"fmt"
)

type SkillEnterpriseService struct {
	repo SkillRepository
}

func NewSkillEnterpriseService(repo SkillRepository) *SkillEnterpriseService {
	return &SkillEnterpriseService{repo: repo}
}

func (s *SkillEnterpriseService) GetSkillWithDeps(ctx context.Context, skillID string) (*SkillWithDeps, error) {
	skill, found := s.repo.GetSkill(skillID)
	if !found {
		return nil, ErrSkillNotFound
	}

	deps, err := s.repo.ListDependencies(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	return &SkillWithDeps{
		Skill:        skill,
		Dependencies: deps,
	}, nil
}

func (s *SkillEnterpriseService) RateSkill(ctx context.Context, skillID, userID string, rating int, comment string) error {
	if rating < 1 || rating > 5 {
		return ErrInvalidRating
	}
	return s.repo.UpsertRating(ctx, skillID, userID, rating, comment)
}

func (s *SkillEnterpriseService) GetSkillMarket(ctx context.Context, category, search string) ([]SkillMarketEntry, error) {
	skills, err := s.repo.ListSkillsByCategory(ctx, category, search)
	if err != nil {
		return nil, err
	}

	var entries []SkillMarketEntry
	for _, skill := range skills {
		avgRating, count, _ := s.repo.GetRatingStats(ctx, skill.ID)
		installCount, _ := s.repo.GetInstallCount(ctx, skill.ID)

		entries = append(entries, SkillMarketEntry{
			Skill:        skill,
			AvgRating:    avgRating,
			RatingCount:  count,
			InstallCount: installCount,
		})
	}

	return entries, nil
}

func (s *SkillEnterpriseService) ResolveDependencies(ctx context.Context, skillID string) ([]DependencyNode, error) {
	return s.repo.ResolveDepTree(ctx, skillID)
}

type SkillWithDeps struct {
	Skill        Skill
	Dependencies []SkillDependency
}

type SkillMarketEntry struct {
	Skill
	AvgRating    float64 `json:"avg_rating"`
	RatingCount  int     `json:"rating_count"`
	InstallCount int     `json:"install_count"`
}

type DependencyNode struct {
	SkillID  string           `json:"skill_id"`
	Name     string           `json:"name"`
	Version  string           `json:"version"`
	Children []DependencyNode `json:"children,omitempty"`
}

type SkillDependency struct {
	SkillID      string
	DependencyID string
	Version      string
}
