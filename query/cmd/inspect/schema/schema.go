// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package schema
package schema

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
		Use:   "schema",
		Short: "Show dataset schema",
		Long:  "Display the Parquet schema description for a dataset, including column names, types, and compression strategies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			return run(cfg, options)
		},
	}

	return cmd
}

type Record struct {
	ColumnName string `json:"column_name"`
	ColumnType string `json:"column_type"`
	IsNullable string `json:"is_nullable"`
}

// todo: again with the common logic. might be able to reduce with generics.
func run(cfg *config.Config, options opts.Opts) error {
	ctx := context.TODO()

	querystr := `
SELECT
	column_name,
	column_type,
	"null" AS is_nullable
FROM (
	DESCRIBE SELECT * FROM SOURCES
);`
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
