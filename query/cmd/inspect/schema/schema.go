// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package schema
package schema

import (
	"context"

	"github.com/OpenCHAMI/legendary-funicular/query/cmd/opts"
	"github.com/OpenCHAMI/legendary-funicular/query/cmd/query"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "schema",
		Short: "Show dataset schema",
		Long:  "Display the Parquet schema description for a dataset, including column names, types, and compression strategies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			querystr := `
		SELECT
			column_name,
			column_type,
			"null" AS is_nullable
		FROM (
			DESCRIBE SELECT * FROM SOURCES
		)`

			return query.ExecStructured(
				querystr,
				context.TODO(),
				cfg,
				opts.FromCobraCmd(cmd),
				query.ScanRecordSchemaInspection,
			)

		},
	}

	return cmd
}
