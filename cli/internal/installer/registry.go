package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// registryPath is the machine-wide list of repositories zharness manages,
// one absolute root per line. It is the only way `update --check --all` can
// find consumer repositories, which carry no hook of their own.
func registryPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "zharness", "repos"), nil
}

// Registered returns the recorded repository roots in file order.
func Registered() ([]string, error) {
	p, err := registryPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var roots []string
	for _, ln := range strings.Split(string(b), "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			roots = append(roots, ln)
		}
	}
	return dedupe(roots), nil
}

func writeRegistry(roots []string) error {
	p, err := registryPath()
	if err != nil {
		return err
	}
	body := ""
	if len(roots) > 0 {
		body = strings.Join(roots, "\n") + "\n"
	}
	return writeFileAtomic(p, []byte(body))
}

// register and unregister are advisory: a read-only or missing home must
// never fail the verb that called them, so errors become a warning line.
func register(root string, stdout *strings.Builder) {
	roots, err := Registered()
	if err == nil {
		for _, r := range roots {
			if r == root {
				return
			}
		}
		err = writeRegistry(append(roots, root))
	}
	if err != nil {
		fmt.Fprintf(stdout, "warning    repo registry not updated: %v\n", err)
	}
}

func unregister(root string, stdout *strings.Builder) {
	roots, err := Registered()
	if err == nil {
		kept := roots[:0]
		for _, r := range roots {
			if r != root {
				kept = append(kept, r)
			}
		}
		if len(kept) == len(roots) {
			return
		}
		err = writeRegistry(kept)
	}
	if err != nil {
		fmt.Fprintf(stdout, "warning    repo registry not updated: %v\n", err)
	}
}
