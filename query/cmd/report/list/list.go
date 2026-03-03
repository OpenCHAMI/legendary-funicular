// Package list
package list

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List available reports",
		Long:  "List all available predefined reports along with a brief description to aid discovery.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.NotImplemented()
		},
	}

	return cmd
}
