// Package interfaces wires the zharness command surface (cobra) to the
// application layer. See cli/docs/CONTRACT.md for the full command
// contract. cli-core only registers the Core section (Phase 3); domain
// and research commands are added in later phases.
package interfaces

import "github.com/spf13/cobra"

// jsonOutput is bound to the --json persistent flag; every command reads
// it directly instead of re-resolving the flag from cmd.Flags().
var jsonOutput bool

// Execute builds and runs the root command, returning the process exit
// code per CONTRACT.md (0 success, 1 user error, 2 system error).
func Execute(version string) int {
	root := NewRootCmd(version)
	err := root.Execute()
	if err == nil {
		return 0
	}
	return handleError(root.ErrOrStderr(), err)
}

// NewRootCmd builds the zharness root command with all registered
// subcommands.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "zharness",
		Short:         "Durable state harness for the workflow skill chain",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")

	root.AddCommand(newIDCmd())
	root.AddCommand(newScaffoldCmd())
	root.AddCommand(newInitCmd(version))
	root.AddCommand(newMigrateCmd(version))
	root.AddCommand(newImportCmd())
	root.AddCommand(newDBCmd())
	root.AddCommand(newQueryCmd())
	root.AddCommand(newIntakeCmd())
	root.AddCommand(newStoryCmd())
	root.AddCommand(newInterventionCmd())
	root.AddCommand(newTraceCmd())
	root.AddCommand(newDecisionCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newResumeCmd(version))
	root.AddCommand(newPreflightCmd(version))
	root.AddCommand(newNextCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newHandoffCmd())
	root.AddCommand(newValidateCmd())
	root.AddCommand(newAuditCmd(version))
	wrapExclusiveMutationCommands(root)

	return root
}
