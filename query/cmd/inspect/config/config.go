// Package config
package config

import (
	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		Long: `Display the effective configuration as resolved from CLI flags and
environment variables. Shows all relevant settings, marks unset values as UNSET,
and redacts secrets as ********.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			return render.Render(options.Output, options.Format, cfg)
		},
	}
	return cmd
}
