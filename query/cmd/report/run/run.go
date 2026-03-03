// Package run
package run

import (
	"context"
	"fmt"
	"slices"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/reports"
	"github.com/seantronsen/openchami-logq/query/internal/sql"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Execute a predefined report",
		Long: `Execute a predefined, parameterized report against the log lake.
Reports encapsulate common queries and accept arguments for filtering and
analysis without requiring raw SQL.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			choice := args[0]
			options := opts.FromCobraCmd(cmd)
			return run(cfg, options, choice)
		},
	}

	return cmd
}

// todo: again with the common logic. might be able to reduce with generics.
func run(cfg *config.Config, options opts.Opts, choice string) error {
	ctx := context.TODO()
	idx := slices.IndexFunc(reports.Registry, func(r reports.Report) bool {
		return r.Name() == choice
	})
	if idx < 0 {
		return fmt.Errorf("unknown report: %s", choice)
	}

	sources := options.BuildSources(cfg)
	engine, err := sql.New(cfg, ctx)
	if err != nil {
		return err
	}
	defer engine.Close()

	enc, err := render.BuildEncoder(options.Output, options.Format)
	if err != nil {
		return err
	}
	defer enc.Close()

	reports.Registry[idx].QueryRecords(engine, sources, enc)

	return nil
}
