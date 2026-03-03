// Package config
package config

import (
	"github.com/seantronsen/openchami-logq/query/internal/dev"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		Long: `Display the effective configuration as resolved from CLI flags and
environment variables. Shows all relevant settings, marks unset values as UNSET,
and redacts secrets as ********.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.NotImplemented()
		},
	}

	return cmd
}
