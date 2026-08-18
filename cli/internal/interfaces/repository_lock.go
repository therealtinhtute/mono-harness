package interfaces

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

const repositoryLockAnnotation = "zharness.repository-lock"

var repositoryLockAcquiredHook func(path string)

var exclusiveMutationCommandPaths = map[string]struct{}{
	"init":               {},
	"migrate":            {},
	"migrate layout":     {},
	"import":             {},
	"db rebuild":         {},
	"intake":             {},
	"story":              {},
	"trace add":          {},
	"decision add":       {},
	"memory add":         {},
	"run create":         {},
	"check record":       {},
	"handoff record":     {},
	"plan complete":      {},
	"plan abandon":       {},
}

func wrapExclusiveMutationCommands(root *cobra.Command) {
	found := make(map[string]bool, len(exclusiveMutationCommandPaths))
	var visit func(*cobra.Command, string)
	visit = func(parent *cobra.Command, prefix string) {
		for _, command := range parent.Commands() {
			path := command.Name()
			if prefix != "" {
				path = prefix + " " + path
			}
			if _, ok := exclusiveMutationCommandPaths[path]; ok {
				wrapExclusiveMutationCommand(command, path)
				found[path] = true
			}
			visit(command, path)
		}
	}
	visit(root, "")
	if len(found) != len(exclusiveMutationCommandPaths) {
		panic(fmt.Sprintf("exclusive mutation command inventory mismatch: found %d of %d", len(found), len(exclusiveMutationCommandPaths)))
	}
}

func wrapExclusiveMutationCommand(command *cobra.Command, path string) {
	original := command.RunE
	if original == nil {
		panic("exclusive mutation command has no RunE: " + path)
	}
	if command.Annotations == nil {
		command.Annotations = map[string]string{}
	}
	command.Annotations[repositoryLockAnnotation] = infrastructure.RepositoryLockExclusive.String()
	command.RunE = func(cmd *cobra.Command, args []string) (err error) {
		lock, lockErr := infrastructure.AcquireRepositoryLock(cmd.Context(), ".", infrastructure.RepositoryLockExclusive)
		if lockErr != nil {
			return mapRepositoryLockError(path, lockErr)
		}
		defer func() {
			if closeErr := lock.Close(); err == nil && closeErr != nil {
				err = mapRepositoryLockError(path, closeErr)
			}
		}()
		if repositoryLockAcquiredHook != nil {
			repositoryLockAcquiredHook(path)
		}
		return original(cmd, args)
	}
}
