package embedded

import "io/fs"

// PlaybookNames returns the six canonical playbook file names embedded under
// playbooks/, sorted.
func PlaybookNames() ([]string, error) {
	entries, err := fs.ReadDir(FS, "playbooks")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// ProjectIdentityTemplate returns the embedded docs/PROJECT.md skeleton
// (unanswered-question form, <=50 lines) shipped with the managed set.
func ProjectIdentityTemplate() ([]byte, error) { return Templates.ReadFile("project.identity.md") }
