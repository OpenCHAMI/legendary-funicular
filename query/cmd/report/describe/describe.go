// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package describe
package describe

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/cmd/query"
	"github.com/seantronsen/openchami-logq/query/cmd/report/registry"
	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/report"
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
			// todo: fix
			// though, it does bring about questions on how to structure it
			// when we enable multiple output formats... regardless, it will
			// likely require a rewrite of the rendering library.
			slog.Warn("the describe subcommand is missing logic for detailing any arguments associated with a given report")

			choice := args[0]
			r, err := registry.Get(choice)
			if err != nil {
				return err
			}

			// semi-redundant b/c of the Args definition, leaving as change guard
			if len(args[1:]) != 0 {
				return fmt.Errorf("received unexpected additional arguments: %s", args[1:])
			}

			querystr, err := report.BuildSchemaQueryStr(r)
			if err != nil {
				return err
			}

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
