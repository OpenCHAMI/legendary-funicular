// Package run
package run

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Execute a predefined report",
		Long: `Execute a predefined, parameterized report against the log lake.
Reports encapsulate common queries and accept arguments for filtering and
analysis without requiring raw SQL.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.NotImplemented()
		},
	}

	return cmd
}
