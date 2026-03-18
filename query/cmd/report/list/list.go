// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package list
package list

import (
	"slices"
	"strings"

	"github.com/OpenCHAMI/legendary-funicular/query/cmd/opts"
	"github.com/OpenCHAMI/legendary-funicular/query/cmd/report/registry"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/render"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/report"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List available reports",
		Long:  "List all available predefined reports along with a brief description to aid discovery.",
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			return run(options)
		},
	}

	return cmd
}

type record struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func run(options opts.Opts) error {

	items := registry.New()
	slices.SortFunc(items, func(a, b report.Report) int {
		return strings.Compare(a.Name(), b.Name())
	})

	enc, err := render.BuildEncoder(options.Output, options.Format)
	if err != nil {
		return err
	}
	defer enc.Close()

	for _, r := range items {
		rec := record{Name: r.Name(), Description: r.Description()}
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}
