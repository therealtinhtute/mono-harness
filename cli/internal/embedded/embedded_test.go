package embedded

import (
	"strings"
	"testing"
)

func TestBuildManifest_PathsPresentAndNonEmpty(t *testing.T) {
	m, err := BuildManifest("dev")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if len(m.Paths) == 0 {
		t.Fatal("manifest has no paths")
	}
	for _, p := range m.Paths {
		data, err := FS.ReadFile(p)
		if err != nil {
			t.Errorf("ReadFile(%s): %v", p, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("%s is empty", p)
		}
	}
}

func TestPlaybookCount(t *testing.T) {
	count, err := PlaybookCount()
	if err != nil {
		t.Fatalf("PlaybookCount: %v", err)
	}
	if count != 6 {
		t.Fatalf("playbook count = %d, want 6", count)
	}
}

func TestBuildManifest_ShimAndAuthorityDocsPresent(t *testing.T) {
	m, err := BuildManifest("dev")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	want := []string{"AGENTS.md", "AUTHORITY.md", "CONTEXT_RULES.md"}
	for _, w := range want {
		found := false
		for _, p := range m.Paths {
			if p == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("manifest missing %s", w)
		}
	}
}

func TestBuildManifest_DocsVersionExposed(t *testing.T) {
	m, err := BuildManifest("0.2.0")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.DocsVersion != "0.2.0" {
		t.Fatalf("DocsVersion = %q, want %q", m.DocsVersion, "0.2.0")
	}
}

func TestBrainstormPlaybook_ExplicitExecutionIntent(t *testing.T) {
	data, err := FS.ReadFile("playbooks/brainstorm.md")
	if err != nil {
		t.Fatalf("ReadFile(playbooks/brainstorm.md): %v", err)
	}
	content := string(data)

	want := []string{
		"explicit execution intent",
		"mode or scope is genuinely ambiguous",
		"does not authorize replacing an existing SPEC",
		"without a second procedural response",
		"Never infer approval from silence or from a vague request",
	}
	for _, phrase := range want {
		if !strings.Contains(content, phrase) {
			t.Errorf("brainstorm.md missing autonomy contract phrase %q", phrase)
		}
	}

	obsolete := "confirm it with the user (ask a short structured question) before producing output"
	if strings.Contains(content, obsolete) {
		t.Errorf("brainstorm.md still contains obsolete unconditional gate %q", obsolete)
	}
}

func TestWorkPlaybook_UsesIDCommand(t *testing.T) {
	data, err := FS.ReadFile("playbooks/work.md")
	if err != nil {
		t.Fatalf("ReadFile(playbooks/work.md): %v", err)
	}
	content := string(data)

	want := []string{
		"zharness id --json",
		"save the returned `id` as the **RUN id**",
		"save this second, distinct ID as the **changeset id**",
		"Never invent a placeholder ULID",
		"never reuse the RUN id as the filename",
	}
	for _, phrase := range want {
		if !strings.Contains(content, phrase) {
			t.Errorf("work.md missing ID contract phrase %q", phrase)
		}
	}
}
