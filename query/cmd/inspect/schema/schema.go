// Package schema 
package schema

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "schema",
		Short: "Show dataset schema",
		Long:  "Display the Parquet schema description for a dataset, including column names, types, and compression strategies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.NotImplemented()
		},
	}

	return cmd
}
