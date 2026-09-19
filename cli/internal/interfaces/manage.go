package interfaces

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/therealtinhtute/skills/cli/internal/installer"
)

// resolveRoot resolves the repository root for the managed-set verbs: an
// explicit --root wins, else the git toplevel of the working directory,
// else the working directory itself.
func resolveRoot(cmd *cobra.Command, flagName string) string {
	r, _ := cmd.Flags().GetString(flagName)
	if r != "" {
		return abs(r)
	}
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return abs(".")
}

func abs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return a
}

func newInstallCmd(version string) *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Scaffold the managed doc set into this repository (idempotent; brownfield report is read-only)",
		RunE: func(c *cobra.Command, args []string) error {
			rootDir := resolveRoot(c, "root")
			out := &strings.Builder{}
			if err := installer.Install(rootDir, version, out); err != nil {
				fmt.Fprint(os.Stderr, out.String())
				return err
			}
			fmt.Print(out.String())
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", "", "target repository root (default: git toplevel of cwd)")
	return cmd
}

func newUpdateCmd(version string) *cobra.Command {
	var root string
	var cont, abort, check, all bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refresh playbooks/WORKFLOW.md from upstream; three-way merge PROJECT.md and the AGENTS.md block (conflicts stop for human resolution)",
		RunE: func(c *cobra.Command, args []string) error {
			if cont && abort {
				return fmt.Errorf("--continue and --abort are mutually exclusive")
			}
			if all && !check {
				return fmt.Errorf("--all requires --check")
			}
			rootDir := resolveRoot(c, "root")
			out := &strings.Builder{}
			if check {
				err := installer.RunCheck(rootDir, all, out)
				fmt.Print(out.String())
				return err
			}
			err := installer.RunUpdate(installer.UpdateOptions{
				Root: rootDir, Version: version, Continue: cont, Abort: abort,
			}, out)
			fmt.Print(out.String())
			return err
		},
	}
	cmd.Flags().StringVar(&root, "root", "", "target repository root (default: git toplevel of cwd)")
	cmd.Flags().BoolVar(&cont, "continue", false, "finalize after resolving conflict markers")
	cmd.Flags().BoolVar(&abort, "abort", false, "restore the pre-update state exactly")
	cmd.Flags().BoolVar(&check, "check", false, "report drift from this binary without writing; exit 1 on drift")
	cmd.Flags().BoolVar(&all, "all", false, "with --check: every repository in ~/.config/zharness/repos")
	return cmd
}

func newUninstallCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the managed doc set only — consumer-owned bytes are never touched",
		RunE: func(c *cobra.Command, args []string) error {
			rootDir := resolveRoot(c, "root")
			out := &strings.Builder{}
			if err := installer.Uninstall(rootDir, out); err != nil {
				fmt.Fprint(os.Stderr, out.String())
				return err
			}
			fmt.Print(out.String())
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", "", "target repository root (default: git toplevel of cwd)")
	return cmd
}
