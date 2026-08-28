// Package interfaces wires the zharness command surface (cobra).
//
// v0.15 slim: the entire lifecycle command surface has been deleted from
// source. State lives in git-committed markdown alone; the two fail-closed
// guarantees — proof re-execution and the independent-judge rule for
// high-risk lanes — live in the repository's pre-commit hook
// (scripts/install-git-hooks.sh), which reads staged bytes and trusts no
// marker an agent writes. See docs/plans/active/zharness-v015-slim.md.
//
// The three managed-set verbs (install / update / uninstall) land here in
// p3-installer; until then the surface is intentionally empty.
package interfaces

import (
	"os"

	"github.com/spf13/cobra"
)

// Execute builds and runs the root command, returning the process exit code.
func Execute(version string) int {
	root := NewRootCmd(version)
	err := root.Execute()
	exit := 0
	if err != nil {
		root.PrintErrln(err)
		exit = 1
	}
	_ = os.Stdout // keep os import stable for future verb output plumbing
	return exit
}

// NewRootCmd builds the zharness root command with its registered verbs.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "zharness",
		Short:         "Installer/updater for the markdown-first workflow harness",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Use: "help", Hidden: true})
	root.AddCommand(newInstallCmd(version))
	root.AddCommand(newUpdateCmd(version))
	root.AddCommand(newUninstallCmd())
	return root
}
