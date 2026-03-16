// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package describe
package describe

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
		Use:   "describe",
		Short: "Describe a report",
		Long: `Show detailed information about a report, including required
		parameters, defaults, and the underlying SQL query (warning: experimental).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			choice := args[0]
			options := opts.FromCobraCmd(cmd)
			return run(cfg, options, choice)
		},
	}

	return cmd
}

// todo: again with the re-use....  need to make a generalized "Schema"
// flavor struct abstraction for all cases where we're simply
// displaying an output schema.

type Record struct {
	ColumnName string `json:"column_name"`
	ColumnType string `json:"column_type"`
	IsNullable string `json:"is_nullable"`
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

	querystr := fmt.Sprintf(`
SELECT
	column_name,
	column_type,
	"null" AS is_nullable
FROM (
	DESCRIBE %s
);`, reports.Registry[idx].BuildQueryString())

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
		if err := rows.Scan(
			&record.ColumnName,
			&record.ColumnType,
			&record.IsNullable,
		); err != nil {
			return err
		}
		if err := enc.Encode(record); err != nil {
			return err
		}
	}

	return nil
}
