package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain keeps every test in this package away from the real
// ~/.config/zharness/repos: Install, update and uninstall all touch it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "zharness-xdg-")
	if err != nil {
		panic(err)
	}
	os.Setenv("XDG_CONFIG_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func isolatedRegistry(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return filepath.Join(dir, "zharness", "repos")
}

func TestRegistry_InstallTwiceRecordsOnce_UninstallRemoves(t *testing.T) {
	reg := isolatedRegistry(t)
	root := tempRepo(t, true)
	mustInstall(t, root)
	mustInstall(t, root)

	got, err := Registered()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != root {
		t.Fatalf("registry = %v, want [%s]", got, root)
	}

	var sb strings.Builder
	if err := Uninstall(root, &sb); err != nil {
		t.Fatalf("uninstall: %v\n%s", err, sb.String())
	}
	b, err := os.ReadFile(reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 0 {
		t.Fatalf("registry after uninstall = %q, want empty", b)
	}
}

func TestRegistry_UpdateRegisters(t *testing.T) {
	isolatedRegistry(t)
	root := tempRepo(t, true)
	mustInstall(t, root)
	if err := writeRegistry(nil); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "test"}, &sb); err != nil {
		t.Fatalf("update: %v\n%s", err, sb.String())
	}
	got, _ := Registered()
	if len(got) != 1 || got[0] != root {
		t.Fatalf("registry = %v, want [%s]", got, root)
	}
}

func TestRegistry_UnwritableHomeWarnsAndInstallSucceeds(t *testing.T) {
	dir := t.TempDir()
	// A regular file where the config dir should be makes MkdirAll fail on
	// every platform, including when the test runs as root.
	blocker := filepath.Join(dir, "cfg")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", blocker)

	out := mustInstall(t, tempRepo(t, true))
	if !strings.Contains(out, "warning    repo registry not updated") {
		t.Fatalf("expected registry warning, got:\n%s", out)
	}
}
