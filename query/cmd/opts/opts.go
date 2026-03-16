// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package opts
package opts

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/seantronsen/openchami-logq/query/internal/config"
	"github.com/seantronsen/openchami-logq/query/internal/utils"
	"github.com/spf13/cobra"
)

type Opts struct {
	Format string
	Output io.Writer
	Source string
}

func FromCobraCmd(cmd *cobra.Command) Opts {

	source, err := cmd.Root().PersistentFlags().GetString("source")
	utils.CheckFatal(err)

	format, err := cmd.Root().PersistentFlags().GetString("format")
	utils.CheckFatal(err)

	output, err := cmd.Root().PersistentFlags().GetString("output")
	utils.CheckFatal(err)

	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		utils.CheckFatal(err)
		w = f
	}

	return Opts{Format: format, Output: w, Source: source}
}

func (o *Opts) BuildSources(cfg *config.Config) []string {
	// todo: once the temp compaction for raw ndjson records is complete,
	// add the location(s) here as well.
	validSources := []string{"logs", "events"}

	var sources []string

	if !slices.Contains(validSources, o.Source) {
		utils.CheckFatal(fmt.Errorf("detected invalid source: '%s'", o.Source))
	}
	sources = append(sources, fmt.Sprintf("s3://%s/%s/**/*.parquet", *cfg.S3BucketParquet, o.Source))
	return sources
}
