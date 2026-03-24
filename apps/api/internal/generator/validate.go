package generator

import "fmt"

func (g *Generator) validate(spec *SkillSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}
	if spec.Version == "" {
		spec.Version = "1.0"
	}
	if spec.RiskLevel == "" {
		spec.RiskLevel = "medium"
	}
	if len(spec.Steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}
	seenIDs := make(map[string]bool)
	for _, step := range spec.Steps {
		if step.ID == "" {
			return fmt.Errorf("step id is required")
		}
		if seenIDs[step.ID] {
			return fmt.Errorf("duplicate step id: %s", step.ID)
		}
		seenIDs[step.ID] = true
		if step.Type == "" {
			return fmt.Errorf("step type is required for step %s", step.ID)
		}
		if step.Type == "tool" && step.ToolRef == "" {
			return fmt.Errorf("tool_ref is required for tool step %s", step.ID)
		}
		if step.Type == "approval" && step.ApproverGroup == "" {
			return fmt.Errorf("approver_group is required for approval step %s", step.ID)
		}
	}
	return nil
}
