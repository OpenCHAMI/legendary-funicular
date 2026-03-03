// Package sql
package sql

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sql",
		Short: "Execute ad-hoc SQL queries",
		Long:  "Execute ad-hoc SQL queries against both raw and compacted datasets in S3.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.NotImplemented()
		},
	}

	return cmd
}
