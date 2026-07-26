package domain

import (
	"fmt"
	"strings"
)

const (
	PreflightModeReduced = "reduced"
	PreflightModeDurable = "durable"
)

var preflightModes = map[string]map[string]string{
	"brainstorm": {
		"":        PreflightModeReduced,
		"auto":    PreflightModeReduced,
		"explore": PreflightModeReduced,
		"lock":    PreflightModeDurable,
	},
	"to-plan": {
		"":      PreflightModeDurable,
		"full":  PreflightModeDurable,
		"phase": PreflightModeDurable,
	},
	"work": {
		"":        PreflightModeReduced,
		"auto":    PreflightModeReduced,
		"simple":  PreflightModeReduced,
		"bounded": PreflightModeReduced,
		"full":    PreflightModeDurable,
		"phase":   PreflightModeDurable,
	},
	"check": {
		"":        PreflightModeDurable,
		"auto":    PreflightModeDurable,
		"full":    PreflightModeDurable,
		"gate":    PreflightModeDurable,
		"review":  PreflightModeReduced,
		"simple":  PreflightModeReduced,
		"bounded": PreflightModeReduced,
	},
	"handoff": {
		"":        PreflightModeDurable,
		"durable": PreflightModeDurable,
	},
	"watzup": {
		"":        PreflightModeReduced,
		"reduced": PreflightModeReduced,
	},
	"git": {
		"":        PreflightModeReduced,
		"reduced": PreflightModeReduced,
	},
	"interview": {
		"":        PreflightModeReduced,
		"reduced": PreflightModeReduced,
	},
}

func ResolvePreflightMode(stage, requested string) (string, error) {
	stage = strings.ToLower(strings.TrimSpace(stage))
	requested = strings.ToLower(strings.TrimSpace(requested))
	modes, ok := preflightModes[stage]
	if !ok {
		return "", &ValidationError{
			Code:    "invalid_stage",
			Message: fmt.Sprintf("preflight: unknown stage %q", stage),
		}
	}
	mode, ok := modes[requested]
	if !ok {
		return "", &ValidationError{
			Code:    "invalid_mode",
			Message: fmt.Sprintf("preflight: mode %q is not valid for stage %q", requested, stage),
		}
	}
	return mode, nil
}
