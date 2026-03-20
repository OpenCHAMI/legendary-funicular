// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package opts
package opts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/utils"
	"github.com/spf13/cobra"
)

type Opts struct {
	Format string
	Output io.Writer
	Source string
	Scope  string
}

func FromCobraCmd(cmd *cobra.Command) Opts {

	scope, err := cmd.Root().PersistentFlags().GetString("scope")
	utils.CheckFatal(err)

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

	return Opts{Format: format, Output: w, Source: source, Scope: scope}
}

func (o *Opts) validate() error {
	validSources := []string{"logs", "events"}
	validScopes := []string{"all", "compacted", "recent"}
	if !slices.Contains(validSources, o.Source) {
		return fmt.Errorf("invalid source: '%s'", o.Source)
	}
	if !slices.Contains(validScopes, o.Scope) {
		return fmt.Errorf("invalid scope: '%s'", o.Scope)
	}
	return nil
}

func (o *Opts) BuildSources(cfg *config.Config) []string {
	if err := o.validate(); err != nil {
		// todo: redo this properly
		utils.CheckFatal(err)
	}

	var sources []string
	if o.Scope == "all" || o.Scope == "compacted" {
		s := fmt.Sprintf(
			"read_parquet( 's3://%s/%s/**/*.parquet', union_by_name = true )",
			*cfg.S3BucketParquet,
			o.Source,
		)
		sources = append(sources, s)
	}

	if o.Scope == "all" || o.Scope == "recent" {
		s := fmt.Sprintf(
			"read_json( 's3://%s/%s/**/*.ndjson.zst', union_by_name = true )",
			*cfg.S3BucketNDJSON,
			o.Source,
		)
		sources = append(sources, s)
	}

	if len(sources) == 0 {
		// todo: redo this properly
		utils.CheckFatal(errors.New("change guard rail failure, unknown source/stream"))
	}

	return sources
}
