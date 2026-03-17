// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package run
package run

import (
	"context"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/cmd/query"
	"github.com/seantronsen/openchami-logq/query/cmd/report/registry"
	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/report"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Execute a predefined report",
		Long: `Execute a predefined, parameterized report against the log lake.
Reports encapsulate common queries and accept arguments for filtering and
analysis without requiring raw SQL.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			choice := args[0]
			r, err := registry.Get(choice)
			if err != nil {
				return err
			}

			reportParams, err := report.ParseRawParams(args[1:])
			if err != nil {
				return err
			}

			querystr, err := r.BuildQueryString(reportParams)
			if err != nil {
				return err
			}

			return query.ExecUnstructured(
				querystr,
				context.TODO(),
				cfg,
				opts.FromCobraCmd(cmd),
			)
		},
	}

	return cmd
}
