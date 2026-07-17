package domain

import "testing"

func TestStoryValidate(t *testing.T) {
	dep := "other-phase"
	cases := []struct {
		name     string
		story    Story
		wantCode string
	}{
		{"valid planned, no dependency", Story{Slug: "cli-domain", Goal: "port commands", Status: StoryPlanned}, ""},
		{"valid with dependency", Story{Slug: "cli-domain", Goal: "port commands", Status: StoryPlanned, DependsOn: &dep}, ""},
		{"valid in-progress", Story{Slug: "cli-domain", Goal: "port commands", Status: StoryInProgress}, ""},
		{"valid checked", Story{Slug: "cli-domain", Goal: "port commands", Status: StoryChecked}, ""},
		{"valid done", Story{Slug: "cli-domain", Goal: "port commands", Status: StoryDone}, ""},
		{"missing slug", Story{Slug: "", Goal: "port commands", Status: StoryPlanned}, "missing_required_field"},
		{"missing goal", Story{Slug: "cli-domain", Goal: "", Status: StoryPlanned}, "missing_required_field"},
		{"invalid status", Story{Slug: "cli-domain", Goal: "port commands", Status: "bogus"}, ""}, // uncoded internal invariant, see below
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.story.Validate()
			if tc.name == "invalid status" {
				if err == nil {
					t.Fatal("Validate() = nil, want an error for invalid status")
				}
				if ve, ok := err.(*ValidationError); ok && ve.Code != "" {
					t.Fatalf("Validate() code = %q, want empty (uncoded internal invariant)", ve.Code)
				}
				return
			}
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
