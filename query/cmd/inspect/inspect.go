// Package inspect
package inspect

import (
	"github.com/seantronsen/openchami-logq/query/cmd/inspect/config"
	"github.com/seantronsen/openchami-logq/query/cmd/inspect/dates"
	"github.com/seantronsen/openchami-logq/query/cmd/inspect/schema"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "inspect",
		Short: "Inspect dataset metadata",
		Long: `Inspect metadata about datasets stored in the log lake, including
available date partitions and schema information.`,
	}

	cmd.AddCommand(
		config.NewCmd(),
		dates.NewCmd(),
		schema.NewCmd(),
	)

	return cmd
}
