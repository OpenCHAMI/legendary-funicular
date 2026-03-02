// Package describe
package describe

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe a report",
		Long: `Show detailed information about a report, including required
parameters, defaults, and the underlying SQL query.`,
		Run: func(cmd *cobra.Command, args []string) {
			dev.NotImplemented()
		},
	}

	return cmd
}
