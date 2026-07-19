package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

func newIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "id",
		Short: "Mint a new ULID without mutating harness state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := ulid.Make().String()
			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{"id": id})
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
}
