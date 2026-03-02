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
		Run: func(cmd *cobra.Command, args []string) {
			dev.NotImplemented()
		},
	}

	return cmd
}
