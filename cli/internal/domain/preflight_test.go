package domain

import "testing"

func TestResolvePreflightMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stage     string
		requested string
		want      string
	}{
		{name: "brainstorm defaults reduced", stage: "brainstorm", want: PreflightModeReduced},
		{name: "brainstorm lock durable", stage: "brainstorm", requested: "lock", want: PreflightModeDurable},
		{name: "to-plan durable", stage: "to-plan", requested: "full", want: PreflightModeDurable},
		{name: "work auto reduced", stage: "work", requested: "auto", want: PreflightModeReduced},
		{name: "work full durable", stage: "work", requested: "full", want: PreflightModeDurable},
		{name: "check review reduced", stage: "check", requested: "review", want: PreflightModeReduced},
		{name: "check gate durable", stage: "check", requested: "gate", want: PreflightModeDurable},
		{name: "handoff durable", stage: "handoff", want: PreflightModeDurable},
		{name: "watzup reduced", stage: "watzup", want: PreflightModeReduced},
		{name: "git reduced", stage: "git", want: PreflightModeReduced},
		{name: "interview reduced", stage: "interview", want: PreflightModeReduced},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolvePreflightMode(tt.stage, tt.requested)
			if err != nil {
				t.Fatalf("ResolvePreflightMode() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ResolvePreflightMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvePreflightModeRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name      string
		stage     string
		requested string
		wantCode  string
	}{
		{name: "unknown stage", stage: "deploy", wantCode: "invalid_stage"},
		{name: "invalid stage mode", stage: "watzup", requested: "full", wantCode: "invalid_mode"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ResolvePreflightMode(tt.stage, tt.requested)
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("error = %T %v, want *ValidationError", err, err)
			}
			if ve.Code != tt.wantCode {
				t.Fatalf("error code = %q, want %q", ve.Code, tt.wantCode)
			}
		})
	}
}
