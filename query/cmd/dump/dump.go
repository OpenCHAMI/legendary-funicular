// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package dump
package dump

import (
	"context"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/cmd/query"
	"github.com/seantronsen/openchami-logq/query/internal/config"
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
			ctx := context.TODO()
			querystr := "SELECT * FROM SOURCES;"
			return query.ExecUnstructured(querystr, ctx, cfg, options)
		},
	}

	return cmd
}
