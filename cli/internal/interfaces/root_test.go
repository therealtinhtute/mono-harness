package interfaces

import "testing"

func TestRootCmd_ExactThreeVerbs(t *testing.T) {
	root := NewRootCmd("test")
	got := map[string]bool{}
	for _, c := range root.Commands() {
		got[c.Name()] = true
	}
	for _, want := range []string{"install", "update", "uninstall"} {
		if !got[want] {
			t.Errorf("missing verb %q", want)
		}
		delete(got, want)
	}
	for name := range got {
		t.Errorf("unexpected extra verb %q", name)
	}
}
