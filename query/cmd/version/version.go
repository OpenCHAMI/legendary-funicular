// Package version
package version

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print build and version information for the openchami-logq CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			dev.NotImplemented()
		},
	}

	return cmd
}
