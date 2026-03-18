// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package sql
package sql

import (
	"context"

	"github.com/OpenCHAMI/legendary-funicular/query/cmd/query"
	"github.com/OpenCHAMI/legendary-funicular/query/cmd/opts"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sql",
		Short: "Execute ad-hoc SQL queries",
		Long:  "Execute ad-hoc SQL queries against both raw and compacted datasets in S3.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			ctx := context.TODO()
			querystr := args[0]
			return query.ExecUnstructured(querystr, ctx, cfg, options)
		},
	}

	return cmd
}
