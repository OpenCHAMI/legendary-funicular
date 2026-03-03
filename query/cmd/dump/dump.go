// Package dump 
package dump

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump dataset records to stdout",
		Long: `Dump dataset records from S3 to stdout, enabling both dataset
migration through standard shell redirection and downstream preprocessing via
pipes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.NotImplemented()
		},
	}

	return cmd
}
