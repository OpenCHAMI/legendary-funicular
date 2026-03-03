// Package dump
package dump

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
		Use:   "dump",
		Short: "Dump dataset records to stdout",
		Long: `Dump dataset records from S3 to stdout, enabling both dataset
migration through standard shell redirection and downstream preprocessing via
pipes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			return run(cfg, options)
		},
	}

	return cmd
}

// todo: nearly identical to the sql subcommand. could use some abstraction...
func run(cfg *config.Config, options opts.Opts) error {
	ctx := context.TODO()
	sources := options.BuildSources(cfg)
	querystr := "SELECT * FROM SOURCES;"
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
