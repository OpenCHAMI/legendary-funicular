// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package dates
package dates

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OpenCHAMI/legendary-funicular/query/cmd/opts"
	"github.com/OpenCHAMI/legendary-funicular/query/cmd/query"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dates",
		Short: "List available dates",
		Long:  "List available dates with captures available in the log lake.",
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			querystr, err := buildQueryString(options)
			if err != nil {
				return err
			}
			return query.ExecStructured(
				querystr,
				context.TODO(),
				cfg,
				opts.FromCobraCmd(cmd),
				scanner,
			)
		},
	}

	return cmd
}

type record struct {
	Date string `json:"date"`
}

func scanner(rows *sql.Rows) (record, error) {
	var record record
	err := rows.Scan(&record.Date)
	return record, err
}

func buildQueryString(options opts.Opts) (string, error) {
	var expr string

	switch options.Source {
	case "events":
		expr = "cloudevent.ts"
	case "logs":
		expr = "ts"
	default:
		return expr, fmt.Errorf("unknown source: %s", options.Source)
	}

	return fmt.Sprintf(`
SELECT DISTINCT
	CAST(%s AS DATE) AS date
FROM SOURCES
ORDER BY date;
			`, expr), nil
}
