// Package sql
package sql

import (
	"context"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/sql"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sql",
		Short: "Execute ad-hoc SQL queries",
		Long:  "Execute ad-hoc SQL queries against both raw and compacted datasets in S3.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			querystr := args[0]
			options := opts.FromCobraCmd(cmd)
			return run(cfg, options, querystr)
		},
	}

	return cmd
}

func run(cfg *config.Config, options opts.Opts, querystr string) error {
	ctx := context.TODO()
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

	rows, err := engine.Query(querystr, sources)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		record, err := sql.ScanRowToMap(rows)
		if err != nil {
			return err
		}
		if err := enc.Encode(record); err != nil {
			return err
		}
	}

	return nil
}
