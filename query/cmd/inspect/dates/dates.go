// Package dates
package dates

import (
	"context"
	"fmt"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/sql"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dates",
		Short: "List available dates",
		Long:  "List available dates with captures available in the log lake.",
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			return run(cfg, options)
		},
	}

	return cmd
}

type Record struct {
	Date string `json:"date"`
}

// todo: again with the common logic. might be able to reduce with generics.
func run(cfg *config.Config, options opts.Opts) error {
	ctx := context.TODO()

	var expr string

	switch options.Source {
	case "events":
		expr = "cloudevent.ts"
	case "logs":
		expr = "ts"
	default:
		return fmt.Errorf("unknown source: %s", options.Source)
	}

	querystr := fmt.Sprintf(`
SELECT DISTINCT
    CAST(%s AS DATE) AS date
FROM SOURCES
ORDER BY date;
	`, expr)
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
		var record Record

		if err := rows.Scan(&record.Date); err != nil {
			return err
		}
		if err := enc.Encode(record); err != nil {
			return err
		}
	}

	return nil
}
