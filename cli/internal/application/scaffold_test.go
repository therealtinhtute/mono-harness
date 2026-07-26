package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/embedded"
)

func TestScaffoldArtifact_EachKind(t *testing.T) {
	for _, kind := range []string{"run", "check", "handoff", "spec", "plan"} {
		dst := filepath.Join(t.TempDir(), kind+".md")
		data, err := ScaffoldArtifact(embedded.Templates, kind, dst)
		if err != nil {
			t.Fatalf("kind %s: %v", kind, err)
		}
		if len(data) == 0 {
			t.Fatalf("kind %s: empty template", kind)
		}
		onDisk, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("kind %s: file not written: %v", kind, err)
		}
		if !strings.Contains(string(onDisk), "type: "+kind) {
			t.Fatalf("kind %s: skeleton missing `type: %s` frontmatter", kind, kind)
		}
	}
}

func TestScaffoldArtifact_UnknownKind(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "x.md")
	_, err := ScaffoldArtifact(embedded.Templates, "bogus", dst)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_kind" {
		t.Fatalf("want unknown_kind ValidationError, got %v", err)
	}
	if _, statErr := os.Stat(dst); statErr == nil {
		t.Fatalf("unknown kind should not write a file")
	}
}

func TestScaffoldArtifact_EmptyPath(t *testing.T) {
	_, err := ScaffoldArtifact(embedded.Templates, "run", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("want missing_required_field ValidationError, got %v", err)
	}
}

func TestScaffoldArtifact_RefuseOverwriteNonEmpty(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "run.md")
	if err := os.WriteFile(dst, []byte("half-filled work in progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ScaffoldArtifact(embedded.Templates, "run", dst)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "file_exists" {
		t.Fatalf("want file_exists ValidationError, got %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "half-filled work in progress\n" {
		t.Fatalf("existing content was clobbered: %q", string(got))
	}
}

func TestScaffoldArtifact_OverwriteEmptyPlaceholder(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "run.md")
	if err := os.WriteFile(dst, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScaffoldArtifact(embedded.Templates, "run", dst); err != nil {
		t.Fatalf("empty placeholder should be fillable: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if !strings.Contains(string(got), "type: run") {
		t.Fatalf("skeleton not written over empty placeholder")
	}
}
