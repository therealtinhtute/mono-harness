package domain

import "fmt"

// Intake type enum (CONTRACT.md `intake`).
const (
	IntakeNewSpec            = "new-spec"
	IntakeSpecSlice          = "spec-slice"
	IntakeChangeRequest      = "change-request"
	IntakeNewInitiative      = "new-initiative"
	IntakeMaintenance        = "maintenance"
	IntakeHarnessImprovement = "harness-improvement"
)

var intakeTypes = map[string]bool{
	IntakeNewSpec:            true,
	IntakeSpecSlice:          true,
	IntakeChangeRequest:      true,
	IntakeNewInitiative:      true,
	IntakeMaintenance:        true,
	IntakeHarnessImprovement: true,
}

// Lane enum, shared by `intake` and `backlog` (CONTRACT.md).
const (
	LaneTiny     = "tiny"
	LaneNormal   = "normal"
	LaneHighRisk = "high-risk"
)

var lanes = map[string]bool{
	LaneTiny:     true,
	LaneNormal:   true,
	LaneHighRisk: true,
}

type Intake struct {
	ID        string
	Type      string
	Summary   string
	Lane      string
	PlanPath  string
	PlanID    string
	CreatedAt string
}

func (i Intake) Validate() error {
	if !intakeTypes[i.Type] {
		return &ValidationError{Code: "invalid_type", Message: fmt.Sprintf("intake: invalid type %q", i.Type)}
	}
	if i.Summary == "" {
		return &ValidationError{Code: "missing_required_field", Message: "intake: summary is required"}
	}
	if !lanes[i.Lane] {
		return &ValidationError{Code: "invalid_lane", Message: fmt.Sprintf("intake: invalid lane %q", i.Lane)}
	}
	return nil
}
