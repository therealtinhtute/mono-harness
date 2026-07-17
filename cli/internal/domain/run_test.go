package domain

import "testing"

func TestRunValidate(t *testing.T) {
	planID := "01J0000000000000000000PLAN"
	cases := []struct {
		name     string
		run      Run
		wantCode string
	}{
		{"valid", Run{StorySlug: "cli-domain", ArtifactPath: ".kit/runs/work/x.md", TraceIDs: []string{"t1"}}, ""},
		{"valid with plan id", Run{StorySlug: "cli-domain", ArtifactPath: ".kit/runs/work/x.md", PlanID: &planID}, ""},
		{"missing story_slug", Run{StorySlug: "", ArtifactPath: ".kit/runs/work/x.md"}, "missing_required_field"},
		{"missing artifact_path", Run{StorySlug: "cli-domain", ArtifactPath: ""}, "missing_required_field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
