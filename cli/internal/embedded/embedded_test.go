package embedded

import "testing"

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
