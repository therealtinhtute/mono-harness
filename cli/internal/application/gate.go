package application

import (
	"fmt"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// gateRequirement is one matrix cell's value, mirroring check/SKILL.md's
// references/gate-checklist.md Validation Matrix table exactly.
type gateRequirement string

const (
	gateRequired gateRequirement = "required"
	gateOptional gateRequirement = "optional"
	gateNA       gateRequirement = "n-a"
)

// proofClasses is the matrix's fixed column order, used to produce a
// stable (non-map-iteration-order) MissingRequired slice.
var proofClasses = []string{"unit", "integration", "e2e", "manual-check", "command-output"}

// validationMatrix is gate-checklist.md's table, ported to code so its
// determinism can be proven by a fixture test (validation-gate-PLAN.md
// T4) without exposing a new CLI command — the `check` skill still reads
// the matrix from gate-checklist.md itself per Step 1.6; this is a
// standalone determinism proof of the same design, not new CLI surface
// (see .kit/implementation-notes.md, validation-gate/T4).
var validationMatrix = map[string]map[string]gateRequirement{
	"tiny": {
		"unit": gateOptional, "integration": gateNA, "e2e": gateNA,
		"manual-check": gateOptional, "command-output": gateRequired,
	},
	"normal": {
		"unit": gateRequired, "integration": gateOptional, "e2e": gateNA,
		"manual-check": gateOptional, "command-output": gateRequired,
	},
	"high-risk": {
		"unit": gateRequired, "integration": gateRequired, "e2e": gateOptional,
		"manual-check": gateRequired, "command-output": gateRequired,
	},
}

// GateVerdict mirrors what check/SKILL.md Step 1.6 point 4 decides by
// hand: PASS when every required cell for the lane has matching gathered
// evidence, FAIL naming every missing required proof class otherwise.
type GateVerdict struct {
	Lane            string   `json:"lane"`
	Verdict         string   `json:"verdict"`
	MissingRequired []string `json:"missing_required"`
}

// EvaluateGate reuses the "invalid_lane" code (see domain/intake.go,
// domain/backlog.go) since an unresolvable lane is the same class of
// input error, even though this function has no CLI-exposed argument of
// its own to attribute it to.
func EvaluateGate(lane string, gathered map[string]bool) (GateVerdict, error) {
	row, ok := validationMatrix[lane]
	if !ok {
		return GateVerdict{}, &domain.ValidationError{
			Code:    "invalid_lane",
			Message: fmt.Sprintf("gate: invalid lane %q", lane),
		}
	}

	missing := []string{}
	for _, class := range proofClasses {
		if row[class] == gateRequired && !gathered[class] {
			missing = append(missing, class)
		}
	}

	verdict := "PASS"
	if len(missing) > 0 {
		verdict = "FAIL"
	}
	return GateVerdict{Lane: lane, Verdict: verdict, MissingRequired: missing}, nil
}
