// Package dates
package dates

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dates",
		Short: "List available dates",
		Long:  "List available dates with captures available in the log lake.",
		Run: func(cmd *cobra.Command, args []string) {
			dev.NotImplemented()
		},
	}

	return cmd
}
